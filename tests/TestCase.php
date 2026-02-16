<?php

namespace Iperamuna\Hypercachio\Tests;

use Tests\TestCase as BaseTestCase;

class TestCase extends BaseTestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        config([
            'hypercachio.role' => 'primary',
            'hypercachio.primary_url' => 'http://test-server.test/api/hypercachio',
            'hypercachio.async_requests' => false,
            'hypercachio.sqlite_path' => __DIR__.'/temp/cache.sqlite',
            'cache.stores.hypercachio' => [
                'driver' => 'hypercachio',
            ],
        ]);
    }
}
