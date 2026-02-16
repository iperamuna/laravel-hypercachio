# Release Notes - v1.0.1

**Laravel Hyper-Cache-IO**

We are pleased to release **v1.0.1**, which introduces improved CLI tooling for easier installation and enhanced CI reliability.

## 🚀 Enhancements

*   **New Install Command**: Running `php artisan hypercachio:install` now automates the setup process by publishing configuration files and automatically injecting the `hypercachio` store configuration into your `config/cache.php`, resolving issues where the cache store wasn't detected by other packages (e.g., Spatie Permission).
*   **Command Line Integration**: The installer handles `vendor:publish` and clears the configuration cache (`config:clear`) to ensure immediate availability of the new driver.

## 🛠 Fixes & Improvements

*   **CI/CD Stability**: Resolved "Driver not supported" errors in test environments by optimizing when and how the cache driver is registered in the Service Provider.
*   **Style Compliance**: Applied Laravel Pint formatting across the codebase to ensure PSR-12 standard compliance.
*   **Robust Testing**: Added a dedicated feature test (`InstallCommandTest`) to verify the installation command correctly modifies configuration files without duplication.

## 📦 Installation

To upgrade or install freshly:

```bash
composer require iperamuna/laravel-hypercachio
php artisan hypercachio:install
```

## 🙏 Acknowledgements

Developed with ❤️ by [Indunil Peramuna](https://iperamuna.online).
