package helix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// EventSubWebSocketURL is the Twitch EventSub WebSocket endpoint.
	EventSubWebSocketURL = "wss://eventsub.wss.twitch.tv/ws"
)

// WebSocket message types
const (
	WSMessageTypeWelcome      = "session_welcome"
	WSMessageTypeKeepalive    = "session_keepalive"
	WSMessageTypeNotification = "notification"
	WSMessageTypeReconnect    = "session_reconnect"
	WSMessageTypeRevocation   = "revocation"
)

// WebSocket close codes from Twitch
const (
	WSCloseInternalError         = 4000 // Internal server error
	WSCloseClientSentInbound     = 4001 // Client sent inbound traffic
	WSCloseClientFailedPingPong  = 4002 // Client failed ping-pong
	WSCloseConnectionUnused      = 4003 // Connection unused (no subscriptions within 10s)
	WSCloseReconnectGraceExpired = 4004 // Reconnect grace time expired
	WSCloseNetworkTimeout        = 4005 // Network timeout
	WSCloseNetworkError          = 4006 // Network error
	WSCloseInvalidReconnect      = 4007 // Invalid reconnect
)

// isExpectedCloseError returns true for errors that are expected when
// a connection is deliberately closed (e.g., during reconnect or shutdown).
func isExpectedCloseError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, websocket.ErrCloseSent) {
		return true
	}
	// Check for common close-related error messages
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "i/o timeout")
}

// WebSocketMessage represents a message received from EventSub WebSocket.
type WebSocketMessage struct {
	Metadata WebSocketMetadata `json:"metadata"`
	Payload  json.RawMessage   `json:"payload"`
}

// WebSocketMetadata contains metadata about the WebSocket message.
type WebSocketMetadata struct {
	MessageID           string    `json:"message_id"`
	MessageType         string    `json:"message_type"`
	MessageTimestamp    time.Time `json:"message_timestamp"`
	SubscriptionType    string    `json:"subscription_type,omitempty"`
	SubscriptionVersion string    `json:"subscription_version,omitempty"`
}

// WebSocketWelcomePayload is the payload for welcome messages.
type WebSocketWelcomePayload struct {
	Session WebSocketSession `json:"session"`
}

// WebSocketSession contains session information.
type WebSocketSession struct {
	ID                      string    `json:"id"`
	Status                  string    `json:"status"`
	ConnectedAt             time.Time `json:"connected_at"`
	KeepaliveTimeoutSeconds int       `json:"keepalive_timeout_seconds"`
	ReconnectURL            string    `json:"reconnect_url,omitempty"`
}

// WebSocketReconnectPayload is the payload for reconnect messages.
type WebSocketReconnectPayload struct {
	Session WebSocketSession `json:"session"`
}

// WebSocketNotificationPayload is the payload for notification messages.
type WebSocketNotificationPayload struct {
	Subscription EventSubSubscription `json:"subscription"`
	Event        json.RawMessage      `json:"event"`
}

// ErrAlreadyConnecting is returned when Connect is called while already connecting.
var ErrAlreadyConnecting = errors.New("connection already in progress")

// EventSubWebSocketClient manages an EventSub WebSocket connection.
type EventSubWebSocketClient struct {
	url              string
	conn             *websocket.Conn
	sessionID        string
	keepaliveTimeout time.Duration

	// Handlers
	onWelcome      func(*WebSocketSession)
	onNotification func(subscription *EventSubSubscription, event json.RawMessage)
	onRevocation   func(*EventSubSubscription)
	onReconnect    func(reconnectURL string)
	onError        func(error)
	onKeepalive    func()

	// State
	mu           sync.RWMutex
	connected    bool
	connecting   bool           // prevents concurrent Connect() calls
	stopChan     chan struct{}
	stopOnce     sync.Once      // ensures stopChan is closed only once
	wg           sync.WaitGroup // tracks readLoop goroutine
	reconnectURL string
}

// EventSubWSOption configures the WebSocket client.
type EventSubWSOption func(*EventSubWebSocketClient)

// WithWSURL sets a custom WebSocket URL (useful for testing).
func WithWSURL(url string) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.url = url
	}
}

// WithWSWelcomeHandler sets the handler for welcome messages.
func WithWSWelcomeHandler(fn func(*WebSocketSession)) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.onWelcome = fn
	}
}

// WithWSNotificationHandler sets the handler for notification messages.
func WithWSNotificationHandler(fn func(*EventSubSubscription, json.RawMessage)) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.onNotification = fn
	}
}

// WithWSRevocationHandler sets the handler for revocation messages.
func WithWSRevocationHandler(fn func(*EventSubSubscription)) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.onRevocation = fn
	}
}

