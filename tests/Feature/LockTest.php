<?php

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Iperamuna\Hypercacheio\HypercacheioStore;

it('can acquire and release atomic locks', function () {
    $lock = Cache::store('hypercacheio')->lock('foo', 10);

    expect($lock->get())->toBeTrue();

    // Lock is held
    $lock2 = Cache::store('hypercacheio')->lock('foo', 10);
    expect($lock2->get())->toBeFalse();

    // Release
    $lock->release();

    // Can acquire again
    expect($lock2->get())->toBeTrue();
    $lock2->release();
});

it('can manipulate locks via owner', function () {
    // Get the underlying store
    $repository = Cache::store('hypercacheio');
    $store = $repository->getStore();

    if (! $store instanceof HypercacheioStore) {
        $this->markTestSkipped('Not using HypercacheioStore');
    }

    // Manually acquire
    $owner = 'test-owner';
    $result = $store->acquireLock('bar', $owner, 10);
    expect($result)->toBeTrue();

    // Check owner
    // Note: getLockOwner applies prefix, acquireLock applied prefix.
    // So 'bar' passed to getLockOwner becomes 'prefix'.'bar'.
    // Logic inside store handles prefix consistently.
    expect($store->getLockOwner('bar'))->toBe('test-owner');

    // Release
    $result = $store->releaseLock('bar', $owner);
    expect($result)->toBeTrue();

    // Check owner gone
    // SQLite fetchColumn returns false if no row, I implemented logic to return ''
    expect($store->getLockOwner('bar'))->toBeEmpty();
});

it('can acquire lock via secondary HTTP call even if async is enabled', function () {
    // Reconfigure as secondary
    config(['hypercacheio.role' => 'secondary']);
    config(['hypercacheio.primary_url' => 'http://test-primary/api/hypercacheio']);
    config(['hypercacheio.async_requests' => true]); // Force async true!
    config(['cache.prefix' => '']);

    // Mock HTTP for Secondary
    Http::fake([
        '*/api/hypercacheio/lock/*' => Http::response(['acquired' => true, 'released' => true], 200),
    ]);

    // Clear cache store to reload config
    Cache::forgetDriver('hypercacheio');

    $lock = Cache::store('hypercacheio')->lock('foo', 10);

    // If async is skipped for lock, this should call Http synchronously and return true
    expect($lock->get())->toBeTrue();

    // Release
    // Should call Http synchronously
    $lock->release(); // Should not throw
});
