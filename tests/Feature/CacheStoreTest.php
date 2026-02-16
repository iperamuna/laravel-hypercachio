<?php

namespace Iperamuna\Hypercachio\Tests\Feature;

use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Iperamuna\Hypercachio\HypercachioStore;
use Iperamuna\Hypercachio\Tests\TestCase;

class CacheStoreTest extends TestCase
{
    protected function tearDown(): void
    {
        // Clean up SQlite file
        $path = config('hypercachio.sqlite_path');
        if (file_exists($path)) {
            @unlink($path);
        }
        parent::tearDown();
    }

    public function test_primary_role_operations()
    {
        // Configured as primary in TestCase
        config(['hypercachio.role' => 'primary']);
        // Ensure async is false for immediate execution in tests
        config(['hypercachio.async_requests' => false]);

        // Clear cache manager stores to ensure fresh instance
        Cache::forgetDriver('hypercachio');

        $store = Cache::store('hypercachio')->getStore();
        $this->assertInstanceOf(HypercachioStore::class, $store);

        // Put -> Local Write
        Cache::store('hypercachio')->put('key1', 'value1', 60);

        // no HTTP request should be sent (mocking not strictly needed if we assert nothing sent, but good practice)
        Http::fake();
        Http::assertNothingSent();

        // Get -> Local Read
        $val = Cache::store('hypercachio')->get('key1');
        $this->assertEquals('value1', $val);

        // Add -> Atomic Insert
        $added = Cache::store('hypercachio')->add('key2', 'val2', 60);
        $this->assertTrue($added);

        $addedAgain = Cache::store('hypercachio')->add('key2', 'val2', 60);
        $this->assertFalse($addedAgain);
    }

    public function test_secondary_role_operations()
    {
        // Reconfigure as secondary
        config(['hypercachio.role' => 'secondary']);
        config(['hypercachio.primary_url' => 'http://test-primary/api/hypercachio']);
        config(['hypercachio.async_requests' => false]); // Force sync for ensuring requests are sent immediately

        // Clear cache manager stores
        Cache::forgetDriver('hypercachio');

        // Mock HTTP for Secondary
        Http::fake([
            '*/api/hypercachio/cache/key_sec' => Http::response(json_encode('value_sec'), 200),
            '*/api/hypercachio/cache/*' => Http::response(['success' => true], 200),
            '*/api/hypercachio/add/*' => Http::response(['added' => true], 200),
        ]);

        $store = Cache::store('hypercachio')->getStore();

        // Put -> HTTP Post
        Cache::store('hypercachio')->put('key_put', 'val_put', 60);

        Http::assertSent(function ($request) {
            return str_contains($request->url(), '/cache/key_put') && $request->method() === 'POST';
        });

        // Get -> HTTP Get (force sync inside store even if async was true, but we forced false anyway)
        $val = Cache::store('hypercachio')->get('key_sec');
        $this->assertEquals('value_sec', $val);

        // Add -> HTTP Post
        $added = Cache::store('hypercachio')->add('key_add', 'val_add', 60);
        $this->assertTrue($added);

        Http::assertSent(function ($request) {
            return str_contains($request->url(), '/add/key_add') && $request->method() === 'POST';
        });
    }
}
