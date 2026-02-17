<?php

if (! function_exists('HypercacheioSecondaryUrls')) {

    /**
     * Parse a delimiter-separated string of secondary server URLs into a structured array.
     *
     * Takes a raw string (typically from an environment variable) containing one or more
     * URLs separated by the given delimiter, and returns an array of associative arrays
     * suitable for the Hypercacheio configuration's `secondaries` key.
     *
     * Each URL is trimmed of surrounding whitespace. Empty and whitespace-only entries
     * are silently discarded. Only entries that pass `FILTER_VALIDATE_URL` are included
     * in the result; invalid URLs are skipped to prevent misconfiguration.
     *
     * @param  string|null  $secondaryUrls  A delimiter-separated string of URLs, or null.
     * @param  string  $delimiter  The character used to separate URLs (default: ',').
     * @return array<int, array{url: string}> An indexed array of associative arrays, each containing a 'url' key.
     *
     * @example
     * // Single URL
     * HypercacheioSecondaryUrls('https://server2.example.com');
     * // Returns: [['url' => 'https://server2.example.com']]
     * @example
     * // Multiple URLs with whitespace
     * HypercacheioSecondaryUrls('https://s2.example.com , https://s3.example.com');
     * // Returns: [['url' => 'https://s2.example.com'], ['url' => 'https://s3.example.com']]
     * @example
     * // Null or empty input
     * HypercacheioSecondaryUrls(null);   // Returns: []
     * HypercacheioSecondaryUrls('');     // Returns: []
     */
    function HypercacheioSecondaryUrls(?string $secondaryUrls, string $delimiter = ','): array
    {
        if ($secondaryUrls === null || trim($secondaryUrls) === '') {
            return [];
        }

        return array_values(
            array_map(
                static fn (string $url): array => ['url' => $url],
                array_filter(
                    array_map('trim', explode($delimiter, $secondaryUrls)),
                    static fn (string $url): bool => $url !== '' && filter_var($url, FILTER_VALIDATE_URL) !== false
                )
            )
        );
    }
}
