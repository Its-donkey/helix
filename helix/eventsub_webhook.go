package helix

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// EventSub webhook message types
const (
	EventSubMessageTypeNotification = "notification"
	EventSubMessageTypeVerification = "webhook_callback_verification"
	EventSubMessageTypeRevocation   = "revocation"
)

// EventSub headers
const (
	EventSubHeaderMessageID           = "Twitch-Eventsub-Message-Id"
	EventSubHeaderMessageTimestamp    = "Twitch-Eventsub-Message-Timestamp"
	EventSubHeaderMessageSignature    = "Twitch-Eventsub-Message-Signature"
	EventSubHeaderMessageType         = "Twitch-Eventsub-Message-Type"
	EventSubHeaderSubscriptionType    = "Twitch-Eventsub-Subscription-Type"
	EventSubHeaderSubscriptionVersion = "Twitch-Eventsub-Subscription-Version"
)

// EventSubWebhookMessage represents a message received from EventSub webhooks.
type EventSubWebhookMessage struct {
	MessageID           string
	MessageTimestamp    time.Time
	MessageType         string
	SubscriptionType    string
	SubscriptionVersion string
	Subscription        EventSubSubscription
	Challenge           string
	Event               json.RawMessage
}

// EventSubWebhookPayload represents the JSON payload from EventSub.
type EventSubWebhookPayload struct {
	Subscription EventSubSubscription `json:"subscription"`
	Challenge    string               `json:"challenge,omitempty"`
	Event        json.RawMessage      `json:"event,omitempty"`
}

// EventSubWebhookHandler handles EventSub webhook requests.
type EventSubWebhookHandler struct {
	secret          string
	maxTimestampAge time.Duration
	onNotification  func(*EventSubWebhookMessage)
	onVerification  func(*EventSubWebhookMessage) bool
	onRevocation    func(*EventSubWebhookMessage)
}

// EventSubWebhookOption configures the webhook handler.
type EventSubWebhookOption func(*EventSubWebhookHandler)

// WithWebhookSecret sets the secret used for signature verification.
func WithWebhookSecret(secret string) EventSubWebhookOption {
	return func(h *EventSubWebhookHandler) {
		h.secret = secret
	}
}

// WithMaxTimestampAge sets the maximum age for message timestamps (default: 10 minutes).
func WithMaxTimestampAge(d time.Duration) EventSubWebhookOption {
	return func(h *EventSubWebhookHandler) {
		h.maxTimestampAge = d
	}
}

// WithNotificationHandler sets the handler for notification messages.
func WithNotificationHandler(fn func(*EventSubWebhookMessage)) EventSubWebhookOption {
	return func(h *EventSubWebhookHandler) {
		h.onNotification = fn
	}
}

// WithVerificationHandler sets the handler for verification challenges.
// Return true to accept the subscription, false to reject.
func WithVerificationHandler(fn func(*EventSubWebhookMessage) bool) EventSubWebhookOption {
	return func(h *EventSubWebhookHandler) {
		h.onVerification = fn
	}
}

// WithRevocationHandler sets the handler for subscription revocations.
func WithRevocationHandler(fn func(*EventSubWebhookMessage)) EventSubWebhookOption {
	return func(h *EventSubWebhookHandler) {
		h.onRevocation = fn
	}
}