// WithWSReconnectHandler sets the handler for reconnect messages.
func WithWSReconnectHandler(fn func(string)) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.onReconnect = fn
	}
}

// WithWSErrorHandler sets the handler for errors.
func WithWSErrorHandler(fn func(error)) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.onError = fn
	}
}

// WithWSKeepaliveHandler sets the handler for keepalive messages.
func WithWSKeepaliveHandler(fn func()) EventSubWSOption {
	return func(c *EventSubWebSocketClient) {
		c.onKeepalive = fn
	}
}

// NewEventSubWebSocketClient creates a new EventSub WebSocket client.
func NewEventSubWebSocketClient(opts ...EventSubWSOption) *EventSubWebSocketClient {
	c := &EventSubWebSocketClient{
		url:      EventSubWebSocketURL,
		stopChan: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Connect establishes a WebSocket connection to EventSub.
// Returns the session ID that should be used when creating subscriptions.
// This method is safe for concurrent use - only one connection attempt will proceed.
func (c *EventSubWebSocketClient) Connect(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.connected {
		sessionID := c.sessionID
		c.mu.Unlock()
		return sessionID, nil
	}
	if c.connecting {
		c.mu.Unlock()
		return "", ErrAlreadyConnecting
	}
	c.connecting = true
	c.mu.Unlock()

	// Ensure we clear connecting flag on exit
	defer func() {
		c.mu.Lock()
		c.connecting = false
		c.mu.Unlock()
	}()

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return "", fmt.Errorf("connecting to websocket: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.stopChan = make(chan struct{})
	c.stopOnce = sync.Once{} // reset for new connection
	c.mu.Unlock()

	// Wait for welcome message
	sessionID, err := c.waitForWelcome(ctx)
	if err != nil {
		// Clean up state on error
		c.mu.Lock()
		c.conn = nil
		c.stopChan = nil
		c.mu.Unlock()
		_ = conn.Close()
		return "", err
	}

	c.mu.Lock()
	c.sessionID = sessionID
	c.connected = true
	c.mu.Unlock()

	// Start message handler
	c.wg.Add(1)
	go c.readLoop()

	return sessionID, nil
}

// waitForWelcome waits for and processes the welcome message.
func (c *EventSubWebSocketClient) waitForWelcome(ctx context.Context) (string, error) {
	// Use 10 seconds per Twitch docs, but respect context deadline if sooner
	deadline := time.Now().Add(10 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	_ = c.conn.SetReadDeadline(deadline)
	defer func() { _ = c.conn.SetReadDeadline(time.Time{}) }()

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return "", fmt.Errorf("reading welcome message: %w", err)
	}

	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return "", fmt.Errorf("parsing welcome message: %w", err)
	}

	if msg.Metadata.MessageType != WSMessageTypeWelcome {
		return "", fmt.Errorf("expected welcome message, got %s", msg.Metadata.MessageType)
	}

	var payload WebSocketWelcomePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return "", fmt.Errorf("parsing welcome payload: %w", err)
	}

	c.mu.Lock()
	c.keepaliveTimeout = time.Duration(payload.Session.KeepaliveTimeoutSeconds) * time.Second
	c.mu.Unlock()

	if c.onWelcome != nil {
		c.onWelcome(&payload.Session)
	}

	return payload.Session.ID, nil
}

// readLoop continuously reads messages from the WebSocket.
func (c *EventSubWebSocketClient) readLoop() {
	defer c.wg.Done()
	defer func() {
		c.mu.Lock()
		c.connected = false
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.mu.Unlock()
	}()

	for {
		// Capture connection and stopChan under lock
		c.mu.RLock()
		conn := c.conn
		stopChan := c.stopChan
		timeout := c.keepaliveTimeout
		c.mu.RUnlock()

		// Check if we should stop
		select {
		case <-stopChan:
			return
		default:
		}

		// Check for nil connection
		if conn == nil {
			return
		}

		// Set read deadline based on keepalive timeout (with buffer)
		if timeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(timeout + 10*time.Second))
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			// Don't report errors for expected connection close scenarios
			if c.onError != nil && !isExpectedCloseError(err) {
				c.onError(fmt.Errorf("reading message: %w", err))
			}
			return
		}

		c.handleMessage(data)
	}
}

// handleMessage processes a received WebSocket message.
func (c *EventSubWebSocketClient) handleMessage(data []byte) {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("parsing message: %w", err))
		}
		return
	}

	// Recover from handler panics to prevent crashing the read loop
	defer func() {
		if r := recover(); r != nil {
			if c.onError != nil {
				c.onError(fmt.Errorf("handler panic: %v", r))
			}
		}
	}()

	switch msg.Metadata.MessageType {
	case WSMessageTypeKeepalive:
		if c.onKeepalive != nil {
			c.onKeepalive()
		}

	case WSMessageTypeNotification:
		c.handleNotification(msg)

	case WSMessageTypeReconnect:
		c.handleReconnect(msg)

	case WSMessageTypeRevocation:
		c.handleRevocation(msg)
	}
}

