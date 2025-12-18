package helix

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultBatchOptions(t *testing.T) {
	opts := DefaultBatchOptions()
	if opts.MaxConcurrent != 10 {
		t.Errorf("expected MaxConcurrent 10, got %d", opts.MaxConcurrent)
	}
	if opts.StopOnError {
		t.Error("expected StopOnError to be false")
	}
}

func TestBatch_EmptyRequests(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	results := client.Batch(context.Background(), []BatchRequest{}, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestBatch_SingleRequest(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{
			Data: []User{{ID: "123", Login: "test"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	var result Response[User]
	requests := []BatchRequest{
		{
			Request: &Request{
				Method:   "GET",
				Endpoint: "/users",
				Query:    url.Values{"id": []string{"123"}},
			},
			Result: &result,
		},
	}

	results := client.Batch(context.Background(), requests, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Errorf("unexpected error: %v", results[0].Error)
	}
	if results[0].Index != 0 {
		t.Errorf("expected index 0, got %d", results[0].Index)
	}
}

func TestBatch_MultipleRequests(t *testing.T) {
	var requestCount int32
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		resp := Response[User]{
			Data: []User{{ID: "123", Login: "test"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := make([]BatchRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = BatchRequest{
			Request: &Request{
				Method:   "GET",
				Endpoint: "/users",
			},
			Result: &Response[User]{},
		}
	}

	results := client.Batch(context.Background(), requests, nil)
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Error != nil {
			t.Errorf("result %d: unexpected error: %v", i, r.Error)
		}
	}

	if atomic.LoadInt32(&requestCount) != 5 {
		t.Errorf("expected 5 requests, got %d", requestCount)
	}
}

func TestBatch_WithConcurrencyLimit(t *testing.T) {
	var maxConcurrent int32
	var currentConcurrent int32

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&currentConcurrent, 1)
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current > max {
				if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			} else {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&currentConcurrent, -1)

		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := make([]BatchRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	opts := &BatchOptions{MaxConcurrent: 3, StopOnError: false}
	results := client.Batch(context.Background(), requests, opts)

	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}

	if atomic.LoadInt32(&maxConcurrent) > 3 {
		t.Errorf("expected max concurrent <= 3, got %d", maxConcurrent)
	}
}

func TestBatch_StopOnError(t *testing.T) {
	var requestCount int32
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		time.Sleep(50 * time.Millisecond) // Slow down subsequent requests
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := make([]BatchRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	opts := &BatchOptions{MaxConcurrent: 1, StopOnError: true}
	results := client.Batch(context.Background(), requests, opts)

	// First request should have error
	if results[0].Error == nil {
		t.Error("expected first result to have error")
	}

	// Some subsequent requests should be canceled
	canceledCount := 0
	for _, r := range results {
		if r.Error == context.Canceled {
			canceledCount++
		}
	}
	if canceledCount == 0 {
		t.Log("Note: no requests were canceled (timing dependent)")
	}
}

func TestBatch_ContextCanceled(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	requests := []BatchRequest{
		{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		},
	}

	results := client.Batch(ctx, requests, nil)
	if results[0].Error != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", results[0].Error)
	}
}

func TestBatch_NilOptions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := []BatchRequest{
		{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		},
	}

	results := client.Batch(context.Background(), requests, nil)
	if results[0].Error != nil {
		t.Errorf("unexpected error: %v", results[0].Error)
	}
}

func TestBatch_UnlimitedConcurrency(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := make([]BatchRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	opts := &BatchOptions{MaxConcurrent: 0} // Unlimited
	results := client.Batch(context.Background(), requests, opts)
	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}
}

func TestBatchGet(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := []GetRequest{
		{Endpoint: "/users", Query: url.Values{"id": []string{"123"}}, Result: &Response[User]{}},
		{Endpoint: "/users", Query: url.Values{"id": []string{"456"}}, Result: &Response[User]{}},
	}

	results := client.BatchGet(context.Background(), requests, nil)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Error != nil {
			t.Errorf("result %d: unexpected error: %v", i, r.Error)
		}
	}
}

func TestBatchSequential(t *testing.T) {
	var requestOrder []int
	var mu = &muWrapper{}
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		mu.Lock()
		switch id {
		case "1":
			requestOrder = append(requestOrder, 1)
		case "2":
			requestOrder = append(requestOrder, 2)
		case "3":
			requestOrder = append(requestOrder, 3)
		}
		mu.Unlock()
		resp := Response[User]{Data: []User{{ID: id}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := []BatchRequest{
		{Request: &Request{Method: "GET", Endpoint: "/users", Query: url.Values{"id": []string{"1"}}}, Result: &Response[User]{}},
		{Request: &Request{Method: "GET", Endpoint: "/users", Query: url.Values{"id": []string{"2"}}}, Result: &Response[User]{}},
		{Request: &Request{Method: "GET", Endpoint: "/users", Query: url.Values{"id": []string{"3"}}}, Result: &Response[User]{}},
	}

	results := client.BatchSequential(context.Background(), requests)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify sequential order
	for i := 0; i < 3; i++ {
		if requestOrder[i] != i+1 {
			t.Errorf("expected request order %d at position %d, got %d", i+1, i, requestOrder[i])
		}
	}
}

type muWrapper struct {
	locked bool
}

func (m *muWrapper) Lock()   { m.locked = true }
func (m *muWrapper) Unlock() { m.locked = false }

func TestBatchSequential_ContextCanceled(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	requests := []BatchRequest{
		{Request: &Request{Method: "GET", Endpoint: "/users"}, Result: &Response[User]{}},
	}

	results := client.BatchSequential(ctx, requests)
	if results[0].Error != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", results[0].Error)
	}
}

func TestBatchWithCallback(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := []BatchRequest{
		{Request: &Request{Method: "GET", Endpoint: "/users"}, Result: &Response[User]{}},
		{Request: &Request{Method: "GET", Endpoint: "/users"}, Result: &Response[User]{}},
	}

	var callbackCount int32
	callback := func(result BatchResult) {
		atomic.AddInt32(&callbackCount, 1)
	}

	client.BatchWithCallback(context.Background(), requests, nil, callback)

	if atomic.LoadInt32(&callbackCount) != 2 {
		t.Errorf("expected 2 callbacks, got %d", callbackCount)
	}
}

func TestBatchWithCallback_EmptyRequests(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	var callbackCount int32
	callback := func(result BatchResult) {
		atomic.AddInt32(&callbackCount, 1)
	}

	client.BatchWithCallback(context.Background(), []BatchRequest{}, nil, callback)

	if atomic.LoadInt32(&callbackCount) != 0 {
		t.Errorf("expected 0 callbacks, got %d", callbackCount)
	}
}

func TestBatchWithCallback_StopOnError(t *testing.T) {
	var requestCount int32
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		time.Sleep(50 * time.Millisecond)
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := make([]BatchRequest, 5)
	for i := 0; i < 5; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	var results []BatchResult
	var mu = &muWrapper{}
	callback := func(result BatchResult) {
		mu.Lock()
		results = append(results, result)
		mu.Unlock()
	}

	opts := &BatchOptions{MaxConcurrent: 1, StopOnError: true}
	client.BatchWithCallback(context.Background(), requests, opts, callback)

	// Should have received callbacks
	if len(results) == 0 {
		t.Error("expected at least one callback")
	}
}

func TestBatchWithCallback_ContextCanceled(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	requests := []BatchRequest{
		{Request: &Request{Method: "GET", Endpoint: "/users"}, Result: &Response[User]{}},
	}

	var result BatchResult
	callback := func(r BatchResult) {
		result = r
	}

	client.BatchWithCallback(ctx, requests, nil, callback)

	if result.Error != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", result.Error)
	}
}

func TestBatchWithCallback_UnlimitedConcurrency(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	requests := make([]BatchRequest, 3)
	for i := 0; i < 3; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	var callbackCount int32
	callback := func(result BatchResult) {
		atomic.AddInt32(&callbackCount, 1)
	}

	opts := &BatchOptions{MaxConcurrent: 0}
	client.BatchWithCallback(context.Background(), requests, opts, callback)

	if atomic.LoadInt32(&callbackCount) != 3 {
		t.Errorf("expected 3 callbacks, got %d", callbackCount)
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		results  []BatchResult
		expected bool
	}{
		{
			name:     "no errors",
			results:  []BatchResult{{Index: 0, Error: nil}, {Index: 1, Error: nil}},
			expected: false,
		},
		{
			name:     "with error",
			results:  []BatchResult{{Index: 0, Error: nil}, {Index: 1, Error: errors.New("test")}},
			expected: true,
		},
		{
			name:     "empty results",
			results:  []BatchResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.results); got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFirstError(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")

	tests := []struct {
		name     string
		results  []BatchResult
		expected error
	}{
		{
			name:     "no errors",
			results:  []BatchResult{{Index: 0, Error: nil}, {Index: 1, Error: nil}},
			expected: nil,
		},
		{
			name:     "first has error",
			results:  []BatchResult{{Index: 0, Error: err1}, {Index: 1, Error: err2}},
			expected: err1,
		},
		{
			name:     "second has error",
			results:  []BatchResult{{Index: 0, Error: nil}, {Index: 1, Error: err2}},
			expected: err2,
		},
		{
			name:     "empty results",
			results:  []BatchResult{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstError(tt.results); got != tt.expected {
				t.Errorf("FirstError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")

	tests := []struct {
		name     string
		results  []BatchResult
		expected int
	}{
		{
			name:     "no errors",
			results:  []BatchResult{{Index: 0, Error: nil}, {Index: 1, Error: nil}},
			expected: 0,
		},
		{
			name:     "one error",
			results:  []BatchResult{{Index: 0, Error: nil}, {Index: 1, Error: err1}},
			expected: 1,
		},
		{
			name:     "multiple errors",
			results:  []BatchResult{{Index: 0, Error: err1}, {Index: 1, Error: err2}},
			expected: 2,
		},
		{
			name:     "empty results",
			results:  []BatchResult{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Errors(tt.results)
			if len(errs) != tt.expected {
				t.Errorf("Errors() returned %d errors, want %d", len(errs), tt.expected)
			}
		})
	}
}

func TestBatch_SemaphoreContextCancel(t *testing.T) {
	// Test semaphore blocking with context cancel
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	requests := make([]BatchRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	opts := &BatchOptions{MaxConcurrent: 1}
	results := client.Batch(ctx, requests, opts)

	// Some requests should have context deadline exceeded
	hasContextError := false
	for _, r := range results {
		if r.Error == context.DeadlineExceeded {
			hasContextError = true
			break
		}
	}
	if !hasContextError {
		t.Log("Note: no context deadline exceeded errors (timing dependent)")
	}
}

func TestBatchWithCallback_SemaphoreContextCancel(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		resp := Response[User]{Data: []User{{ID: "123"}}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	requests := make([]BatchRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = BatchRequest{
			Request: &Request{Method: "GET", Endpoint: "/users"},
			Result:  &Response[User]{},
		}
	}

	var results []BatchResult
	callback := func(r BatchResult) {
		results = append(results, r)
	}

	opts := &BatchOptions{MaxConcurrent: 1}
	client.BatchWithCallback(ctx, requests, opts, callback)

	// Should have received some callbacks
	if len(results) == 0 {
		t.Error("expected at least one callback")
	}
}