// NewEventSubWebhookHandler creates a new EventSub webhook handler.
func NewEventSubWebhookHandler(opts ...EventSubWebhookOption) *EventSubWebhookHandler {
	h := &EventSubWebhookHandler{
		maxTimestampAge: 10 * time.Minute,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// ServeHTTP implements http.Handler for the webhook handler.
func (h *EventSubWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Verify signature
	if h.secret != "" {
		if !h.verifySignature(r.Header, body) {
			http.Error(w, "Invalid signature", http.StatusForbidden)
			return
		}
	}

	// Parse message
	msg, err := h.parseMessage(r.Header, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check timestamp age (replay attack prevention)
	if time.Since(msg.MessageTimestamp) > h.maxTimestampAge {
		http.Error(w, "Message timestamp too old", http.StatusBadRequest)
		return
	}

	// Handle based on message type
	switch msg.MessageType {
	case EventSubMessageTypeVerification:
		h.handleVerification(w, msg)
	case EventSubMessageTypeNotification:
		h.handleNotification(w, msg)
	case EventSubMessageTypeRevocation:
		h.handleRevocation(w, msg)
	default:
		http.Error(w, "Unknown message type", http.StatusBadRequest)
	}
}

// verifySignature verifies the HMAC-SHA256 signature of the message.
func (h *EventSubWebhookHandler) verifySignature(headers http.Header, body []byte) bool {
	messageID := headers.Get(EventSubHeaderMessageID)
	timestamp := headers.Get(EventSubHeaderMessageTimestamp)
	signature := headers.Get(EventSubHeaderMessageSignature)

	if messageID == "" || timestamp == "" || signature == "" {
		return false
	}

	// Build the message to sign: message_id + message_timestamp + body
	message := []byte(messageID + timestamp)
	message = append(message, body...)

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(message)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// parseMessage parses the webhook message from headers and body.
func (h *EventSubWebhookHandler) parseMessage(headers http.Header, body []byte) (*EventSubWebhookMessage, error) {
	timestamp, err := time.Parse(time.RFC3339, headers.Get(EventSubHeaderMessageTimestamp))
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	var payload EventSubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	return &EventSubWebhookMessage{
		MessageID:           headers.Get(EventSubHeaderMessageID),
		MessageTimestamp:    timestamp,
		MessageType:         headers.Get(EventSubHeaderMessageType),
		SubscriptionType:    headers.Get(EventSubHeaderSubscriptionType),
		SubscriptionVersion: headers.Get(EventSubHeaderSubscriptionVersion),
		Subscription:        payload.Subscription,
		Challenge:           payload.Challenge,
		Event:               payload.Event,
	}, nil
}

// handleVerification handles webhook verification challenges.
func (h *EventSubWebhookHandler) handleVerification(w http.ResponseWriter, msg *EventSubWebhookMessage) {
	accept := true
	if h.onVerification != nil {
		accept = h.onVerification(msg)
	}

	if accept {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(msg.Challenge))
	} else {
		http.Error(w, "Subscription rejected", http.StatusBadRequest)
	}
}

// handleNotification handles event notifications.
func (h *EventSubWebhookHandler) handleNotification(w http.ResponseWriter, msg *EventSubWebhookMessage) {
	if h.onNotification != nil {
		h.onNotification(msg)
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleRevocation handles subscription revocations.
func (h *EventSubWebhookHandler) handleRevocation(w http.ResponseWriter, msg *EventSubWebhookMessage) {
	if h.onRevocation != nil {
		h.onRevocation(msg)
	}
	w.WriteHeader(http.StatusNoContent)
}

// VerifyEventSubSignature verifies an EventSub webhook signature.
// This is useful for custom handler implementations.
func VerifyEventSubSignature(secret, messageID, timestamp string, body []byte, signature string) bool {
	message := []byte(messageID + timestamp)
	message = append(message, body...)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// ParseEventSubEvent parses the event data from a webhook message into the provided type.
func ParseEventSubEvent[T any](msg *EventSubWebhookMessage) (*T, error) {
	var event T
	if err := json.Unmarshal(msg.Event, &event); err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}
	return &event, nil
}

// EventSubMiddleware returns middleware that verifies EventSub signatures.
// Use this with your own http.Handler if you need more control.
func EventSubMiddleware(secret string, maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read and buffer body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Verify signature
			if !VerifyEventSubSignature(
				secret,
				r.Header.Get(EventSubHeaderMessageID),
				r.Header.Get(EventSubHeaderMessageTimestamp),
				body,
				r.Header.Get(EventSubHeaderMessageSignature),
			) {
				http.Error(w, "Invalid signature", http.StatusForbidden)
				return
			}

			// Check timestamp
			timestamp, err := time.Parse(time.RFC3339, r.Header.Get(EventSubHeaderMessageTimestamp))
			if err != nil || time.Since(timestamp) > maxAge {
				http.Error(w, "Invalid or expired timestamp", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MessageDeduplicator helps prevent processing duplicate EventSub messages.
// It is safe for concurrent use.
type MessageDeduplicator struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	maxAge  time.Duration
	maxSize int
}

// NewMessageDeduplicator creates a new message deduplicator.
func NewMessageDeduplicator(maxAge time.Duration, maxSize int) *MessageDeduplicator {
	return &MessageDeduplicator{
		seen:    make(map[string]time.Time),
		maxAge:  maxAge,
		maxSize: maxSize,
	}
}

// IsDuplicate returns true if this message ID has been seen before.
// It also marks the message as seen. This method is safe for concurrent use.
func (d *MessageDeduplicator) IsDuplicate(messageID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Clean up old entries if we're at capacity
	if len(d.seen) >= d.maxSize {
		for id, seenAt := range d.seen {
			if now.Sub(seenAt) > d.maxAge {
				delete(d.seen, id)
			}
		}
	}

	// Check if we've seen this message
	if _, ok := d.seen[messageID]; ok {
		return true
	}

	// Mark as seen
	d.seen[messageID] = now
	return false
}

// Clear removes all tracked message IDs. This method is safe for concurrent use.
func (d *MessageDeduplicator) Clear() {
	d.mu.Lock()
	d.seen = make(map[string]time.Time)
	d.mu.Unlock()
}

// Common revocation reasons
const (
	RevocationReasonAuthorizationRevoked = "authorization_revoked"
	RevocationReasonUserRemoved          = "user_removed"
	RevocationReasonNotificationFailures = "notification_failures_exceeded"
	RevocationReasonVersionRemoved       = "version_removed"
	RevocationReasonModeratorsChanged    = "moderator_removed"
)

// GetRevocationReason extracts the revocation reason from a subscription status.
func GetRevocationReason(subscription EventSubSubscription) string {
	// Status contains the reason for revoked subscriptions
	if strings.HasPrefix(subscription.Status, "authorization_revoked") {
		return RevocationReasonAuthorizationRevoked
	}
	return subscription.Status
}
