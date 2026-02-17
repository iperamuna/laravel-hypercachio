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
