<?php

use Illuminate\Support\Facades\Route;
use Iperamuna\Hypercachio\Http\Controllers\CacheController;
use Iperamuna\Hypercachio\Http\Middleware\HyperCachioSecurity;

Route::prefix(config('hypercachio.api_url'))
    ->middleware([HyperCachioSecurity::class])
    ->group(function () {
        Route::get('/cache/{key}', [CacheController::class, 'get']);
        Route::post('/add/{key}', [CacheController::class, 'add']);
        Route::post('/cache/{key}', [CacheController::class, 'put']);
        Route::delete('/cache/{key}', [CacheController::class, 'forget']);
        Route::post('/lock/{key}', [CacheController::class, 'lock']);
        Route::delete('/lock/{key}', [CacheController::class, 'releaseLock']);
    });
