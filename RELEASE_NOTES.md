# Release Notes - v1.5.1

**Laravel Hyper-Cache-IO**

This update moves the binary storage location and adds a configuration toggle for direct SQLite execution.

## 📁 Binary Path Change
*   **Location**: We have moved the default Go build output from `storage/hypercacheio/bin` to `resources/hypercacheio/bin`. 
*   **Commitable Binaries**: This allows you to compile binaries locally and commit them to your version control system, enabling production deployments without requiring a Go installation on the server.

## ⚙️ Configuration
*   **Direct SQLite Toggle**: Added `HYPERCACHEIO_GO_DIRECT_SQLITE` to the `go_server` config. Set to `true` (default) for maximum performance, or `false` to use the legacy Artisan relay.

---

# Release Notes - v1.5.0

**Laravel Hyper-Cache-IO**

This major update introduces direct SQLite integration for the Go server, providing significant performance improvements.

## 🚀 Native SQLite Backend
The Go server now interacts directly with the SQLite database.
*   **Performance**: By bypassing the `php artisan` bootstrap overhead, cache operation response times have dropped from ~50ms to **less than 1ms**.
*   **PHP Serialization**: Implemented native PHP serialization support in Go to ensure full compatibility with data stored by the Laravel application.
*   **Atomic Operations**: Native support for atomic `add` and distributed locking directly in the Go daemon.

---

# Release Notes - v1.4.1

**Laravel Hyper-Cache-IO**

This update focuses on simplifying the storage architecture and optimizing the package footprint.

## 📁 Storage Consolidation

We have moved the default storage location from `storage/cache/hypercacheio/` to a more direct `storage/hypercacheio/`. This affects:
*   **Database**: The SQLite file is now at `storage/hypercacheio/hypercacheio.sqlite`.
*   **Go Binaries**: Compiled Go servers are now stored in `storage/hypercacheio/bin/`.

These paths remain fully configurable via `config/hypercacheio.php`.

## 📦 Package Optimization

*   **Binaries Removed**: Pre-compiled Go binaries have been removed from the core package. This significantly reduces the installation size and ensures that the server is always compiled specifically for your environment's OS and architecture.
*   **CI Improvements**: CI pipelines have been adjusted to handle the new directory structures.

## 📦 Upgrade

If you are upgrading from `v1.4.0`, please move your existing data:
```bash
mv storage/cache/hypercacheio storage/hypercacheio
```

Then update your package:
```bash
composer update iperamuna/laravel-hypercacheio
```

---

# Release Notes - v1.4.0

**Laravel Hyper-Cache-IO**

This major update introduces a high-performance Go-based synchronization server, robust service management, and enhanced connectivity diagnostics with firewall troubleshooting.

## 🚀 Standalone Go Server (Experimental)

We have introduced a standalone Go binary to handle inter-node cache synchronization.
*   **Performance**: Drastically reduces the overhead of handling cache requests by bypassing the full Laravel framework stack for internal peer-to-peer communication.
*   **Concurrency**: Built on Go's lightweight goroutines, allowing thousands of concurrent cache operations with minimal memory usage.
*   **Daemon Mode**: Runs as a persistent background process.

## 🛡️ Service Management

You can now easily run the Go server as a system service with auto-restart capabilities:
*   **Command**: `php artisan hypercacheio:go-server make-service`
*   **Linux**: Generates a `systemd` unit file for high-availability deployments.
*   **macOS**: Generates a `launchd` plist for local development or Apple Silicon servers.

## 🔍 Intelligent Connectivity Diagnostics

The `hypercacheio:connectivity-check` command is now significantly more powerful:
*   **Pre-flight checks**: Pings the local node first to ensure the cache server (Laravel or Go) is actually listening.
*   **Firewall Advice**: If a port is blocked, the command detects your OS and provides the exact `ufw`, `firewalld`, or `socketfilterfw` commands to unblock it.

## 🛠 Improvements

*   **Auto-Installation**: The Go server CLI can now detect if Go is missing and offer to install it via system package managers.
*   **Improved CI**: GitHub Actions now verify Go server compilation across AMD64 and ARM64 architectures.
*   **Test Suite**: Expanded coverage with new tests for the Go bridge and service management commands.

## 📦 Upgrade

```bash
composer update iperamuna/laravel-hypercacheio
```

To switch to the Go server, update your `.env`:
```dotenv
HYPERCACHEIO_SERVER_TYPE=go
```

Then compile and start the server:
```bash
php artisan hypercacheio:go-server compile
php artisan hypercacheio:go-server start
```

---

# Release Notes - v1.3.2

**Laravel Hyper-Cache-IO**

This release focuses on code quality, documentation, and test coverage improvements for the `HypercacheioSecondaryUrls` helper and the overall package configuration.

## 🛠 Improvements

*   **`HypercacheioSecondaryUrls` Helper Rewrite**: The helper function has been completely rewritten with:
    *   Proper PHPDoc blocks with full `@param`, `@return`, and `@example` tags.
    *   URL trimming — whitespace around URLs is automatically stripped.
    *   Empty-entry filtering — consecutive delimiters, leading/trailing delimiters no longer produce empty entries.
    *   URL validation — only entries passing `FILTER_VALIDATE_URL` are included; invalid URLs are silently discarded.
    *   Fixed `$delimeter` → `$delimiter` typo and added explicit `string` type hint.
    *   Cleaner functional implementation using `array_map`/`array_filter`.

