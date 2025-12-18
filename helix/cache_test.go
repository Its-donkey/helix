package helix

import (
	"context"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache(100)
	if cache == nil {
		t.Fatal("expected cache to not be nil")
	}
	if cache.maxSize != 100 {
		t.Errorf("expected maxSize 100, got %d", cache.maxSize)
	}
	if cache.entries == nil {
		t.Error("expected entries map to be initialized")
	}
}

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(0)
	ctx := context.Background()

	// Test set and get
	cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	result := cache.Get(ctx, "key1")
	if string(result) != "value1" {
		t.Errorf("expected value1, got %s", string(result))
	}

	// Test get non-existent key
	result = cache.Get(ctx, "nonexistent")
	if result != nil {
		t.Errorf("expected nil for non-existent key, got %v", result)
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(0)
	ctx := context.Background()

	// Set with very short TTL
	cache.Set(ctx, "expiring", []byte("value"), time.Millisecond)

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	// Should return nil for expired entry
	result := cache.Get(ctx, "expiring")
	if result != nil {
		t.Errorf("expected nil for expired key, got %v", result)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(0)
	ctx := context.Background()

	cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	cache.Delete(ctx, "key1")

	result := cache.Get(ctx, "key1")
	if result != nil {
		t.Errorf("expected nil after delete, got %v", result)
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(0)
	ctx := context.Background()

	cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	cache.Set(ctx, "key2", []byte("value2"), time.Minute)
	cache.Clear(ctx)

	if cache.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", cache.Size())
	}
}

func TestMemoryCache_Size(t *testing.T) {
	cache := NewMemoryCache(0)
	ctx := context.Background()

	if cache.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", cache.Size())
	}

	cache.Set(ctx, "key1", []byte("value1"), time.Minute)
	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}

	cache.Set(ctx, "key2", []byte("value2"), time.Minute)
	if cache.Size() != 2 {
		t.Errorf("expected size 2, got %d", cache.Size())
	}
}

func TestMemoryCache_MaxSize_EvictExpired(t *testing.T) {
	cache := NewMemoryCache(2)
	ctx := context.Background()

	// Add entry with short TTL
	cache.Set(ctx, "expiring", []byte("value"), time.Millisecond)
	cache.Set(ctx, "permanent", []byte("value"), time.Minute)

	// Wait for first entry to expire
	time.Sleep(5 * time.Millisecond)

	// Adding third entry should evict the expired one
	cache.Set(ctx, "new", []byte("value"), time.Minute)

	// Should have evicted expired entry
	if cache.Size() != 2 {
		t.Errorf("expected size 2 after eviction, got %d", cache.Size())
	}

	// Expired entry should be gone
	if cache.Get(ctx, "expiring") != nil {
		t.Error("expected expired entry to be evicted")
	}
}

func TestMemoryCache_MaxSize_EvictOldest(t *testing.T) {
	cache := NewMemoryCache(2)
	ctx := context.Background()

	// Add two entries
	cache.Set(ctx, "first", []byte("value1"), time.Minute)
	time.Sleep(time.Millisecond) // Ensure different timestamps
	cache.Set(ctx, "second", []byte("value2"), 2*time.Minute)

	// Adding third entry should evict the oldest (first has earlier expiry)
	cache.Set(ctx, "third", []byte("value3"), time.Minute)

	if cache.Size() != 2 {
		t.Errorf("expected size 2, got %d", cache.Size())
	}

	// First entry should be evicted (has earliest expiry)
	if cache.Get(ctx, "first") != nil {
		t.Error("expected first entry to be evicted")
	}
}

func TestCacheKey(t *testing.T) {
	key1 := CacheKey("/users", "id=123")
	key2 := CacheKey("/users", "id=123")
	key3 := CacheKey("/users", "id=456")

	if key1 != key2 {
		t.Error("expected same inputs to produce same key")
	}
	if key1 == key3 {
		t.Error("expected different inputs to produce different keys")
	}
	if len(key1) != 64 {
		t.Errorf("expected 64 character hex hash, got %d", len(key1))
	}
}

func TestWithCache(t *testing.T) {
	cache := NewMemoryCache(100)
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	client := NewClient("test-client-id", authClient, WithCache(cache, time.Minute))

	if client.cache != cache {
		t.Error("expected cache to be set")
	}
	if client.cacheTTL != time.Minute {
		t.Errorf("expected cacheTTL of 1 minute, got %v", client.cacheTTL)
	}
	if !client.cacheEnabled {
		t.Error("expected cacheEnabled to be true")
	}
}

func TestWithCacheEnabled(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	client := NewClient("test-client-id", authClient, WithCacheEnabled(true))
	if !client.cacheEnabled {
		t.Error("expected cacheEnabled to be true")
	}

	client = NewClient("test-client-id", authClient, WithCacheEnabled(false))
	if client.cacheEnabled {
		t.Error("expected cacheEnabled to be false")
	}
}

func TestClient_ClearCache(t *testing.T) {
	cache := NewMemoryCache(100)
	ctx := context.Background()
	cache.Set(ctx, "key1", []byte("value1"), time.Minute)

	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	client := NewClient("test-client-id", authClient, WithCache(cache, time.Minute))

	client.ClearCache(ctx)

	if cache.Size() != 0 {
		t.Errorf("expected cache to be empty, got size %d", cache.Size())
	}
}

func TestClient_ClearCache_NoCache(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	client := NewClient("test-client-id", authClient)

	// Should not panic when no cache is set
	client.ClearCache(context.Background())
}

func TestClient_InvalidateCache(t *testing.T) {
	cache := NewMemoryCache(100)
	ctx := context.Background()

	key := CacheKey("/users", "id=123")
	cache.Set(ctx, key, []byte("cached"), time.Minute)

	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	client := NewClient("test-client-id", authClient, WithCache(cache, time.Minute))

	client.InvalidateCache(ctx, "/users", "id=123")

	if cache.Get(ctx, key) != nil {
		t.Error("expected cache entry to be invalidated")
	}
}

func TestClient_InvalidateCache_NoCache(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	client := NewClient("test-client-id", authClient)

	// Should not panic when no cache is set
	client.InvalidateCache(context.Background(), "/users", "id=123")
}

func TestNoCacheContext(t *testing.T) {
	ctx := context.Background()
	noCacheCtx := NoCacheContext(ctx)

	if shouldSkipCache(ctx) {
		t.Error("expected regular context to not skip cache")
	}
	if !shouldSkipCache(noCacheCtx) {
		t.Error("expected NoCacheContext to skip cache")
	}
}

func TestShouldSkipCache_NoValue(t *testing.T) {
	ctx := context.Background()
	if shouldSkipCache(ctx) {
		t.Error("expected context without value to not skip cache")
	}
}