// handleNotification processes a notification message.
func (c *EventSubWebSocketClient) handleNotification(msg WebSocketMessage) {
	if c.onNotification == nil {
		return
	}

	var payload WebSocketNotificationPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("parsing notification: %w", err))
		}
		return
	}

	c.onNotification(&payload.Subscription, payload.Event)
}

// handleReconnect processes a reconnect message.
func (c *EventSubWebSocketClient) handleReconnect(msg WebSocketMessage) {
	var payload WebSocketReconnectPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("parsing reconnect: %w", err))
		}
		return
	}

	c.mu.Lock()
	c.reconnectURL = payload.Session.ReconnectURL
	c.mu.Unlock()

	if c.onReconnect != nil {
		c.onReconnect(payload.Session.ReconnectURL)
	}
}

// handleRevocation processes a revocation message.
func (c *EventSubWebSocketClient) handleRevocation(msg WebSocketMessage) {
	if c.onRevocation == nil {
		return
	}

	var payload WebSocketNotificationPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		if c.onError != nil {
			c.onError(fmt.Errorf("parsing revocation: %w", err))
		}
		return
	}

	c.onRevocation(&payload.Subscription)
}

// Close closes the WebSocket connection.
func (c *EventSubWebSocketClient) Close() error {
	c.mu.Lock()
	// Close if connected OR if there's an in-progress connection (conn != nil)
	if !c.connected && c.conn == nil {
		c.mu.Unlock()
		return nil
	}

	// Signal readLoop to stop (only once to prevent panic)
	c.stopOnce.Do(func() {
		if c.stopChan != nil {
			close(c.stopChan)
		}
	})
	c.connected = false
	conn := c.conn
	c.mu.Unlock()

	// Close connection to unblock ReadMessage in readLoop
	if conn != nil {
		// Set immediate deadline to ensure ReadMessage unblocks quickly
		_ = conn.SetReadDeadline(time.Now())
		_ = conn.Close()
	}

	// Wait for readLoop to finish
	c.wg.Wait()

	return nil
}

// SessionID returns the current session ID.
func (c *EventSubWebSocketClient) SessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// IsConnected returns whether the client is connected.
func (c *EventSubWebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Reconnect connects to a new URL (typically from a reconnect message).
func (c *EventSubWebSocketClient) Reconnect(ctx context.Context, url string) (string, error) {
	// Stop old readLoop first
	c.mu.Lock()
	oldConn := c.conn
	if c.stopChan != nil {
		c.stopOnce.Do(func() {
			close(c.stopChan)
		})
	}
	c.mu.Unlock()

	// Close old connection to unblock ReadMessage in readLoop
	if oldConn != nil {
		// Set immediate deadline to ensure ReadMessage unblocks quickly
		_ = oldConn.SetReadDeadline(time.Now())
		_ = oldConn.Close()
	}

	// Wait for old readLoop to finish
	c.wg.Wait()

	// Connect to new URL
	newConn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return "", fmt.Errorf("connecting to reconnect URL: %w", err)
	}

	c.mu.Lock()
	c.conn = newConn
	c.stopChan = make(chan struct{})
	c.stopOnce = sync.Once{} // reset for new connection
	c.mu.Unlock()

	// Wait for welcome on new connection
	sessionID, err := c.waitForWelcome(ctx)
	if err != nil {
		// Clean up state on error
		c.mu.Lock()
		c.conn = nil
		c.stopChan = nil
		c.mu.Unlock()
		_ = newConn.Close()
		return "", err
	}

	c.mu.Lock()
	c.sessionID = sessionID
	c.connected = true
	c.mu.Unlock()

	// Start new read loop
	c.wg.Add(1)
	go c.readLoop()

	return sessionID, nil
}

// ParseWSEvent parses the event JSON into the specified type.
func ParseWSEvent[T any](eventData json.RawMessage) (*T, error) {
	var event T
	if err := json.Unmarshal(eventData, &event); err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}
	return &event, nil
}

// EventSubWebSocket provides a higher-level interface for EventSub WebSocket.
type EventSubWebSocket struct {
	client    *Client
	ws        *EventSubWebSocketClient
	sessionID string

	mu       sync.RWMutex
	handlers map[string]func(json.RawMessage)

	// User-provided handlers
	onRevocation func(eventType string, reason string)
	onReconnect  func()
	onError      func(error)
}

