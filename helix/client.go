// Package helix provides a client for the Twitch Helix API.
package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// HelixBaseURL is the base URL for the Twitch Helix API.
	HelixBaseURL = "https://api.twitch.tv/helix"

	// DefaultRateLimit is the default rate limit for the API.
	DefaultRateLimit = 800 // requests per minute
)

// TokenProvider is an interface for providing access tokens.
type TokenProvider interface {
	GetToken() *Token
}

// Client is a Twitch Helix API client.
type Client struct {
	clientID      string
	authClient    *AuthClient   // For full auth client functionality
	tokenProvider TokenProvider // For token-only providers (e.g., extension JWT)
	httpClient    *http.Client

	// Rate limiting
	rateLimitLimit     int       // Points added per minute (bucket size)
	rateLimitRemaining int       // Points remaining in bucket
	rateLimitReset     time.Time // When bucket resets
	rateMu             sync.Mutex

	// Retry configuration
	maxRetries     int           // Maximum retries on 429 (default: 3)
	retryEnabled   bool          // Whether to auto-retry on 429 (default: true)
	maxRetryWait   time.Duration // Maximum wait time for retry (default: 60s)
	baseRetryDelay time.Duration // Base delay for exponential backoff (default: 1s)
	useExpBackoff  bool          // Use exponential backoff instead of reset time (default: false)

	// Middleware
	middleware []Middleware

	// Cache
	cache        Cache
	cacheTTL     time.Duration
	cacheEnabled bool

	// Base URL (can be overridden for testing)
	baseURL string

	// Ingest base URL (can be overridden for testing)
	ingestBaseURL string
}

// Option is a function that configures the client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithIngestBaseURL sets a custom ingest base URL (useful for testing).
func WithIngestBaseURL(url string) Option {
	return func(c *Client) {
		c.ingestBaseURL = url
	}
}

// WithRetry configures retry behavior for rate-limited requests.
func WithRetry(enabled bool, maxRetries int) Option {
	return func(c *Client) {
		c.retryEnabled = enabled
		c.maxRetries = maxRetries
	}
}

// WithMaxRetryWait sets the maximum time to wait for a retry.
func WithMaxRetryWait(d time.Duration) Option {
	return func(c *Client) {
		c.maxRetryWait = d
	}
}

// WithExponentialBackoff enables exponential backoff for retries.
// baseDelay is the initial delay, which doubles with each retry attempt.
func WithExponentialBackoff(baseDelay time.Duration) Option {
	return func(c *Client) {
		c.useExpBackoff = true
		c.baseRetryDelay = baseDelay
	}
}

// WithMiddleware adds middleware to the client.
func WithMiddleware(mw ...Middleware) Option {
	return func(c *Client) {
		c.middleware = append(c.middleware, mw...)
	}
}

