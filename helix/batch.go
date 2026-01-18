package helix

import (
	"context"
	"net/url"
	"sync"
)

// BatchRequest represents a single request in a batch.
type BatchRequest struct {
	Request *Request
	Result  interface{}
}

// BatchResult contains the result of a batch request.
type BatchResult struct {
	Index int
	Error error
}

// BatchOptions configures batch execution behavior.
type BatchOptions struct {
	// MaxConcurrent limits concurrent requests (0 = unlimited)
	MaxConcurrent int
	// StopOnError stops processing remaining requests on first error
	StopOnError bool
}

// DefaultBatchOptions returns default batch options.
func DefaultBatchOptions() BatchOptions {
	return BatchOptions{
		MaxConcurrent: 10,
		StopOnError:   false,
	}
}

// Batch executes multiple requests concurrently with configurable parallelism.
// Results are returned in the same order as the input requests.
func (c *Client) Batch(ctx context.Context, requests []BatchRequest, opts *BatchOptions) []BatchResult {
	if opts == nil {
		defaultOpts := DefaultBatchOptions()
		opts = &defaultOpts
	}

	results := make([]BatchResult, len(requests))

	if len(requests) == 0 {
		return results
	}

	// Create semaphore for concurrency control
	var sem chan struct{}
	if opts.MaxConcurrent > 0 {
		sem = make(chan struct{}, opts.MaxConcurrent)
	}

	var wg sync.WaitGroup
	var stopMu sync.Mutex
	var resultsMu sync.Mutex
	stopped := false

	for i, req := range requests {
		// Check if we should stop
		if opts.StopOnError {
			stopMu.Lock()
			if stopped {
				stopMu.Unlock()
				resultsMu.Lock()
				results[i] = BatchResult{Index: i, Error: context.Canceled}
				resultsMu.Unlock()
				continue
			}
			stopMu.Unlock()
		}

		// Check context
		if ctx.Err() != nil {
			resultsMu.Lock()
			results[i] = BatchResult{Index: i, Error: ctx.Err()}
			resultsMu.Unlock()
			continue
		}

		// Acquire semaphore
		if sem != nil {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				resultsMu.Lock()
				results[i] = BatchResult{Index: i, Error: ctx.Err()}
				resultsMu.Unlock()
				continue
			}
		}

		wg.Add(1)
		go func(idx int, batchReq BatchRequest) {
			defer wg.Done()
			if sem != nil {
				defer func() { <-sem }()
			}

			err := c.Do(ctx, batchReq.Request, batchReq.Result)

			resultsMu.Lock()
			results[idx] = BatchResult{Index: idx, Error: err}
			resultsMu.Unlock()

			if err != nil && opts.StopOnError {
				stopMu.Lock()
				stopped = true
				stopMu.Unlock()
			}
		}(i, req)
	}

	wg.Wait()
	return results
}

// BatchGet executes multiple GET requests concurrently.
func (c *Client) BatchGet(ctx context.Context, requests []GetRequest, opts *BatchOptions) []BatchResult {
	batchRequests := make([]BatchRequest, len(requests))
	for i, req := range requests {
		batchRequests[i] = BatchRequest{
			Request: &Request{
				Method:   "GET",
				Endpoint: req.Endpoint,
				Query:    req.Query,
			},
			Result: req.Result,
		}
	}
	return c.Batch(ctx, batchRequests, opts)
}

// GetRequest represents a GET request for batch operations.
type GetRequest struct {
	Endpoint string
	Query    url.Values
	Result   interface{}
}

// BatchSequential executes requests sequentially, useful when order matters
// or when you want to stop on first error without race conditions.
func (c *Client) BatchSequential(ctx context.Context, requests []BatchRequest) []BatchResult {
	results := make([]BatchResult, len(requests))

	for i, req := range requests {
		if ctx.Err() != nil {
			results[i] = BatchResult{Index: i, Error: ctx.Err()}
			continue
		}

		err := c.Do(ctx, req.Request, req.Result)
		results[i] = BatchResult{Index: i, Error: err}
	}

	return results
}

// BatchWithCallback executes requests and calls the callback for each result.
// This is useful for processing results as they complete.
// Note: The callback is serialized with a mutex to prevent concurrent calls.
// If the callback panics, the panic is recovered and the goroutine continues.
func (c *Client) BatchWithCallback(ctx context.Context, requests []BatchRequest, opts *BatchOptions, callback func(BatchResult)) {
	if opts == nil {
		defaultOpts := DefaultBatchOptions()
		opts = &defaultOpts
	}

	if len(requests) == 0 {
		return
	}

	var sem chan struct{}
	if opts.MaxConcurrent > 0 {
		sem = make(chan struct{}, opts.MaxConcurrent)
	}

	var wg sync.WaitGroup
	var stopMu sync.Mutex
	stopped := false
	var callbackMu sync.Mutex

	// safeCallback calls the callback with panic recovery to prevent deadlock
	safeCallback := func(result BatchResult) {
		callbackMu.Lock()
		defer callbackMu.Unlock()
		defer func() { _ = recover() }()
		callback(result)
	}

	for i, req := range requests {
		if opts.StopOnError {
			stopMu.Lock()
			shouldStop := stopped
			stopMu.Unlock()
			if shouldStop {
				safeCallback(BatchResult{Index: i, Error: context.Canceled})
				continue
			}
		}

		if ctx.Err() != nil {
			safeCallback(BatchResult{Index: i, Error: ctx.Err()})
			continue
		}

		if sem != nil {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				safeCallback(BatchResult{Index: i, Error: ctx.Err()})
				continue
			}
		}

		wg.Add(1)
		go func(idx int, batchReq BatchRequest) {
			defer wg.Done()
			if sem != nil {
				defer func() { <-sem }()
			}

			err := c.Do(ctx, batchReq.Request, batchReq.Result)
			result := BatchResult{Index: idx, Error: err}

			safeCallback(result)

			if err != nil && opts.StopOnError {
				stopMu.Lock()
				stopped = true
				stopMu.Unlock()
			}
		}(i, req)
	}

	wg.Wait()
}

// HasErrors returns true if any batch result contains an error.
func HasErrors(results []BatchResult) bool {
	for _, r := range results {
		if r.Error != nil {
			return true
		}
	}
	return false
}

// FirstError returns the first error from batch results, or nil if none.
func FirstError(results []BatchResult) error {
	for _, r := range results {
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

// Errors returns all errors from batch results.
func Errors(results []BatchResult) []error {
	var errs []error
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, r.Error)
		}
	}
	return errs
}
