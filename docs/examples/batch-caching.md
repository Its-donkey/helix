---
layout: default
title: Batch Operations & Caching Examples
description: Efficient API usage with batch requests, caching, rate limiting, and middleware.
---

## Overview

When building applications that make many API calls, optimization is essential to stay within rate limits and provide responsive experiences.

**Batch Requests**: Execute multiple operations efficiently
- `Batch`: Concurrent requests for maximum throughput
- `BatchGet`: Concurrent with built-in rate limiting
- `BatchSequential`: Sequential execution for ordered operations
- `BatchWithCallback`: Progress tracking for large operations

**Caching**: Reduce redundant API calls
- In-memory cache with configurable TTL
- Cache key isolation for multi-tenant apps
- Manual invalidation when needed

**Rate Limiting**: Stay within Twitch's limits (800 requests/minute)
- Automatic tracking via response headers
- Wait functions for graceful handling

**Middleware**: Extend client functionality
- Logging, retries, metrics, custom headers
- Chainable for complex behaviors

## Batch Requests

Execute multiple API calls concurrently for maximum throughput:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Batch get users (concurrent requests)
    userLogins := []string{"user1", "user2", "user3", "user4", "user5"}

    results, err := helix.Batch(ctx, userLogins, func(ctx context.Context, login string) (*helix.GetUsersResponse, error) {
        return client.GetUsers(ctx, &helix.GetUsersParams{
            Logins: []string{login},
        })
    })
    if err != nil {
        log.Fatal(err)
    }

    for i, result := range results {
        if result.Error != nil {
            fmt.Printf("Failed to get user %s: %v\n", userLogins[i], result.Error)
            continue
        }
        if len(result.Value.Data) > 0 {
            fmt.Printf("User: %s (ID: %s)\n", result.Value.Data[0].DisplayName, result.Value.Data[0].ID)
        }
    }
}
```

## BatchGet with Built-in Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Get multiple channel information with automatic rate limiting
    broadcasterIDs := []string{"12345", "67890", "11111", "22222", "33333"}

    results, err := helix.BatchGet(ctx, client, broadcasterIDs,
        func(ctx context.Context, client *helix.Client, id string) (*helix.GetChannelInformationResponse, error) {
            return client.GetChannelInformation(ctx, &helix.GetChannelInformationParams{
                BroadcasterIDs: []string{id},
            })
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("Error: %v\n", result.Error)
            continue
        }
        if len(result.Value.Data) > 0 {
            ch := result.Value.Data[0]
            fmt.Printf("Channel: %s - %s\n", ch.BroadcasterName, ch.Title)
        }
    }
}
```

## Sequential Batch (Rate Limited)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Process sequentially (useful when order matters or for heavy operations)
    userIDs := []string{"12345", "67890", "11111"}

    results, err := helix.BatchSequential(ctx, userIDs, func(ctx context.Context, id string) (*helix.GetUsersResponse, error) {
        return client.GetUsers(ctx, &helix.GetUsersParams{
            IDs: []string{id},
        })
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        if result.Error == nil && len(result.Value.Data) > 0 {
            fmt.Printf("User: %s\n", result.Value.Data[0].DisplayName)
        }
    }
}
```

## Batch with Callback

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync/atomic"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    userIDs := make([]string, 100)
    for i := range userIDs {
        userIDs[i] = fmt.Sprintf("%d", 10000+i)
    }

    var processed int64

    // Process with progress callback
    err := helix.BatchWithCallback(ctx, userIDs,
        func(ctx context.Context, id string) (*helix.GetUsersResponse, error) {
            return client.GetUsers(ctx, &helix.GetUsersParams{
                IDs: []string{id},
            })
        },
        func(input string, result *helix.GetUsersResponse, err error) {
            atomic.AddInt64(&processed, 1)
            current := atomic.LoadInt64(&processed)
            if current%10 == 0 {
                fmt.Printf("Progress: %d/%d\n", current, len(userIDs))
            }
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Processed %d users\n", processed)
}
```

## Caching Setup

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    // Create client with caching enabled
    cache := helix.NewMemoryCache(5 * time.Minute) // 5 minute TTL
    client := helix.NewClient("your-client-id", authClient,
        helix.WithCache(cache),
    )

    // First request - hits API
    start := time.Now()
    users1, _ := client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })
    fmt.Printf("First request: %v (API call)\n", time.Since(start))

    // Second request - hits cache
    start = time.Now()
    users2, _ := client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })
    fmt.Printf("Second request: %v (cached)\n", time.Since(start))

    fmt.Printf("Same data: %v\n", users1.Data[0].ID == users2.Data[0].ID)

    // Invalidate cache for specific endpoint
    client.InvalidateCache(ctx, "users")

    // Third request - hits API again
    start = time.Now()
    _, _ = client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })
    fmt.Printf("Third request: %v (API call after invalidation)\n", time.Since(start))
}
```

## Custom Cache Implementation

```go
package main

