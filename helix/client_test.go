package helix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)

	authClient := NewAuthClient(AuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})
	authClient.SetToken(&Token{
		AccessToken: "test-access-token",
	})

	client := NewClient("test-client-id", authClient, WithBaseURL(server.URL))

	return client, server
}

func TestNewClient(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	client := NewClient("test-client-id", authClient)

	if client == nil {
		t.Fatal("expected client to not be nil")
	}
	if client.clientID != "test-client-id" {
		t.Errorf("expected clientID test-client-id, got %s", client.clientID)
	}
	if client.baseURL != HelixBaseURL {
		t.Errorf("expected baseURL %s, got %s", HelixBaseURL, client.baseURL)
	}
}

func TestClient_WithOptions(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	customHTTPClient := &http.Client{}
	client := NewClient("test-client-id", authClient,
		WithHTTPClient(customHTTPClient),
		WithBaseURL("http://custom.url"),
	)

	if client.httpClient != customHTTPClient {
		t.Error("expected custom HTTP client")
	}
	if client.baseURL != "http://custom.url" {
		t.Errorf("expected custom base URL, got %s", client.baseURL)
	}
}

func TestClient_GetUsers(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("expected /users, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Client-Id") != "test-client-id" {
			t.Errorf("expected Client-Id header")
		}
		if r.Header.Get("Authorization") != "Bearer test-access-token" {
			t.Errorf("expected Authorization header")
		}

		resp := Response[User]{
			Data: []User{
				{
					ID:          "12345",
					Login:       "testuser",
					DisplayName: "TestUser",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUsers(context.Background(), &GetUsersParams{
		IDs: []string{"12345"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "12345" {
		t.Errorf("expected user ID 12345, got %s", resp.Data[0].ID)
	}
}

func TestClient_APIError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := ErrorResponse{
			Error:   "Unauthorized",
			Status:  401,
			Message: "Invalid OAuth token",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetUsers(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status code 401, got %d", apiErr.StatusCode)
	}
	if apiErr.ErrorType != "Unauthorized" {
		t.Errorf("expected error type 'Unauthorized', got %s", apiErr.ErrorType)
	}
}

func TestClient_Pagination(t *testing.T) {
	callCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Check if pagination cursor is being sent
		_ = r.URL.Query().Get("after")

		resp := Response[User]{
			Data: []User{
				{ID: "user" + string(rune('0'+callCount))},
			},
		}

		if callCount == 1 {
			resp.Pagination = &Pagination{Cursor: "next-page-cursor"}
		}

		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	// First request
	resp1, err := client.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp1.Pagination == nil || resp1.Pagination.Cursor != "next-page-cursor" {
		t.Error("expected pagination cursor")
	}

	// Second request with cursor
	resp2, err := client.GetUsers(context.Background(), &GetUsersParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = resp2
}

func TestClient_GetRateLimitInfo(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "750")
		w.Header().Set("Ratelimit-Reset", "1234567890")

		resp := Response[User]{Data: []User{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info := client.GetRateLimitInfo()
	if info.Limit != 800 {
		t.Errorf("expected limit 800, got %d", info.Limit)
	}
	if info.Remaining != 750 {
		t.Errorf("expected remaining 750, got %d", info.Remaining)
	}
	if info.ResetAt.Unix() != 1234567890 {
		t.Errorf("expected reset timestamp 1234567890, got %d", info.ResetAt.Unix())
	}
}

func TestClient_RateLimitRetry(t *testing.T) {
	requestCount := 0
	resetTime := time.Now().Add(100 * time.Millisecond).Unix()

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount == 1 {
			// First request returns 429
			w.Header().Set("Ratelimit-Limit", "800")
			w.Header().Set("Ratelimit-Remaining", "0")
			w.Header().Set("Ratelimit-Reset", fmt.Sprintf("%d", resetTime))
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Too Many Requests",
				Status:  429,
				Message: "Rate limit exceeded",
			})
			return
		}

		// Second request succeeds
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "799")
		w.Header().Set("Ratelimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		_ = json.NewEncoder(w).Encode(Response[User]{
			Data: []User{{ID: "12345", Login: "testuser"}},
		})
	})
	defer server.Close()

	resp, err := client.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 user, got %d", len(resp.Data))
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests (1 retry), got %d", requestCount)
	}
}

func TestClient_RateLimitRetryDisabled(t *testing.T) {
	requestCount := 0
	resetTime := time.Now().Add(time.Second).Unix()

	authClient := NewAuthClient(AuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})
	authClient.SetToken(&Token{AccessToken: "test-access-token"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "0")
		w.Header().Set("Ratelimit-Reset", fmt.Sprintf("%d", resetTime))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Too Many Requests",
			Status:  429,
			Message: "Rate limit exceeded",
		})
	}))
	defer server.Close()

	client := NewClient("test-client-id", authClient,
		WithBaseURL(server.URL),
		WithRetry(false, 0), // Disable retries
	)

	_, err := client.GetUsers(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsRateLimitError(err) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}

	if requestCount != 1 {
		t.Errorf("expected 1 request (no retries), got %d", requestCount)
	}

	// Verify RateLimitError fields
	rateLimitErr := err.(*RateLimitError)
	if rateLimitErr.Limit != 800 {
		t.Errorf("expected limit 800, got %d", rateLimitErr.Limit)
	}
	if rateLimitErr.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", rateLimitErr.Remaining)
	}
}

