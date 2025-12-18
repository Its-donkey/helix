package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestClient_Use(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	client := NewClient("test-client-id", authClient)

	if len(client.middleware) != 0 {
		t.Errorf("expected 0 middleware, got %d", len(client.middleware))
	}

	mw1 := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		return next(ctx, req)
	}
	mw2 := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		return next(ctx, req)
	}

	client.Use(mw1)
	if len(client.middleware) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(client.middleware))
	}

	client.Use(mw2)
	if len(client.middleware) != 2 {
		t.Errorf("expected 2 middleware, got %d", len(client.middleware))
	}
}

func TestClient_Use_Multiple(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	client := NewClient("test-client-id", authClient)

	mw1 := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		return next(ctx, req)
	}
	mw2 := func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		return next(ctx, req)
	}

	client.Use(mw1, mw2)
	if len(client.middleware) != 2 {
		t.Errorf("expected 2 middleware, got %d", len(client.middleware))
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var logs []string
	var mu sync.Mutex
	logger := func(format string, args ...interface{}) {
		mu.Lock()
		logs = append(logs, format)
		mu.Unlock()
	}

	mw := LoggingMiddleware(logger)

	// Test successful request
	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if len(logs) != 2 {
		t.Errorf("expected 2 log entries, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "->") {
		t.Error("expected request log to contain '->'")
	}
	if !strings.Contains(logs[1], "<-") {
		t.Error("expected response log to contain '<-'")
	}
}

func TestLoggingMiddleware_Error(t *testing.T) {
	var logs []string
	var mu sync.Mutex
	logger := func(format string, args ...interface{}) {
		mu.Lock()
		logs = append(logs, format)
		mu.Unlock()
	}

	mw := LoggingMiddleware(logger)

	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		return nil, context.Canceled
	}

	_, err := mw(context.Background(), req, next)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("expected 2 log entries, got %d", len(logs))
	}
	if !strings.Contains(logs[1], "error") {
		t.Error("expected error log entry")
	}
}