import (
    "context"
    "sync"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

// RedisCache implements helix.Cache using Redis
type RedisCache struct {
    // Add your Redis client here
    mu sync.RWMutex
    data map[string]cacheEntry
}

type cacheEntry struct {
    value     []byte
    expiresAt time.Time
}

func NewRedisCache() *RedisCache {
    return &RedisCache{
        data: make(map[string]cacheEntry),
    }
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.data[key]
    if !ok || time.Now().After(entry.expiresAt) {
        return nil, false
    }
    return entry.value, true
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.data[key] = cacheEntry{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    }
}

func (c *RedisCache) Delete(ctx context.Context, key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.data, key)
}

func main() {
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    cache := NewRedisCache()
    client := helix.NewClient("your-client-id", authClient,
        helix.WithCache(cache),
        helix.WithCacheTTL(10*time.Minute),
    )

    // Use client normally - caching is automatic
    _ = client
}
```

## Cache Key Isolation

```go
package main

import (
    "context"
    "fmt"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Generate cache key with context isolation
    // This prevents cache collisions between different tokens/environments
    key := helix.CacheKeyWithContext("users", "https://api.twitch.tv", "access-token-123")
    fmt.Printf("Cache key: %s\n", key)

    // Generate token hash for cache isolation
    hash := helix.TokenHash("access-token-123")
    fmt.Printf("Token hash: %s\n", hash)

    // Invalidate cache with context
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID: "your-client-id",
    })
    cache := helix.NewMemoryCache(5 * time.Minute)
    client := helix.NewClient("your-client-id", authClient, helix.WithCache(cache))

    client.InvalidateCacheWithContext(ctx, "users")
}
```

## Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Make a request
    _, _ = client.GetUsers(ctx, &helix.GetUsersParams{
        Logins: []string{"twitchdev"},
    })

    // Check rate limit status
    rateLimit := client.GetRateLimitInfo()
    fmt.Printf("Rate Limit: %d/%d\n", rateLimit.Remaining, rateLimit.Limit)
    fmt.Printf("Resets at: %s\n", rateLimit.Reset.Format(time.RFC3339))

    // Wait if rate limited
    if rateLimit.Remaining < 10 {
        waitTime := client.WaitForRateLimit(ctx)
        if waitTime > 0 {
            fmt.Printf("Rate limited, waiting %v\n", waitTime)
            time.Sleep(waitTime)
        }
    }

    // Continue making requests...
}
```

## Middleware

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

// LoggingMiddleware logs all API requests
func LoggingMiddleware(next helix.RoundTripper) helix.RoundTripper {
    return helix.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
        start := time.Now()
        log.Printf("→ %s %s", req.Method, req.URL.Path)

        resp, err := next.RoundTrip(req)

        duration := time.Since(start)
        if err != nil {
            log.Printf("← ERROR: %v (%v)", err, duration)
        } else {
            log.Printf("← %d %s (%v)", resp.StatusCode, resp.Status, duration)
        }

        return resp, err
    })
}

// RetryMiddleware retries failed requests
func RetryMiddleware(maxRetries int) helix.Middleware {
    return func(next helix.RoundTripper) helix.RoundTripper {
        return helix.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
            var resp *http.Response
            var err error

            for i := 0; i <= maxRetries; i++ {
                resp, err = next.RoundTrip(req)
                if err == nil && resp.StatusCode < 500 {
                    return resp, nil
                }

                if i < maxRetries {
                    time.Sleep(time.Duration(i+1) * time.Second)
                    log.Printf("Retrying request (attempt %d/%d)", i+2, maxRetries+1)
                }
            }

            return resp, err
        })
    }
}

// MetricsMiddleware tracks request metrics
type Metrics struct {
    TotalRequests int64
    Errors        int64
    TotalLatency  time.Duration
}

func MetricsMiddleware(metrics *Metrics) helix.Middleware {
    return func(next helix.RoundTripper) helix.RoundTripper {
        return helix.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
            start := time.Now()
            metrics.TotalRequests++

            resp, err := next.RoundTrip(req)

            metrics.TotalLatency += time.Since(start)
            if err != nil || (resp != nil && resp.StatusCode >= 400) {
                metrics.Errors++
            }

            return resp, err
        })
    }
}

