---
layout: default
title: Advanced Features
description: This guide covers advanced features for optimizing performance and handling complex use cases.
---

## Overview

Optimize your application with these advanced capabilities:

**Batch Operations**: Execute multiple API requests efficiently
- `Batch`: Concurrent requests for maximum throughput
- `BatchSequential`: Ordered execution with easier error handling
- `BatchWithCallback`: Progress tracking for large operations

**Rate Limiting**: Stay within Twitch's API limits
- Automatic tracking via response headers
- Wait functions for graceful handling
- Retry logic with rate limit awareness

**Caching**: Reduce redundant API calls
- In-memory cache with configurable TTL
- Custom cache implementations (Redis, etc.)
- Cache key isolation for multi-tenant apps

**Middleware**: Extend client functionality
- Request/response logging
- Automatic retries with backoff
- Custom headers and request modification

## Prerequisites

These advanced features work with any authenticated client. Ensure you have:
- A valid `AuthClient` configured with appropriate credentials
- Required scopes for the specific API endpoints you're calling

## Batch Operations

Execute multiple API requests efficiently with built-in concurrency control.

### Basic Batch

```go
// Execute multiple requests concurrently
results := client.Batch(ctx, []helix.BatchRequest{
    {
        Request: &helix.Request{
            Method:   "GET",
            Endpoint: "/users",
            Query:    url.Values{"id": []string{"123"}},
        },
        Result: &user1Response,
    },
    {
        Request: &helix.Request{
            Method:   "GET",
            Endpoint: "/users",
            Query:    url.Values{"id": []string{"456"}},
        },
        Result: &user2Response,
    },
}, nil)

// Check for errors
if helix.HasErrors(results) {
    for _, r := range results {
        if r.Error != nil {
            log.Printf("Request %d failed: %v", r.Index, r.Error)
        }
    }
}
```

### Batch Options

```go
opts := &helix.BatchOptions{
    MaxConcurrent: 5,    // Limit to 5 concurrent requests (0 = unlimited)
    StopOnError:   true, // Stop on first error
}

results := client.Batch(ctx, requests, opts)
```

### BatchGet (Convenience)

```go
// Simplified batch GET requests
results := client.BatchGet(ctx, []helix.GetRequest{
    {Endpoint: "/users", Query: url.Values{"id": []string{"123"}}, Result: &user1},
    {Endpoint: "/users", Query: url.Values{"id": []string{"456"}}, Result: &user2},
    {Endpoint: "/streams", Query: url.Values{"user_id": []string{"123"}}, Result: &stream1},
}, nil)
```

### BatchSequential

```go
// Execute requests one at a time (preserves order, easier error handling)
results := client.BatchSequential(ctx, requests)

for _, r := range results {
    if r.Error != nil {
        // Stop processing on error
        break
    }
}
```

### BatchWithCallback

```go
// Process results as they complete
var mu sync.Mutex
completed := 0

client.BatchWithCallback(ctx, requests, nil, func(result helix.BatchResult) {
    mu.Lock()
    defer mu.Unlock()
    completed++

    if result.Error != nil {
        log.Printf("Request %d failed: %v", result.Index, result.Error)
    } else {
        log.Printf("Request %d completed (%d/%d)", result.Index, completed, len(requests))
    }
})
```

### Helper Functions

```go
// Check if any request failed
if helix.HasErrors(results) { ... }

// Get first error (or nil)
if err := helix.FirstError(results); err != nil { ... }

// Get all errors
for _, err := range helix.Errors(results) { ... }
```

## Rate Limiting

The client tracks rate limit information from API responses.

### Checking Rate Limits

```go
// Get current rate limit info
info := client.GetRateLimitInfo()
fmt.Printf("Remaining: %d/%d, Resets at: %v\n",
    info.Remaining, info.Limit, info.Reset)
```

### Waiting for Rate Limits

```go
// Wait until rate limit resets (with context for cancellation)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := client.WaitForRateLimit(ctx); err != nil {
    log.Printf("Wait cancelled: %v", err)
}
```

### Automatic Rate Limit Handling

```go
// Implement retry logic with rate limit awareness
func doRequestWithRetry(ctx context.Context, fn func() error) error {
    for i := 0; i < 3; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        // Check if rate limited
        var apiErr *helix.APIError
        if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
            if waitErr := client.WaitForRateLimit(ctx); waitErr != nil {
                return waitErr
            }
            continue
        }

        return err
    }
    return errors.New("max retries exceeded")
}
```

## Caching

The client supports caching API responses to reduce redundant requests.

### Enabling Cache

