package helix

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Cache is an interface for caching API responses.
type Cache interface {
	// Get retrieves a cached response. Returns nil if not found or expired.
	Get(ctx context.Context, key string) []byte
	// Set stores a response in the cache with the given TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	// Delete removes a cached response.
	Delete(ctx context.Context, key string)
	// Clear removes all cached responses.
	Clear(ctx context.Context)
}

// MemoryCache is an in-memory cache implementation.
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	maxSize int // Maximum number of entries (0 = unlimited)
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]*cacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves a cached response.
func (c *MemoryCache) Get(ctx context.Context, key string) []byte {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		c.Delete(ctx, key)
		return nil
	}

	return entry.value
}

// Set stores a response in the cache.
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict expired entries if at max size
	if c.maxSize > 0 && len(c.entries) >= c.maxSize {
		c.evictExpired()
		// If still at max, evict oldest
		if len(c.entries) >= c.maxSize {
			c.evictOldest()
		}
	}

	c.entries[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a cached response.
func (c *MemoryCache) Delete(ctx context.Context, key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

// Clear removes all cached responses.
func (c *MemoryCache) Clear(ctx context.Context) {
	c.mu.Lock()
	c.entries = make(map[string]*cacheEntry)
	c.mu.Unlock()
}

// evictExpired removes all expired entries (must be called with lock held).
// Note: Deleting from a map during iteration is safe in Go when done in the same goroutine.
func (c *MemoryCache) evictExpired() {
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}

// evictOldest removes the oldest entry (must be called with lock held).
func (c *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// Size returns the number of entries in the cache.
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// CacheKey generates a cache key from a request.
func CacheKey(endpoint string, query string) string {
	hash := sha256.Sum256([]byte(endpoint + "?" + query))
	return hex.EncodeToString(hash[:])
}

// WithCache sets a cache for the client.
func WithCache(cache Cache, ttl time.Duration) Option {
	return func(c *Client) {
		c.cache = cache
		c.cacheTTL = ttl
		c.cacheEnabled = true
	}
}

// WithCacheEnabled enables or disables caching.
func WithCacheEnabled(enabled bool) Option {
	return func(c *Client) {
		c.cacheEnabled = enabled
	}
}

// ClearCache clears all cached responses.
func (c *Client) ClearCache(ctx context.Context) {
	if c.cache != nil {
		c.cache.Clear(ctx)
	}
}

// InvalidateCache removes a specific cached response.
func (c *Client) InvalidateCache(ctx context.Context, endpoint string, query string) {
	if c.cache != nil {
		c.cache.Delete(ctx, CacheKey(endpoint, query))
	}
}

// NoCacheContext returns a context that bypasses the cache for a single request.
func NoCacheContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, noCacheKey{}, true)
}

type noCacheKey struct{}

func shouldSkipCache(ctx context.Context) bool {
	if v, ok := ctx.Value(noCacheKey{}).(bool); ok {
		return v
	}
	return false
}
