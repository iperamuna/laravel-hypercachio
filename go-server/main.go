package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/yvasiyarov/php_session_decoder/php_serialize"
	_ "modernc.org/sqlite"
)

var (
	port         int
	host         string
	apiToken     string
	sslEnabled   bool
	sslCert      string
	sslKey       string
	artisanPath  string
	sqlitePath   string
	cachePrefix  string
	directSqlite bool

	db *sql.DB
)

type Payload struct {
	Value interface{} `json:"value"`
	TTL   *int        `json:"ttl"`
	Owner string      `json:"owner"`
}

func main() {
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&host, "host", "127.0.0.1", "Host to listen on")
	flag.StringVar(&apiToken, "token", "", "API Token for authentication")
	flag.BoolVar(&sslEnabled, "ssl", false, "Enable SSL")
	flag.StringVar(&sslCert, "cert", "", "SSL Certificate path")
	flag.StringVar(&sslKey, "key", "", "SSL Key path")
	flag.StringVar(&artisanPath, "artisan", "php artisan", "Path to artisan command")
	flag.StringVar(&sqlitePath, "sqlite-path", "", "Path to SQLite database")
	flag.StringVar(&cachePrefix, "prefix", "", "Cache key prefix")
	flag.BoolVar(&directSqlite, "direct-sqlite", true, "Execute caching internally via SQLite")
	flag.Parse()

	if apiToken == "" {
		log.Fatal("API Token is required")
	}

	if sqlitePath != "" && directSqlite {
		var err error
		db, err = sql.Open("sqlite", sqlitePath)
		if err != nil {
			log.Fatalf("Failed to open SQLite database: %s", err)
		}
		// Configure connections
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		log.Printf("Direct SQLite connection enabled: %s", sqlitePath)

		if err := initSqlite(); err != nil {
			log.Fatalf("Failed to initialize SQLite schema: %s", err)
		}

		if cachePrefix != "" {
			log.Printf("Using cache prefix: %s", cachePrefix)
		}
	} else {
		log.Printf("Direct SQLite disabled. Using artisan relay: %s", artisanPath)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/hypercacheio/cache/", handleCache)
	mux.HandleFunc("/api/hypercacheio/add/", handleAdd)
	mux.HandleFunc("/api/hypercacheio/lock/", handleLock)
	mux.HandleFunc("/api/hypercacheio/ping", handlePing)

	serverAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Starting Hypercacheio Go server on %s", serverAddr)

	var err error
	if sslEnabled {
		if sslCert == "" || sslKey == "" {
			log.Fatal("SSL Certificate and Key are required when SSL is enabled")
		}
		err = http.ListenAndServeTLS(serverAddr, sslCert, sslKey, authMiddleware(mux))
	} else {
		err = http.ListenAndServe(serverAddr, authMiddleware(mux))
	}

	if err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Hypercacheio-Token")
		if token != apiToken {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleCache(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/hypercacheio/cache/")
	if key == "" && r.Method != "DELETE" {
		http.Error(w, `{"error": "Key is required"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		if db != nil {
			handleSqliteGet(w, cachePrefix+key)
		} else {
			executeArtisan(w, "get", key, "")
		}
	case "POST":
		body, _ := io.ReadAll(r.Body)
		if db != nil {
			handleSqlitePut(w, cachePrefix+key, body)
		} else {
			executeArtisan(w, "put", key, string(body))
		}
	case "DELETE":
		if key == "" {
			if db != nil {
				db.Exec("DELETE FROM cache")
				writeJSON(w, map[string]bool{"success": true})
			} else {
				executeArtisan(w, "flush", "", "")
			}
		} else {
			if db != nil {
				db.Exec("DELETE FROM cache WHERE key = ?", cachePrefix+key)
				writeJSON(w, map[string]bool{"success": true})
			} else {
				executeArtisan(w, "forget", key, "")
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/api/hypercacheio/add/")
	body, _ := io.ReadAll(r.Body)
	if db != nil {
		handleSqliteAdd(w, cachePrefix+key, body)
	} else {
		executeArtisan(w, "add", key, string(body))
	}
}

func handleLock(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/hypercacheio/lock/")
	switch r.Method {
	case "POST":
		body, _ := io.ReadAll(r.Body)
		if db != nil {
			handleSqliteLockAcquire(w, cachePrefix+key, body)
		} else {
			executeArtisan(w, "lock", key, string(body))
		}
	case "DELETE":
		body, _ := io.ReadAll(r.Body)
		if db != nil {
			handleSqliteLockRelease(w, cachePrefix+key, body)
		} else {
			executeArtisan(w, "releaseLock", key, string(body))
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	if db != nil {
		hostName, _ := os.Hostname()
		writeJSON(w, map[string]interface{}{
			"message":  "pong",
			"role":     "go-server",
			"hostname": hostName,
			"time":     time.Now().Unix(),
		})
	} else {
		executeArtisan(w, "ping", "", "")
	}
}

// -------------------------------------------------------------
// Direct SQLite Implementations
// -------------------------------------------------------------

func initSqlite() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache(
			key TEXT PRIMARY KEY,
			value BLOB NOT NULL,
			expiration INTEGER
		);
		CREATE TABLE IF NOT EXISTS cache_locks(
			key TEXT PRIMARY KEY,
			owner TEXT NOT NULL,
			expiration INTEGER
		);
	`)
	return err
}

func handleSqliteGet(w http.ResponseWriter, key string) {
	var valBlob []byte
	var expiration sql.NullInt64

	err := db.QueryRow("SELECT value, expiration FROM cache WHERE key = ?", key).Scan(&valBlob, &expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, map[string]interface{}{"data": nil})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if expiration.Valid && expiration.Int64 < time.Now().Unix() {
		writeJSON(w, map[string]interface{}{"data": nil})
		return
	}

	decoder := php_serialize.NewUnSerializer(string(valBlob))
	parsed, err := decoder.Decode()
	if err != nil {
		log.Printf("Failed to deserialize value for key %s: %v", key, err)
		writeJSON(w, map[string]interface{}{"data": nil})
		return
	}
	writeJSON(w, map[string]interface{}{"data": parsed})
}

func handleSqlitePut(w http.ResponseWriter, key string, body []byte) {
	var payload Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var expiration *int64
	if payload.TTL != nil && *payload.TTL > 0 {
		exp := time.Now().Unix() + int64(*payload.TTL)
		expiration = &exp
	}

	encoded, err := php_serialize.Serialize(payload.Value)
	if err != nil {
		log.Printf("Failed to encode payload value: %v", err)
		http.Error(w, "Failed to serialize value", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("REPLACE INTO cache(key, value, expiration) VALUES(?, ?, ?)", key, encoded, expiration)
	if err != nil {
		log.Printf("Failed to put cache: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]bool{"success": true})
}

func handleSqliteAdd(w http.ResponseWriter, key string, body []byte) {
	var payload Payload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var expiration *int64
	if payload.TTL != nil && *payload.TTL > 0 {
		exp := time.Now().Unix() + int64(*payload.TTL)
		expiration = &exp
	}

	encoded, err := php_serialize.Serialize(payload.Value)
	if err != nil {
		http.Error(w, "Failed to serialize value", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO cache(key, value, expiration) VALUES(?, ?, ?)", key, encoded, expiration)
	if err != nil {
		// Possibly already exists. Check if expired.
		var existingExp sql.NullInt64
		errGet := db.QueryRow("SELECT expiration FROM cache WHERE key = ?", key).Scan(&existingExp)
		if errGet == nil && existingExp.Valid && existingExp.Int64 < time.Now().Unix() {
			_, errUpdate := db.Exec("UPDATE cache SET value = ?, expiration = ? WHERE key = ?", encoded, expiration, key)
			if errUpdate == nil {
				writeJSON(w, map[string]bool{"added": true})
				return
			}
		}

		writeJSON(w, map[string]bool{"added": false})
		return
	}

	writeJSON(w, map[string]bool{"added": true})
}

func handleSqliteLockAcquire(w http.ResponseWriter, key string, body []byte) {
	var payload Payload
	json.Unmarshal(body, &payload)

	ttl := 0
	if payload.TTL != nil {
		ttl = *payload.TTL
	}
	expiration := time.Now().Unix() + int64(ttl)

	_, err := db.Exec("INSERT INTO cache_locks(key, owner, expiration) VALUES(?, ?, ?)", key, payload.Owner, expiration)
	if err != nil {
		// Check if expired
		var existingExp sql.NullInt64
		errGet := db.QueryRow("SELECT expiration FROM cache_locks WHERE key = ?", key).Scan(&existingExp)
		if errGet == nil && existingExp.Valid && existingExp.Int64 < time.Now().Unix() {
			_, errUpdate := db.Exec("UPDATE cache_locks SET owner = ?, expiration = ? WHERE key = ?", payload.Owner, expiration, key)
			if errUpdate == nil {
				writeJSON(w, map[string]bool{"acquired": true})
				return
			}
		}

		writeJSON(w, map[string]bool{"acquired": false})
		return
	}

	writeJSON(w, map[string]bool{"acquired": true})
}

func handleSqliteLockRelease(w http.ResponseWriter, key string, body []byte) {
	var payload Payload
	json.Unmarshal(body, &payload)

	res, err := db.Exec("DELETE FROM cache_locks WHERE key = ? AND owner = ?", key, payload.Owner)
	if err != nil {
		writeJSON(w, map[string]bool{"released": false})
		return
	}
	rows, _ := res.RowsAffected()
	writeJSON(w, map[string]bool{"released": rows > 0})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// -------------------------------------------------------------
// Legacy Artisan Relay
// -------------------------------------------------------------

func executeArtisan(w http.ResponseWriter, action, key, payload string) {
	args := []string{"artisan", "hypercacheio:server-handler", action}
	if key != "" {
		args = append(args, "--key="+key)
	}
	if payload != "" {
		args = append(args, "--payload="+payload)
	}

	cmdParts := strings.Fields(artisanPath)
	fullArgs := append(cmdParts[1:], "hypercacheio:server-handler", action)
	if key != "" {
		fullArgs = append(fullArgs, "--key="+key)
	}
	if payload != "" {
		fullArgs = append(fullArgs, "--payload="+payload)
	}

	cmd := exec.Command(cmdParts[0], fullArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Artisan error: %s, output: %s", err, string(output))
		http.Error(w, fmt.Sprintf(`{"error": "Artisan command failed", "details": %q}`, string(output)), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