func TestRetryMiddleware_Success(t *testing.T) {
	mw := RetryMiddleware(3)

	callCount := 0
	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		callCount++
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestRetryMiddleware_RetryOnError(t *testing.T) {
	mw := RetryMiddleware(3)

	callCount := 0
	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		callCount++
		if callCount < 3 {
			return nil, context.DeadlineExceeded
		}
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryMiddleware_RetryOnStatus(t *testing.T) {
	mw := RetryMiddleware(3)

	callCount := 0
	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		callCount++
		if callCount < 3 {
			return &MiddlewareResponse{StatusCode: 503}, nil // Service Unavailable
		}
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryMiddleware_ExhaustedRetries_Error(t *testing.T) {
	mw := RetryMiddleware(2)

	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		return nil, context.DeadlineExceeded
	}

	_, err := mw(context.Background(), req, next)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestRetryMiddleware_ExhaustedRetries_Status(t *testing.T) {
	mw := RetryMiddleware(2)

	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		return &MiddlewareResponse{StatusCode: 503}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}
}

func TestRetryMiddleware_CustomStatuses(t *testing.T) {
	mw := RetryMiddleware(2, 500, 502)

	callCount := 0
	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		callCount++
		if callCount == 1 {
			return &MiddlewareResponse{StatusCode: 500}, nil // Should retry
		}
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestHeaderMiddleware(t *testing.T) {
	headers := map[string]string{
		"X-Custom-Header": "custom-value",
	}
	mw := HeaderMiddleware(headers)

	var capturedCtx context.Context
	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		capturedCtx = ctx
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	_, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify headers are in context
	ctxHeaders := headersFromContext(capturedCtx)
	if ctxHeaders == nil {
		t.Fatal("expected headers in context")
	}
	if ctxHeaders["X-Custom-Header"] != "custom-value" {
		t.Errorf("expected custom header value, got %s", ctxHeaders["X-Custom-Header"])
	}
}

func TestContextWithHeaders(t *testing.T) {
	ctx := context.Background()

	// Add first set of headers
	ctx = contextWithHeaders(ctx, map[string]string{"Header1": "Value1"})
	headers := headersFromContext(ctx)
	if headers["Header1"] != "Value1" {
		t.Errorf("expected Header1=Value1, got %s", headers["Header1"])
	}

	// Add second set, should merge
	ctx = contextWithHeaders(ctx, map[string]string{"Header2": "Value2"})
	headers = headersFromContext(ctx)
	if headers["Header1"] != "Value1" {
		t.Errorf("expected Header1=Value1 after merge, got %s", headers["Header1"])
	}
	if headers["Header2"] != "Value2" {
		t.Errorf("expected Header2=Value2 after merge, got %s", headers["Header2"])
	}

	// Overwrite existing header
	ctx = contextWithHeaders(ctx, map[string]string{"Header1": "NewValue"})
	headers = headersFromContext(ctx)
	if headers["Header1"] != "NewValue" {
		t.Errorf("expected Header1=NewValue after overwrite, got %s", headers["Header1"])
	}
}

func TestHeadersFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	headers := headersFromContext(ctx)
	if headers != nil {
		t.Errorf("expected nil for empty context, got %v", headers)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	var capturedMetrics RequestMetrics
	collector := func(metrics RequestMetrics) {
		capturedMetrics = metrics
	}

	mw := MetricsMiddleware(collector)

	req := &Request{Method: "GET", Endpoint: "/users"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		return &MiddlewareResponse{StatusCode: 200}, nil
	}

	resp, err := mw(context.Background(), req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if capturedMetrics.Method != "GET" {
		t.Errorf("expected method GET, got %s", capturedMetrics.Method)
	}
	if capturedMetrics.Endpoint != "/users" {
		t.Errorf("expected endpoint /users, got %s", capturedMetrics.Endpoint)
	}
	if capturedMetrics.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", capturedMetrics.StatusCode)
	}
	if capturedMetrics.Duration < 0 {
		t.Errorf("expected non-negative duration, got %d", capturedMetrics.Duration)
	}
	if capturedMetrics.Error != nil {
		t.Errorf("expected no error, got %v", capturedMetrics.Error)
	}
}

func TestMetricsMiddleware_Error(t *testing.T) {
	var capturedMetrics RequestMetrics
	collector := func(metrics RequestMetrics) {
		capturedMetrics = metrics
	}

	mw := MetricsMiddleware(collector)

	req := &Request{Method: "POST", Endpoint: "/channels"}
	next := func(ctx context.Context, r *Request) (*MiddlewareResponse, error) {
		return nil, context.Canceled
	}

	_, err := mw(context.Background(), req, next)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	if capturedMetrics.Method != "POST" {
		t.Errorf("expected method POST, got %s", capturedMetrics.Method)
	}
	if capturedMetrics.Endpoint != "/channels" {
		t.Errorf("expected endpoint /channels, got %s", capturedMetrics.Endpoint)
	}
	if capturedMetrics.StatusCode != 0 {
		t.Errorf("expected status code 0 on error, got %d", capturedMetrics.StatusCode)
	}
	if capturedMetrics.Error != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", capturedMetrics.Error)
	}
}

func TestMiddleware_Integration(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		// Check for custom header
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Error("expected custom header")
		}
		resp := Response[User]{
			Data: []User{{ID: "123", Login: "test"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	var logCount int
	var metricsCollected bool
	var mu sync.Mutex

	client.Use(
		LoggingMiddleware(func(format string, args ...interface{}) {
			mu.Lock()
			logCount++
			mu.Unlock()
		}),
		HeaderMiddleware(map[string]string{"X-Test-Header": "test-value"}),
		MetricsMiddleware(func(m RequestMetrics) {
			mu.Lock()
			metricsCollected = true
			mu.Unlock()
		}),
	)

	_, err := client.GetUsers(context.Background(), &GetUsersParams{IDs: []string{"123"}})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mu.Lock()
	if logCount != 2 {
		t.Errorf("expected 2 log calls, got %d", logCount)
	}
	if !metricsCollected {
		t.Error("expected metrics to be collected")
	}
	mu.Unlock()
}