func main() {
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })

    metrics := &Metrics{}

    // Create client with middleware stack
    client := helix.NewClient("your-client-id", authClient,
        helix.WithMiddleware(LoggingMiddleware),
        helix.WithMiddleware(RetryMiddleware(3)),
        helix.WithMiddleware(MetricsMiddleware(metrics)),
    )

    ctx := context.Background()

    // Make some requests
    for i := 0; i < 5; i++ {
        _, _ = client.GetUsers(ctx, &helix.GetUsersParams{
            Logins: []string{"twitchdev"},
        })
    }

    // Print metrics
    fmt.Printf("\n=== Metrics ===\n")
    fmt.Printf("Total Requests: %d\n", metrics.TotalRequests)
    fmt.Printf("Errors: %d\n", metrics.Errors)
    fmt.Printf("Average Latency: %v\n", metrics.TotalLatency/time.Duration(metrics.TotalRequests))
}
```

## Complete Example: Efficient Multi-Channel Dashboard

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

type ChannelData struct {
    User    *helix.User
    Channel *helix.ChannelInformation
    Stream  *helix.Stream
}

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    _, _ = authClient.GetAppAccessToken(ctx)

    // Setup client with caching
    cache := helix.NewMemoryCache(1 * time.Minute)
    client := helix.NewClient("your-client-id", authClient,
        helix.WithCache(cache),
    )

    // Channels to monitor
    channels := []string{"streamer1", "streamer2", "streamer3", "streamer4", "streamer5"}

    // Fetch all data concurrently
    channelData := make(map[string]*ChannelData)
    var mu sync.Mutex

    // Get users
    userResults, _ := helix.Batch(ctx, channels, func(ctx context.Context, login string) (*helix.GetUsersResponse, error) {
        return client.GetUsers(ctx, &helix.GetUsersParams{Logins: []string{login}})
    })

    for i, result := range userResults {
        if result.Error == nil && len(result.Value.Data) > 0 {
            mu.Lock()
            channelData[channels[i]] = &ChannelData{User: &result.Value.Data[0]}
            mu.Unlock()
        }
    }

    // Get user IDs for next requests
    var userIDs []string
    for _, login := range channels {
        if data, ok := channelData[login]; ok && data.User != nil {
            userIDs = append(userIDs, data.User.ID)
        }
    }

    // Get channel info and streams concurrently
    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        channelResults, _ := helix.Batch(ctx, userIDs, func(ctx context.Context, id string) (*helix.GetChannelInformationResponse, error) {
            return client.GetChannelInformation(ctx, &helix.GetChannelInformationParams{BroadcasterIDs: []string{id}})
        })
        for i, result := range channelResults {
            if result.Error == nil && len(result.Value.Data) > 0 {
                mu.Lock()
                for _, data := range channelData {
                    if data.User != nil && data.User.ID == userIDs[i] {
                        data.Channel = &result.Value.Data[0]
                        break
                    }
                }
                mu.Unlock()
            }
        }
    }()

    go func() {
        defer wg.Done()
        streams, _ := client.GetStreams(ctx, &helix.GetStreamsParams{UserIDs: userIDs})
        mu.Lock()
        for _, stream := range streams.Data {
            for _, data := range channelData {
                if data.User != nil && data.User.ID == stream.UserID {
                    s := stream // Copy
                    data.Stream = &s
                    break
                }
            }
        }
        mu.Unlock()
    }()

    wg.Wait()

    // Display dashboard
    fmt.Println("=== Channel Dashboard ===\n")
    for login, data := range channelData {
        if data.User == nil {
            fmt.Printf("❌ %s: Not found\n", login)
            continue
        }

        status := "⚫ Offline"
        viewers := 0
        if data.Stream != nil {
            status = "🔴 LIVE"
            viewers = data.Stream.ViewerCount
        }

        title := "N/A"
        game := "N/A"
        if data.Channel != nil {
            title = data.Channel.Title
            game = data.Channel.GameName
        }

        fmt.Printf("%s %s\n", status, data.User.DisplayName)
        fmt.Printf("   Title: %s\n", title)
        fmt.Printf("   Game: %s\n", game)
        if viewers > 0 {
            fmt.Printf("   Viewers: %d\n", viewers)
        }
        fmt.Println()
    }

    // Show rate limit status
    rateLimit := client.GetRateLimitInfo()
    fmt.Printf("Rate Limit: %d/%d remaining\n", rateLimit.Remaining, rateLimit.Limit)
}
```