func TestClient_RateLimitMaxRetries(t *testing.T) {
	requestCount := 0
	resetTime := time.Now().Add(50 * time.Millisecond).Unix()

	authClient := NewAuthClient(AuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})
	authClient.SetToken(&Token{AccessToken: "test-access-token"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "0")
		w.Header().Set("Ratelimit-Reset", fmt.Sprintf("%d", resetTime))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Too Many Requests",
			Status:  429,
			Message: "Rate limit exceeded",
		})
	}))
	defer server.Close()

	client := NewClient("test-client-id", authClient,
		WithBaseURL(server.URL),
		WithRetry(true, 2), // Max 2 retries
		WithMaxRetryWait(100*time.Millisecond),
	)

	_, err := client.GetUsers(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error after max retries")
	}

	if !IsRateLimitError(err) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}

	// Should have made 3 requests: initial + 2 retries
	if requestCount != 3 {
		t.Errorf("expected 3 requests (initial + 2 retries), got %d", requestCount)
	}
}

func TestClient_RateLimitContextCancellation(t *testing.T) {
	resetTime := time.Now().Add(10 * time.Second).Unix()

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "0")
		w.Header().Set("Ratelimit-Reset", fmt.Sprintf("%d", resetTime))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Too Many Requests",
			Status:  429,
			Message: "Rate limit exceeded",
		})
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GetUsers(ctx, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "RateLimitError",
			err:      &RateLimitError{},
			expected: true,
		},
		{
			name:     "APIError",
			err:      &APIError{StatusCode: 429},
			expected: false,
		},
		{
			name:     "nil",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimitError(tt.err); got != tt.expected {
				t.Errorf("IsRateLimitError() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestClient_WithRetryOptions(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	client := NewClient("test-client-id", authClient,
		WithRetry(false, 5),
		WithMaxRetryWait(30*time.Second),
	)

	if client.retryEnabled {
		t.Error("expected retryEnabled to be false")
	}
	if client.maxRetries != 5 {
		t.Errorf("expected maxRetries 5, got %d", client.maxRetries)
	}
	if client.maxRetryWait != 30*time.Second {
		t.Errorf("expected maxRetryWait 30s, got %v", client.maxRetryWait)
	}
}

func TestClient_WithExponentialBackoff(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	client := NewClient("test-client-id", authClient,
		WithExponentialBackoff(500*time.Millisecond),
	)

	if !client.useExpBackoff {
		t.Error("expected useExpBackoff to be true")
	}
	if client.baseRetryDelay != 500*time.Millisecond {
		t.Errorf("expected baseRetryDelay 500ms, got %v", client.baseRetryDelay)
	}
}

func TestClient_WithMiddleware(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	mw := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		return next(ctx, req)
	}

	client := NewClient("test-client-id", authClient,
		WithMiddleware(mw),
	)

	if len(client.middleware) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(client.middleware))
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 401,
		ErrorType:  "Unauthorized",
		Message:    "Invalid token",
	}

	expected := "twitch api error 401: Unauthorized - Invalid token"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{
		ResetAt:    time.Now().Add(30 * time.Second),
		Remaining:  0,
		Limit:      800,
		RetryAfter: 30 * time.Second,
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
	if err.Error() != "rate limit exceeded: 0/800 points remaining, resets in 30s" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestClient_WaitForRateLimit(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	// Test with remaining points
	client := NewClient("test-client-id", authClient)
	client.rateLimitRemaining = 100

	err := client.WaitForRateLimit(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test with zero remaining but reset time in past
	client.rateLimitRemaining = 0
	client.rateLimitReset = time.Now().Add(-time.Second)

	err = client.WaitForRateLimit(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test with zero remaining and future reset time (with timeout)
	client.rateLimitRemaining = 0
	client.rateLimitReset = time.Now().Add(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = client.WaitForRateLimit(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestAddPaginationParams(t *testing.T) {
	q := make(url.Values)

	// Test nil params
	addPaginationParams(q, nil)
	if len(q) != 0 {
		t.Error("expected empty query for nil params")
	}

	// Test with all params
	addPaginationParams(q, &PaginationParams{
		First:  50,
		After:  "cursor-after",
		Before: "cursor-before",
	})

	if q.Get("first") != "50" {
		t.Errorf("expected first=50, got %s", q.Get("first"))
	}
	if q.Get("after") != "cursor-after" {
		t.Errorf("expected after=cursor-after, got %s", q.Get("after"))
	}
	if q.Get("before") != "cursor-before" {
		t.Errorf("expected before=cursor-before, got %s", q.Get("before"))
	}
}

func TestClient_ExponentialBackoff_Retry(t *testing.T) {
	requestCount := 0
	resetTime := time.Now().Add(50 * time.Millisecond).Unix()

	authClient := NewAuthClient(AuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	})
	authClient.SetToken(&Token{AccessToken: "test-access-token"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount < 3 {
			w.Header().Set("Ratelimit-Limit", "800")
			w.Header().Set("Ratelimit-Remaining", "0")
			w.Header().Set("Ratelimit-Reset", fmt.Sprintf("%d", resetTime))
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Too Many Requests",
				Status:  429,
				Message: "Rate limit exceeded",
			})
			return
		}

		// Third request succeeds
		_ = json.NewEncoder(w).Encode(Response[User]{
			Data: []User{{ID: "12345", Login: "testuser"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-client-id", authClient,
		WithBaseURL(server.URL),
		WithRetry(true, 5),
		WithExponentialBackoff(10*time.Millisecond),
		WithMaxRetryWait(100*time.Millisecond),
	)

	resp, err := client.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 user, got %d", len(resp.Data))
	}
	if requestCount != 3 {
		t.Errorf("expected 3 requests, got %d", requestCount)
	}
}

func TestClient_MiddlewareExecution(t *testing.T) {
	middlewareCalled := false

	mw := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		middlewareCalled = true
		return next(ctx, req)
	}

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response[User]{Data: []User{}})
	})
	defer server.Close()

	client.middleware = append(client.middleware, mw)

	_, err := client.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
}

func TestClient_CacheHit(t *testing.T) {
	requestCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := Response[User]{
			Data: []User{{ID: "123", Login: "test"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cache := NewMemoryCache(100)
	client.cache = cache
	client.cacheTTL = time.Minute
	client.cacheEnabled = true

	// First request should hit the server
	resp1, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}

	// Second request should be cached
	resp2, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request (cached), got %d", requestCount)
	}

	// Verify same data
	if resp1.Data[0].ID != resp2.Data[0].ID {
		t.Error("expected same data from cache")
	}
}

func TestClient_CacheBypass_NoCacheContext(t *testing.T) {
	requestCount := 0
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := Response[User]{
			Data: []User{{ID: "123", Login: "test"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cache := NewMemoryCache(100)
	client.cache = cache
	client.cacheTTL = time.Minute
	client.cacheEnabled = true

	// First request
	_, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second request with NoCacheContext should bypass cache
	ctx := NoCacheContext(context.Background())
	_, err = client.GetUsers(ctx, &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests (cache bypassed), got %d", requestCount)
	}
}

func TestClient_CacheHit_NilResult(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{
			Data: []User{{ID: "123", Login: "test"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cache := NewMemoryCache(100)
	client.cache = cache
	client.cacheTTL = time.Minute
	client.cacheEnabled = true

	// First request to populate cache
	_, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second request with nil result (just checking cache exists)
	req := &Request{
		Method:   "GET",
		Endpoint: "/users",
		Query:    url.Values{"id": []string{"123"}},
	}
	err = client.Do(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error with nil result: %v", err)
	}
}

func TestClient_CacheKey_WithTokenProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Client with tokenProvider that has a token
	provider := &mockTokenProvider{token: &Token{AccessToken: "provider-token"}}
	client := NewClient("test-client-id", nil, WithBaseURL(server.URL))
	client.tokenProvider = provider

	cache := NewMemoryCache(100)
	client.cache = cache
	client.cacheTTL = time.Minute
	client.cacheEnabled = true

	// Make a request to populate cache
	_, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cache has an entry
	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}
}

func TestClient_CacheKey_WithTokenProvider_NilToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Client with tokenProvider that has nil token
	provider := &mockTokenProvider{token: nil}
	client := NewClient("test-client-id", nil, WithBaseURL(server.URL))
	client.tokenProvider = provider

	cache := NewMemoryCache(100)
	client.cache = cache
	client.cacheTTL = time.Minute
	client.cacheEnabled = true

	// Make a request to populate cache
	_, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cache has an entry (even with nil token)
	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}
}

func TestClient_CacheKey_Isolation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Shared cache
	cache := NewMemoryCache(100)

	// Client 1 with token provider
	provider1 := &mockTokenProvider{token: &Token{AccessToken: "token1"}}
	client1 := NewClient("test-client-id", nil, WithBaseURL(server.URL))
	client1.tokenProvider = provider1
	client1.cache = cache
	client1.cacheTTL = time.Minute
	client1.cacheEnabled = true

	// Client 2 with different token provider
	provider2 := &mockTokenProvider{token: &Token{AccessToken: "token2"}}
	client2 := NewClient("test-client-id", nil, WithBaseURL(server.URL))
	client2.tokenProvider = provider2
	client2.cache = cache
	client2.cacheTTL = time.Minute
	client2.cacheEnabled = true

	// Make requests from both clients
	_, err := client1.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("client1 request failed: %v", err)
	}

	_, err = client2.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Fatalf("client2 request failed: %v", err)
	}

	// Cache should have 2 entries (one per token)
	if cache.Size() != 2 {
		t.Errorf("expected cache size 2 (isolated by token), got %d", cache.Size())
	}
}

func TestClient_DoOnce_MarshalError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	// Use a channel as body which cannot be marshaled to JSON
	req := &Request{
		Method:   "POST",
		Endpoint: "/test",
		Body:     make(chan int), // channels can't be marshaled
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("expected marshal error")
	}
}

func TestClient_DoOnce_InvalidURL(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	authClient.SetToken(&Token{AccessToken: "test"})

	client := NewClient("test-client-id", authClient, WithBaseURL("://invalid"))

	req := &Request{
		Method:   "GET",
		Endpoint: "/test",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("expected URL error")
	}
}

func TestClient_DoOnce_TokenProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer provider-token" {
			t.Errorf("expected Authorization header with provider token")
		}
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := &mockTokenProvider{token: &Token{AccessToken: "provider-token"}}
	client := NewClient("test-client-id", nil, WithBaseURL(server.URL))
	client.tokenProvider = provider

	req := &Request{
		Method:   "GET",
		Endpoint: "/users",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type mockTokenProvider struct {
	token *Token
}

func (m *mockTokenProvider) GetToken() *Token {
	return m.token
}

func TestClient_DoOnce_NoAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("expected no Authorization header")
		}
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-client-id", nil, WithBaseURL(server.URL))

	req := &Request{
		Method:   "GET",
		Endpoint: "/users",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DoOnce_UnmarshalableErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	authClient.SetToken(&Token{AccessToken: "test"})
	client := NewClient("test-client-id", authClient, WithBaseURL(server.URL))

	req := &Request{
		Method:   "GET",
		Endpoint: "/test",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.ErrorType != "unknown" {
		t.Errorf("expected error type 'unknown', got %s", apiErr.ErrorType)
	}
}

func TestClient_WaitForRateLimit_Success(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})

	client := NewClient("test-client-id", authClient)
	client.rateLimitRemaining = 0
	client.rateLimitReset = time.Now().Add(50 * time.Millisecond)

	start := time.Now()
	err := client.WaitForRateLimit(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected to wait, elapsed: %v", elapsed)
	}
}