```go
// Use built-in memory cache
cache := helix.NewMemoryCache(1000) // Max 1000 entries
client := helix.NewClient(authClient, helix.WithCache(cache))
```

### Cache Operations

```go
// Clear entire cache
client.ClearCache()

// Invalidate specific endpoint
client.InvalidateCache("/users")

// Invalidate with token context (for multi-tenant scenarios)
client.InvalidateCacheWithContext(ctx, "/users")
```

### Custom Cache Implementation

```go
// Implement the Cache interface
type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte, ttl time.Duration)
    Delete(key string)
    Clear()
}

// Example: Redis cache
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(key string) ([]byte, bool) {
    val, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        return nil, false
    }
    return val, true
}

// ... implement other methods

client := helix.NewClient(authClient, helix.WithCache(&RedisCache{client: redisClient}))
```

### Cache Key Generation

```go
// Generate cache keys with token isolation
key := helix.CacheKeyWithContext(ctx, "GET", "/users", query)

// Generate token hash for cache isolation
hash := helix.TokenHash(token)
```

## Middleware

Add custom middleware to intercept and modify requests/responses.

### Adding Middleware

```go
// Middleware function signature
type Middleware func(next http.RoundTripper) http.RoundTripper

// Add logging middleware
client.Use(func(next http.RoundTripper) http.RoundTripper {
    return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
        start := time.Now()

        resp, err := next.RoundTrip(req)

        log.Printf("%s %s - %v (%v)",
            req.Method, req.URL.Path, resp.StatusCode, time.Since(start))

        return resp, err
    })
})
```

### Common Middleware Examples

#### Request Logging

```go
func LoggingMiddleware(next http.RoundTripper) http.RoundTripper {
    return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
        log.Printf("-> %s %s", req.Method, req.URL)
        resp, err := next.RoundTrip(req)
        if resp != nil {
            log.Printf("<- %d %s", resp.StatusCode, req.URL)
        }
        return resp, err
    })
}
```

#### Retry Middleware

```go
func RetryMiddleware(maxRetries int) Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
            var resp *http.Response
            var err error

            for i := 0; i <= maxRetries; i++ {
                resp, err = next.RoundTrip(req)
                if err == nil && resp.StatusCode < 500 {
                    return resp, nil
                }

                if i < maxRetries {
                    time.Sleep(time.Duration(i+1) * time.Second)
                }
            }

            return resp, err
        })
    }
}
```

#### Custom Headers

```go
func CustomHeadersMiddleware(headers map[string]string) Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
            for k, v := range headers {
                req.Header.Set(k, v)
            }
            return next.RoundTrip(req)
        })
    }
}
```

## Low-Level Request Execution

For advanced use cases, you can execute raw requests directly.

### Using Do()

```go
// Execute a custom request
var result MyCustomResponse
err := client.Do(ctx, &helix.Request{
    Method:   "POST",
    Endpoint: "/some/endpoint",
    Query:    url.Values{"param": []string{"value"}},
    Body:     requestBody,
}, &result)
```

### Request Structure

```go
type Request struct {
    Method   string      // HTTP method (GET, POST, PUT, PATCH, DELETE)
    Endpoint string      // API endpoint path (e.g., "/users")
    Query    url.Values  // Query parameters
    Body     interface{} // Request body (will be JSON encoded)
}
```

## Performance Tips

### 1. Use Batch Requests

```go
// Instead of multiple sequential calls:
user1, _ := client.GetUsers(ctx, &helix.GetUsersParams{IDs: []string{"123"}})
user2, _ := client.GetUsers(ctx, &helix.GetUsersParams{IDs: []string{"456"}})

// Use a single call with multiple IDs:
users, _ := client.GetUsers(ctx, &helix.GetUsersParams{IDs: []string{"123", "456"}})
```

### 2. Enable Caching for Repeated Requests

```go
// Cache frequently accessed data
cache := helix.NewMemoryCache(1000)
client := helix.NewClient(authClient, helix.WithCache(cache))

// First call hits API
users, _ := client.GetUsers(ctx, params)

// Second call uses cache
users, _ := client.GetUsers(ctx, params) // No API call
```

### 3. Limit Concurrent Requests

```go
// Avoid overwhelming the API
opts := &helix.BatchOptions{
    MaxConcurrent: 10, // Twitch recommends limiting concurrent requests
}
```

### 4. Use Context for Timeouts

```go
// Set reasonable timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.GetUsers(ctx, params)
```

## See Also

- [Quick Start](quickstart.md) - Basic setup and usage
- [Auth](auth.md) - Authentication and token management
- [EventSub](eventsub.md) - Real-time event subscriptions

