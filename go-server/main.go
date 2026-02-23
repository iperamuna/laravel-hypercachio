package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var (
	port        int
	host        string
	apiToken    string
	sslEnabled  bool
	sslCert     string
	sslKey      string
	artisanPath string
)

func main() {
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&host, "host", "127.0.0.1", "Host to listen on")
	flag.StringVar(&apiToken, "token", "", "API Token for authentication")
	flag.BoolVar(&sslEnabled, "ssl", false, "Enable SSL")
	flag.StringVar(&sslCert, "cert", "", "SSL Certificate path")
	flag.StringVar(&sslKey, "key", "", "SSL Key path")
	flag.StringVar(&artisanPath, "artisan", "php artisan", "Path to artisan command")
	flag.Parse()

	if apiToken == "" {
		log.Fatal("API Token is required")
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
		executeArtisan(w, "get", key, "")
	case "POST":
		body, _ := io.ReadAll(r.Body)
		executeArtisan(w, "put", key, string(body))
	case "DELETE":
		if key == "" {
			executeArtisan(w, "flush", "", "")
		} else {
			executeArtisan(w, "forget", key, "")
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
	executeArtisan(w, "add", key, string(body))
}

func handleLock(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/hypercacheio/lock/")
	switch r.Method {
	case "POST":
		body, _ := io.ReadAll(r.Body)
		executeArtisan(w, "lock", key, string(body))
	case "DELETE":
		body, _ := io.ReadAll(r.Body)
		executeArtisan(w, "releaseLock", key, string(body))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	executeArtisan(w, "ping", "", "")
}

func executeArtisan(w http.ResponseWriter, action, key, payload string) {
	args := []string{"artisan", "hypercacheio:server-handler", action}
	if key != "" {
		args = append(args, "--key="+key)
	}
	if payload != "" {
		args = append(args, "--payload="+payload)
	}

	// Split artisan command in case it contains space (e.g. "php artisan")
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
