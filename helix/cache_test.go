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

func TestCacheKeyWithContext(t *testing.T) {
	// Same endpoint/query but different baseURL should produce different keys
	key1 := CacheKeyWithContext("https://api.twitch.tv", "/users", "id=123", "")
	key2 := CacheKeyWithContext("https://api.test.tv", "/users", "id=123", "")
	if key1 == key2 {
		t.Error("different baseURL should produce different keys")
	}

	// Same endpoint/query but different tokens should produce different keys
	key3 := CacheKeyWithContext("https://api.twitch.tv", "/users", "id=123", "token1hash")
	key4 := CacheKeyWithContext("https://api.twitch.tv", "/users", "id=123", "token2hash")
	if key3 == key4 {
		t.Error("different tokens should produce different keys")
	}

	// Same everything should produce same key
	key5 := CacheKeyWithContext("https://api.twitch.tv", "/users", "id=123", "samehash")
	key6 := CacheKeyWithContext("https://api.twitch.tv", "/users", "id=123", "samehash")
	if key5 != key6 {
		t.Error("same inputs should produce same key")
	}

	// Empty token hash should still work
	key7 := CacheKeyWithContext("https://api.twitch.tv", "/users", "id=123", "")
	if len(key7) != 64 {
		t.Errorf("expected 64 character hex hash, got %d", len(key7))
	}
}

func TestTokenHash(t *testing.T) {
	// Different tokens should produce different hashes
	hash1 := TokenHash("token1")
	hash2 := TokenHash("token2")
	if hash1 == hash2 {
		t.Error("different tokens should produce different hashes")
	}

	// Same token should produce same hash
	hash3 := TokenHash("same-token")
	hash4 := TokenHash("same-token")
	if hash3 != hash4 {
		t.Error("same token should produce same hash")
	}

	// Empty token should return empty string
	hash5 := TokenHash("")
	if hash5 != "" {
		t.Error("empty token should return empty hash")
	}

	// Hash should be 32 characters (128 bits = 16 bytes = 32 hex chars)
	if len(hash1) != 32 {
		t.Errorf("expected 32 character hex hash, got %d", len(hash1))
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

func TestMemoryCache_MutationProtection(t *testing.T) {
	cache := NewMemoryCache(0)
	ctx := context.Background()

	// Test that mutating the input slice doesn't affect cached data
	t.Run("input mutation", func(t *testing.T) {
		original := []byte("original value")
		cache.Set(ctx, "key1", original, time.Minute)

		// Mutate the original slice
		original[0] = 'X'

		// Get should return the unmutated value
		result := cache.Get(ctx, "key1")
		if string(result) != "original value" {
			t.Errorf("expected 'original value', got '%s'", string(result))
		}
	})

	// Test that mutating the returned slice doesn't affect cached data
	t.Run("output mutation", func(t *testing.T) {
		cache.Set(ctx, "key2", []byte("cached value"), time.Minute)

		// Get and mutate the result
		result1 := cache.Get(ctx, "key2")
		result1[0] = 'X'

		// Get again - should still have original value
		result2 := cache.Get(ctx, "key2")
		if string(result2) != "cached value" {
			t.Errorf("expected 'cached value', got '%s'", string(result2))
		}
	})

	// Test that two Gets return independent copies
	t.Run("independent copies", func(t *testing.T) {
		cache.Set(ctx, "key3", []byte("test"), time.Minute)

		result1 := cache.Get(ctx, "key3")
		result2 := cache.Get(ctx, "key3")

		// Mutate first result
		result1[0] = 'X'

		// Second result should be unaffected
		if string(result2) != "test" {
			t.Errorf("expected 'test', got '%s'", string(result2))
		}
	})
}
