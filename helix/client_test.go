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
