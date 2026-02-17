<?php

require_once __DIR__.'/../../helpers/HypercacheioSecondaryUrls.php';

it('returns an empty array for null input', function () {
    $result = HypercacheioSecondaryUrls(null);

    expect($result)->toBe([]);
});

it('returns an empty array for an empty string', function () {
    $result = HypercacheioSecondaryUrls('');

    expect($result)->toBe([]);
});

it('returns an empty array for a whitespace-only string', function () {
    $result = HypercacheioSecondaryUrls('   ');

    expect($result)->toBe([]);
});

it('parses a single valid URL', function () {
    $result = HypercacheioSecondaryUrls('https://server2.example.com');

    expect($result)->toBe([
        ['url' => 'https://server2.example.com'],
    ]);
});

it('parses multiple comma-separated URLs', function () {
    $result = HypercacheioSecondaryUrls('https://s2.example.com,https://s3.example.com');

    expect($result)->toBe([
        ['url' => 'https://s2.example.com'],
        ['url' => 'https://s3.example.com'],
    ]);
});

it('trims whitespace around URLs', function () {
    $result = HypercacheioSecondaryUrls('  https://s2.example.com , https://s3.example.com  ');

    expect($result)->toBe([
        ['url' => 'https://s2.example.com'],
        ['url' => 'https://s3.example.com'],
    ]);
});

it('filters out empty entries from consecutive delimiters', function () {
    $result = HypercacheioSecondaryUrls('https://s2.example.com,,https://s3.example.com');

    expect($result)->toBe([
        ['url' => 'https://s2.example.com'],
        ['url' => 'https://s3.example.com'],
    ]);
});

it('filters out invalid URLs', function () {
    $result = HypercacheioSecondaryUrls('https://valid.example.com,not-a-url,https://also-valid.example.com');

    expect($result)->toBe([
        ['url' => 'https://valid.example.com'],
        ['url' => 'https://also-valid.example.com'],
    ]);
});

it('returns an empty array when all URLs are invalid', function () {
    $result = HypercacheioSecondaryUrls('not-valid,also-bad,nope');

    expect($result)->toBe([]);
});

it('supports a custom delimiter', function () {
    $result = HypercacheioSecondaryUrls('https://s2.example.com|https://s3.example.com', '|');

    expect($result)->toBe([
        ['url' => 'https://s2.example.com'],
        ['url' => 'https://s3.example.com'],
    ]);
});

it('handles a trailing delimiter gracefully', function () {
    $result = HypercacheioSecondaryUrls('https://s2.example.com,https://s3.example.com,');

    expect($result)->toBe([
        ['url' => 'https://s2.example.com'],
        ['url' => 'https://s3.example.com'],
    ]);
});

it('handles a leading delimiter gracefully', function () {
    $result = HypercacheioSecondaryUrls(',https://s2.example.com,https://s3.example.com');

    expect($result)->toBe([
        ['url' => 'https://s2.example.com'],
        ['url' => 'https://s3.example.com'],
    ]);
});

it('re-indexes the result array sequentially', function () {
    $result = HypercacheioSecondaryUrls('invalid,,https://valid.example.com');

    expect($result)->toBe([
        ['url' => 'https://valid.example.com'],
    ]);
    expect(array_keys($result))->toBe([0]);
});