// NewClient creates a new Helix API client.
func NewClient(clientID string, authClient *AuthClient, opts ...Option) *Client {
	c := &Client{
		clientID:           clientID,
		authClient:         authClient,
		httpClient:         &http.Client{Timeout: 30 * time.Second},
		baseURL:            HelixBaseURL,
		ingestBaseURL:      IngestBaseURL,
		rateLimitLimit:     DefaultRateLimit,
		rateLimitRemaining: DefaultRateLimit,
		retryEnabled:       true,
		maxRetries:         3,
		maxRetryWait:       60 * time.Second,
		baseRetryDelay:     time.Second,
		cacheTTL:           5 * time.Minute,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Request represents an API request.
type Request struct {
	Method   string
	Endpoint string
	Query    url.Values
	Body     interface{}
}

// Response represents a generic API response.
type Response[T any] struct {
	Data         []T         `json:"data"`
	Pagination   *Pagination `json:"pagination,omitempty"`
	Total        *int        `json:"total,omitempty"`
	TotalCost    *int        `json:"total_cost,omitempty"`
	MaxTotalCost *int        `json:"max_total_cost,omitempty"`
}

// Pagination contains pagination information.
type Pagination struct {
	Cursor string `json:"cursor,omitempty"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// APIError represents a Twitch API error.
type APIError struct {
	StatusCode int
	ErrorType  string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("twitch api error %d: %s - %s", e.StatusCode, e.ErrorType, e.Message)
}

// RateLimitError is returned when the API rate limit is exceeded and retries are exhausted.
type RateLimitError struct {
	ResetAt    time.Time     // When the rate limit resets
	Remaining  int           // Points remaining (usually 0)
	Limit      int           // Total bucket size
	RetryAfter time.Duration // How long until reset
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %d/%d points remaining, resets in %v", e.Remaining, e.Limit, e.RetryAfter.Round(time.Second))
}

// IsRateLimitError returns true if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// cacheKey generates a cache key that includes base URL and token hash
// to prevent cache pollution across different clients or tokens.
func (c *Client) cacheKey(endpoint, query string) string {
	tokenHash := ""
	if c.authClient != nil {
		if token := c.authClient.GetToken(); token != nil {
			tokenHash = TokenHash(token.AccessToken)
		}
	} else if c.tokenProvider != nil {
		if token := c.tokenProvider.GetToken(); token != nil {
			tokenHash = TokenHash(token.AccessToken)
		}
	}
	return CacheKeyWithContext(c.baseURL, endpoint, query, tokenHash)
}

// Do executes an API request with automatic retry on rate limit (429).
func (c *Client) Do(ctx context.Context, req *Request, result interface{}) error {
	// Check cache for GET requests
	if c.cacheEnabled && c.cache != nil && req.Method == http.MethodGet && !shouldSkipCache(ctx) {
		key := c.cacheKey(req.Endpoint, req.Query.Encode())
		if cached := c.cache.Get(ctx, key); cached != nil {
			if result != nil {
				return json.Unmarshal(cached, result)
			}
			return nil
		}
	}

	// Execute with middleware if configured
	if len(c.middleware) > 0 {
		return c.doWithMiddleware(ctx, req, result)
	}

	return c.doWithRetry(ctx, req, result)
}

// doWithMiddleware executes request through the middleware chain.
func (c *Client) doWithMiddleware(ctx context.Context, req *Request, result interface{}) error {
	// Build middleware chain
	var chain MiddlewareNext
	chain = func(ctx context.Context, req *Request) (*MiddlewareResponse, error) {
		// Final handler - execute the actual request and capture response info
		return c.doWithRetryAndResponse(ctx, req, result)
	}

	// Wrap chain with middleware in reverse order
	for i := len(c.middleware) - 1; i >= 0; i-- {
		mw := c.middleware[i]
		next := chain
		chain = func(ctx context.Context, req *Request) (*MiddlewareResponse, error) {
			return mw(ctx, req, next)
		}
	}

	_, err := chain(ctx, req)
	return err
}

// doWithRetry executes a request with retry logic.
func (c *Client) doWithRetry(ctx context.Context, req *Request, result interface{}) error {
	_, err := c.doWithRetryAndResponse(ctx, req, result)
	return err
}

// doWithRetryAndResponse executes a request with retry logic and returns response info for middleware.
func (c *Client) doWithRetryAndResponse(ctx context.Context, req *Request, result interface{}) (*MiddlewareResponse, error) {
	var lastErr error
	var lastResp *MiddlewareResponse

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err := c.doOnceWithResponse(ctx, req, result)
		if err == nil {
			return resp, nil
		}

		lastResp = resp

		// Check if it's a rate limit error
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusTooManyRequests {
			if !c.retryEnabled || attempt >= c.maxRetries {
				// Return a RateLimitError with reset info
				c.rateMu.Lock()
				resetAt := c.rateLimitReset
				remaining := c.rateLimitRemaining
				limit := c.rateLimitLimit
				c.rateMu.Unlock()

				retryAfter := time.Until(resetAt)
				if retryAfter < 0 {
					retryAfter = 0
				}

				return lastResp, &RateLimitError{
					ResetAt:    resetAt,
					Remaining:  remaining,
					Limit:      limit,
					RetryAfter: retryAfter,
				}
			}

			// Calculate wait time
			var waitTime time.Duration
			if c.useExpBackoff {
				// Exponential backoff: baseDelay * 2^attempt
				waitTime = c.baseRetryDelay * (1 << uint(attempt))
			} else {
				// Wait until reset time
				c.rateMu.Lock()
				resetAt := c.rateLimitReset
				c.rateMu.Unlock()
				waitTime = time.Until(resetAt)
			}

			// Apply bounds
			if waitTime < 0 {
				waitTime = c.baseRetryDelay
			}
			if waitTime > c.maxRetryWait {
				waitTime = c.maxRetryWait
			}

			// Wait before retry
			select {
			case <-ctx.Done():
				return lastResp, ctx.Err()
			case <-time.After(waitTime):
				// Continue to next attempt
			}

			lastErr = err
			continue
		}

		// Not a rate limit error, return immediately
		return lastResp, err
	}

	return lastResp, lastErr
}

// doOnce executes a single API request without retries.
func (c *Client) doOnce(ctx context.Context, req *Request, result interface{}) error {
	_, err := c.doOnceWithResponse(ctx, req, result)
	return err
}

// doOnceWithResponse executes a single API request and returns response info for middleware.
func (c *Client) doOnceWithResponse(ctx context.Context, req *Request, result interface{}) (*MiddlewareResponse, error) {
	// Build URL
	reqURL := c.baseURL + req.Endpoint
	if len(req.Query) > 0 {
		reqURL += "?" + req.Query.Encode()
	}

	// Build body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Client-Id", c.clientID)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers from middleware context
	if customHeaders := headersFromContext(ctx); customHeaders != nil {
		for k, v := range customHeaders {
			httpReq.Header.Set(k, v)
		}
	}

	// Set authorization
	var token *Token
	if c.authClient != nil {
		token = c.authClient.GetToken()
	} else if c.tokenProvider != nil {
		token = c.tokenProvider.GetToken()
	}
	if token != nil && token.AccessToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Update rate limit info from headers
	c.updateRateLimit(resp)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Build middleware response with real data
	mwResp := &MiddlewareResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return mwResp, &APIError{
				StatusCode: resp.StatusCode,
				ErrorType:  "unknown",
				Message:    string(body),
			}
		}
		return mwResp, &APIError{
			StatusCode: resp.StatusCode,
			ErrorType:  errResp.Error,
			Message:    errResp.Message,
		}
	}

	// Parse response
	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return mwResp, fmt.Errorf("parsing response: %w", err)
		}
	}

	// Cache successful GET responses
	if c.cacheEnabled && c.cache != nil && req.Method == http.MethodGet && !shouldSkipCache(ctx) && len(body) > 0 {
		key := c.cacheKey(req.Endpoint, req.Query.Encode())
		c.cache.Set(ctx, key, body, c.cacheTTL)
	}

	return mwResp, nil
}

// updateRateLimit updates rate limit information from response headers.
// Only updates state when headers are present AND parse successfully to avoid poisoning state.
func (c *Client) updateRateLimit(resp *http.Response) {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	if limit := resp.Header.Get("Ratelimit-Limit"); limit != "" {
		var l int
		if n, err := fmt.Sscanf(limit, "%d", &l); err == nil && n == 1 && l > 0 {
			c.rateLimitLimit = l
		}
	}

	if remaining := resp.Header.Get("Ratelimit-Remaining"); remaining != "" {
		var r int
		if n, err := fmt.Sscanf(remaining, "%d", &r); err == nil && n == 1 && r >= 0 {
			c.rateLimitRemaining = r
		}
	}

	if reset := resp.Header.Get("Ratelimit-Reset"); reset != "" {
		var r int64
		if n, err := fmt.Sscanf(reset, "%d", &r); err == nil && n == 1 && r > 0 {
			c.rateLimitReset = time.Unix(r, 0)
		}
	}
}

// RateLimitInfo contains rate limit status information.
type RateLimitInfo struct {
	Limit     int       // Points added per minute (bucket size)
	Remaining int       // Points remaining in bucket
	ResetAt   time.Time // When the bucket resets to full
}

// GetRateLimitInfo returns current rate limit information.
func (c *Client) GetRateLimitInfo() RateLimitInfo {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()
	return RateLimitInfo{
		Limit:     c.rateLimitLimit,
		Remaining: c.rateLimitRemaining,
		ResetAt:   c.rateLimitReset,
	}
}

// WaitForRateLimit blocks until the rate limit resets or context is cancelled.
// Returns immediately if there are remaining points in the bucket.
func (c *Client) WaitForRateLimit(ctx context.Context) error {
	c.rateMu.Lock()
	remaining := c.rateLimitRemaining
	resetAt := c.rateLimitReset
	c.rateMu.Unlock()

	if remaining > 0 {
		return nil
	}

	waitTime := time.Until(resetAt)
	if waitTime <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		return nil
	}
}

// Helper functions for common request patterns

// get performs a GET request.
func (c *Client) get(ctx context.Context, endpoint string, query url.Values, result interface{}) error {
	return c.Do(ctx, &Request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
		Query:    query,
	}, result)
}

// post performs a POST request.
func (c *Client) post(ctx context.Context, endpoint string, query url.Values, body interface{}, result interface{}) error {
	return c.Do(ctx, &Request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Query:    query,
		Body:     body,
	}, result)
}

// put performs a PUT request.
func (c *Client) put(ctx context.Context, endpoint string, query url.Values, body interface{}, result interface{}) error {
	return c.Do(ctx, &Request{
		Method:   http.MethodPut,
		Endpoint: endpoint,
		Query:    query,
		Body:     body,
	}, result)
}

// patch performs a PATCH request.
func (c *Client) patch(ctx context.Context, endpoint string, query url.Values, body interface{}, result interface{}) error {
	return c.Do(ctx, &Request{
		Method:   http.MethodPatch,
		Endpoint: endpoint,
		Query:    query,
		Body:     body,
	}, result)
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, endpoint string, query url.Values, result interface{}) error {
	return c.Do(ctx, &Request{
		Method:   http.MethodDelete,
		Endpoint: endpoint,
		Query:    query,
	}, result)
}

// Pagination helpers

// PaginationParams contains common pagination parameters.
type PaginationParams struct {
	First  int    // Maximum number of items to return (1-100)
	After  string // Cursor for forward pagination
	Before string // Cursor for backward pagination
}

// addPaginationParams adds pagination parameters to a query.
func addPaginationParams(q url.Values, p *PaginationParams) {
	if p == nil {
		return
	}
	if p.First > 0 {
		q.Set("first", fmt.Sprintf("%d", p.First))
	}
	if p.After != "" {
		q.Set("after", p.After)
	}
	if p.Before != "" {
		q.Set("before", p.Before)
	}
}