*   **Configuration Comments**: All config block comments in `config/hypercacheio.php` have been enhanced with detailed descriptions, `Env:` variable references, expected formats, default values, and practical guidance.

*   **README Enhancements**:
    *   Added `HYPERCACHEIO_SECONDARY_URLS` and `HYPERCACHEIO_ASYNC` environment variable documentation to the Primary Server setup section.
    *   Added a full **Environment Variables Reference** table.
    *   Refreshed the Advanced Configuration code block to match the actual config.

## ✅ New Tests

*   **13 new Pest tests** for `HypercacheioSecondaryUrls` covering:
    *   Null, empty string, and whitespace-only input
    *   Single and multiple URL parsing (with and without whitespace)
    *   Empty-entry filtering from consecutive delimiters
    *   Invalid URL rejection via `FILTER_VALIDATE_URL`
    *   Custom delimiter support
    *   Leading/trailing delimiter handling
    *   Sequential array re-indexing

## 📦 Upgrade

```bash
composer update iperamuna/laravel-hypercacheio
```

This update is fully backward compatible. No configuration changes are required.

---

# Release Notes - v1.3.1

**Laravel Hyper-Cache-IO**

This patch release fixes a critical issue with distributed locking on Secondary nodes.

## 🐛 Bug Fixes

*   **Lock Reliability**: Fixed an issue where `Cache::lock()` and `release()` would fail silently on Secondary servers if `async_requests` was enabled (default). These operations are now forced to run synchronously to ensure they return the correct `true` or `false` result, regardless of the async configuration.

## 📦 Upgrade

```bash
composer update iperamuna/laravel-hypercacheio
```

This update is strongly recommended for all users operating in a Primary/Secondary cluster configuration.

---

# Release Notes - v1.3.0

**Laravel Hyper-Cache-IO**

This major update introduces a new Connectivity Check command, improved async processing architecture, and significant code optimizations.

## ✨ New Features

*   **Connectivity Check Command**: Run `php artisan hypercacheio:connectivity-check` to instantly verify communication between your Primary and Secondary nodes. It tests all endpoints (Ping, Add, Get, Put, Delete, Lock) and provides a detailed report.
*   **Fire-and-Forget Architecture**: Async requests now use a robust Promise tracking system with graceful shutdown. This ensures that write operations from Secondary nodes return immediately to the user while guaranteeing completion in the background, eliminating previous reliability issues with fire-and-forget requests.

## 🛠 Improvements

*   **Refactoring**: Core logic has been optimized by extracting shared SQLite operations into the `InteractsWithSqlite` trait, reducing code duplication and improving maintainability.
*   **Configuration Alignment**: The `CacheController` now respects the specific store configuration defined in `config/cache.php`, ensuring consistency with the `HypercacheioStore` implementation.
*   **Testing**: The entire test suite has been converted to **Pest PHP** for cleaner, more expressive tests.

## 📦 Upgrade

```bash
composer update iperamuna/laravel-hypercacheio
```

This update is fully backward compatible. For best results, ensure your configuration files are up to date.

---

# Release Notes - v1.2.1

**Laravel Hyper-Cache-IO**

This patch release adds names to the internal API routes for better identification and management within Laravel applications.

## ✨ New Features

*   **Named Routes**: All internal API routes now have names prefixed with `hypercacheio.`.
    *   `hypercacheio.cache.get`
    *   `hypercacheio.cache.add`
    *   `hypercacheio.cache.put`
    *   `hypercacheio.cache.forget`
    *   `hypercacheio.cache.lock`
    *   `hypercacheio.cache.release-lock`

This makes it easier to use these routes with Laravel's `route()` helper and improves visibility in tools like `php artisan route:list`.

## 📦 Upgrade

```bash
composer update iperamuna/laravel-hypercacheio
```

This update is drop-in compatible and recommended for all users.

---

# Release Notes - v1.2.0

**Laravel Hyper-Cache-IO**

This version introduces a complete rebrand of the package from **Hypercacheio** to **Hypercacheio**.

## 🔄 Rebranding & Renaming

*   **Namespace Change**: All classes are now under the `Iperamuna\Hypercacheio` namespace.
*   **Class Renaming**: Core classes have been renamed (e.g., `HypercacheioStore` -> `HypercacheioStore`).
*   **Configuration**: The config file is now `config/hypercacheio.php` and uses refined environment variables with the `HYPERCACHEIO_` prefix.
*   **Artisan Command**: The installation command is now `php artisan hypercacheio:install`.
*   **Routes**: Internal API routes now use the `/api/hypercacheio` prefix by default.

## 📦 Upgrade Guide

If you are upgrading from `v1.1.0` or earlier:

1.  **Update Composer**:
    ```bash
    composer require iperamuna/laravel-hypercacheio:^1.2
    ```
2.  **Update Configuration**:
    Rename your `config/hypercacheio.php` to `config/hypercacheio.php` and update the array keys. Alternatively, run the new install command:
    ```bash
    php artisan hypercacheio:install
    ```
3.  **Update Environment Variables**:
    Search your `.env` for `HYPERCACHEIO_` and replace with `HYPERCACHEIO_`.
4.  **Update Code References**:
    Any direct references to `Hypercacheio` classes or the `hypercacheio` cache driver should be updated to `Hypercacheio` and `hypercacheio` respectively.

## 🙏 Acknowledgements

Developed with ❤️ by [Indunil Peramuna](https://iperamuna.online).