// EventSubWebSocketOption configures the high-level EventSub WebSocket manager.
type EventSubWebSocketOption func(*EventSubWebSocket)

// WithEventSubRevocationHandler sets the handler for subscription revocations.
// The handler receives the event type that was revoked and the reason.
func WithEventSubRevocationHandler(fn func(eventType string, reason string)) EventSubWebSocketOption {
	return func(e *EventSubWebSocket) {
		e.onRevocation = fn
	}
}

// WithEventSubReconnectHandler sets the handler called after a successful reconnect.
func WithEventSubReconnectHandler(fn func()) EventSubWebSocketOption {
	return func(e *EventSubWebSocket) {
		e.onReconnect = fn
	}
}

// WithEventSubErrorHandler sets the handler for WebSocket errors.
func WithEventSubErrorHandler(fn func(error)) EventSubWebSocketOption {
	return func(e *EventSubWebSocket) {
		e.onError = fn
	}
}

// NewEventSubWebSocket creates a new high-level EventSub WebSocket manager.
// Returns nil if helixClient is nil.
func NewEventSubWebSocket(helixClient *Client, opts ...EventSubWebSocketOption) *EventSubWebSocket {
	if helixClient == nil {
		return nil
	}

	e := &EventSubWebSocket{
		client:   helixClient,
		handlers: make(map[string]func(json.RawMessage)),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Connect establishes the WebSocket connection.
// If already connected, the existing connection is closed first.
func (e *EventSubWebSocket) Connect(ctx context.Context) error {
	// Close existing connection if any
	if e.ws != nil {
		_ = e.ws.Close()
		e.ws = nil
		e.sessionID = ""
	}

	e.ws = NewEventSubWebSocketClient(
		WithWSNotificationHandler(func(sub *EventSubSubscription, event json.RawMessage) {
			e.mu.RLock()
			handler, ok := e.handlers[sub.Type]
			e.mu.RUnlock()
			if ok {
				handler(event)
			}
		}),
		WithWSRevocationHandler(func(sub *EventSubSubscription) {
			// Remove handler for revoked subscription
			e.mu.Lock()
			delete(e.handlers, sub.Type)
			e.mu.Unlock()

			// Notify user if handler is set
			if e.onRevocation != nil {
				e.onRevocation(sub.Type, sub.Status)
			}
		}),
		WithWSReconnectHandler(func(reconnectURL string) {
			// Handle reconnect automatically
			go e.handleReconnect(reconnectURL)
		}),
		WithWSErrorHandler(func(err error) {
			if e.onError != nil {
				e.onError(err)
			}
		}),
	)

	sessionID, err := e.ws.Connect(ctx)
	if err != nil {
		e.ws = nil
		return err
	}
	e.sessionID = sessionID
	return nil
}

// handleReconnect handles the reconnect process when Twitch sends a reconnect message.
func (e *EventSubWebSocket) handleReconnect(reconnectURL string) {
	// Copy ws under lock to avoid race with Close/Connect
	e.mu.RLock()
	ws := e.ws
	e.mu.RUnlock()

	if ws == nil {
		if e.onError != nil {
			e.onError(fmt.Errorf("reconnect failed: connection already closed"))
		}
		return
	}

	// Use a background context with timeout for reconnection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newSessionID, err := ws.Reconnect(ctx, reconnectURL)
	if err != nil {
		if e.onError != nil {
			e.onError(fmt.Errorf("reconnect failed: %w", err))
		}
		return
	}

	e.mu.Lock()
	e.sessionID = newSessionID
	e.mu.Unlock()

	if e.onReconnect != nil {
		e.onReconnect()
	}
}

// Subscribe creates a subscription for the given event type.
// Returns an error if not connected.
func (e *EventSubWebSocket) Subscribe(ctx context.Context, eventType, version string, condition map[string]string, handler func(json.RawMessage)) error {
	if e.sessionID == "" {
		return errors.New("not connected: call Connect first")
	}

	// Create subscription via API first
	_, err := e.client.CreateEventSubSubscription(ctx, &CreateEventSubSubscriptionParams{
		Type:      eventType,
		Version:   version,
		Condition: condition,
		Transport: CreateEventSubTransport{
			Method:    "websocket",
			SessionID: e.sessionID,
		},
	})
	if err != nil {
		return err
	}

	// Only register handler after successful subscription
	e.mu.Lock()
	e.handlers[eventType] = handler
	e.mu.Unlock()

	return nil
}

// Close closes the WebSocket connection.
func (e *EventSubWebSocket) Close() error {
	if e.ws != nil {
		return e.ws.Close()
	}
	return nil
}