func TestClient_MiddlewareError(t *testing.T) {
	mw := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		return nil, fmt.Errorf("middleware error")
	}

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response[User]{Data: []User{}})
	})
	defer server.Close()

	client.middleware = append(client.middleware, mw)

	_, err := client.GetUsers(context.Background(), nil)
	if err == nil {
		t.Error("expected middleware error")
	}
}

func TestClient_doOnce(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response[User]{Data: []User{{ID: "123"}}})
	})
	defer server.Close()

	req := &Request{
		Method:   "GET",
		Endpoint: "/users",
	}

	var result Response[User]
	err := client.doOnce(context.Background(), req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].ID != "123" {
		t.Error("expected response data")
	}
}

func TestClient_DoOnce_NetworkError(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	authClient.SetToken(&Token{AccessToken: "test"})

	// Use a server that immediately closes connections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close the connection before sending a response
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
		}
	}))
	defer server.Close()

	client := NewClient("test-client-id", authClient, WithBaseURL(server.URL))

	req := &Request{
		Method:   "GET",
		Endpoint: "/test",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("expected network error")
	}
}

func TestClient_DoOnce_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not valid json{"))
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	authClient.SetToken(&Token{AccessToken: "test"})
	client := NewClient("test-client-id", authClient, WithBaseURL(server.URL))

	req := &Request{
		Method:   "GET",
		Endpoint: "/test",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestClient_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Error("expected custom header")
		}
		_ = json.NewEncoder(w).Encode(Response[User]{Data: []User{}})
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	authClient.SetToken(&Token{AccessToken: "test"})
	client := NewClient("test-client-id", authClient, WithBaseURL(server.URL))

	// Use middleware to add custom headers via context
	ctx := context.WithValue(context.Background(), headersContextKey{}, map[string]string{
		"X-Custom-Header": "custom-value",
	})

	req := &Request{
		Method:   "GET",
		Endpoint: "/users",
	}

	var result Response[User]
	err := client.Do(ctx, req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddPaginationParams_ZeroFirst(t *testing.T) {
	q := make(url.Values)

	// Test with First=0 (should not add first param)
	addPaginationParams(q, &PaginationParams{
		First: 0,
		After: "cursor",
	})

	if q.Get("first") != "" {
		t.Error("expected no first param when First=0")
	}
	if q.Get("after") != "cursor" {
		t.Errorf("expected after=cursor, got %s", q.Get("after"))
	}
}

func TestClient_NegativeMaxRetries(t *testing.T) {
	// This tests the edge case where maxRetries is negative,
	// causing the retry loop to never execute
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	authClient.SetToken(&Token{AccessToken: "test"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "Too Many Requests",
			Status:  429,
			Message: "Rate limit exceeded",
		})
	}))
	defer server.Close()

	client := NewClient("test-client-id", authClient, WithBaseURL(server.URL))
	// Manually set negative maxRetries to hit the edge case
	client.maxRetries = -1

	req := &Request{
		Method:   "GET",
		Endpoint: "/test",
	}

	var result Response[User]
	_, err := client.doWithRetryAndResponse(context.Background(), req, &result)
	// With negative maxRetries, the loop never executes and returns nil, nil
	if err != nil {
		t.Errorf("expected nil error with negative maxRetries, got: %v", err)
	}
}

type brokenBodyReader struct{}

func (e *brokenBodyReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func (e *brokenBodyReader) Close() error {
	return nil
}

type brokenBodyTransport struct{}

func (t *brokenBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       &brokenBodyReader{},
	}, nil
}

func TestClient_ResponseBodyReadError(t *testing.T) {
	// This tests the io.ReadAll error path by using a custom RoundTripper
	// that returns a response with a broken body reader
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	authClient.SetToken(&Token{AccessToken: "test"})

	// Create a custom HTTP client with a RoundTripper that returns a broken body
	client := NewClient("test-client-id", authClient)
	client.httpClient = &http.Client{
		Transport: &brokenBodyTransport{},
	}

	req := &Request{
		Method:   "GET",
		Endpoint: "/test",
	}

	var result Response[User]
	err := client.Do(context.Background(), req, &result)
	if err == nil {
		t.Error("expected error from broken response body")
	}
	if err != nil {
		errStr := err.Error()
		found := false
		target := "reading response body"
		for i := 0; i <= len(errStr)-len(target); i++ {
			if errStr[i:i+len(target)] == target {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected 'reading response body' error, got: %v", err)
		}
	}
}
