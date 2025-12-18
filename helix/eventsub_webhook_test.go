package helix

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEventSubWebhookHandler_Verification(t *testing.T) {
	secret := "test-webhook-secret"
	challenge := "test-challenge-string"

	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithVerificationHandler(func(msg *EventSubWebhookMessage) bool {
			receivedMsg = msg
			return true
		}),
	)

	// Create verification payload
	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "webhook_callback_verification_pending",
		},
		Challenge: challenge,
	}
	body, _ := json.Marshal(payload)

	// Create request with proper headers
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	messageID := "msg-123"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)
	req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeStreamOnline)

	// Sign the message
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	// Execute request
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != challenge {
		t.Errorf("expected challenge %s, got %s", challenge, w.Body.String())
	}
	if receivedMsg == nil {
		t.Error("verification handler was not called")
	}
	if receivedMsg.Challenge != challenge {
		t.Errorf("expected challenge in message, got %s", receivedMsg.Challenge)
	}
}

func TestEventSubWebhookHandler_Notification(t *testing.T) {
	secret := "test-webhook-secret"

	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithNotificationHandler(func(msg *EventSubWebhookMessage) {
			receivedMsg = msg
		}),
	)

	// Create notification payload
	event := StreamOnlineEvent{
		ID: "stream123",
		EventSubBroadcaster: EventSubBroadcaster{
			BroadcasterUserID:    "12345",
			BroadcasterUserLogin: "testuser",
			BroadcasterUserName:  "TestUser",
		},
		Type:      "live",
		StartedAt: time.Now(),
	}
	eventJSON, _ := json.Marshal(event)

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "enabled",
		},
		Event: eventJSON,
	}
	body, _ := json.Marshal(payload)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	messageID := "msg-456"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)
	req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeStreamOnline)

	// Sign the message
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	// Execute request
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if receivedMsg == nil {
		t.Fatal("notification handler was not called")
	}
	if receivedMsg.SubscriptionType != EventSubTypeStreamOnline {
		t.Errorf("expected subscription type %s, got %s", EventSubTypeStreamOnline, receivedMsg.SubscriptionType)
	}

	// Parse the event
	parsedEvent, err := ParseEventSubEvent[StreamOnlineEvent](receivedMsg)
	if err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}
	if parsedEvent.ID != "stream123" {
		t.Errorf("expected event ID stream123, got %s", parsedEvent.ID)
	}
	if parsedEvent.BroadcasterUserID != "12345" {
		t.Errorf("expected broadcaster ID 12345, got %s", parsedEvent.BroadcasterUserID)
	}
}

func TestEventSubWebhookHandler_InvalidSignature(t *testing.T) {
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret("correct-secret"),
	)

	body := []byte(`{"subscription":{},"challenge":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)
	req.Header.Set(EventSubHeaderMessageSignature, "sha256=invalid")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_ExpiredTimestamp(t *testing.T) {
	secret := "test-secret"
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithMaxTimestampAge(5*time.Minute),
	)

	body := []byte(`{"subscription":{},"challenge":"test"}`)
	messageID := "msg-123"
	// Timestamp 10 minutes ago
	timestamp := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)

	// Sign with old timestamp
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for expired timestamp, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_Revocation(t *testing.T) {
	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithRevocationHandler(func(msg *EventSubWebhookMessage) {
			receivedMsg = msg
		}),
	)

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "authorization_revoked",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-789")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeRevocation)
	req.Header.Set(EventSubHeaderSubscriptionType, EventSubTypeStreamOnline)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if receivedMsg == nil {
		t.Error("revocation handler was not called")
	}
	if receivedMsg.Subscription.Status != "authorization_revoked" {
		t.Errorf("expected status authorization_revoked, got %s", receivedMsg.Subscription.Status)
	}
}

func TestVerifyEventSubSignature(t *testing.T) {
	secret := "my-secret"
	messageID := "msg-123"
	timestamp := "2024-01-01T00:00:00Z"
	body := []byte(`{"test": "data"}`)

	// Create valid signature
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	validSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name      string
		signature string
		expected  bool
	}{
		{"valid signature", validSig, true},
		{"invalid signature", "sha256=invalid", false},
		{"empty signature", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyEventSubSignature(secret, messageID, timestamp, body, tt.signature)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMessageDeduplicator(t *testing.T) {
	dedup := NewMessageDeduplicator(time.Minute, 100)

	// First occurrence should not be duplicate
	if dedup.IsDuplicate("msg-1") {
		t.Error("first occurrence should not be duplicate")
	}

	// Second occurrence should be duplicate
	if !dedup.IsDuplicate("msg-1") {
		t.Error("second occurrence should be duplicate")
	}

	// Different message should not be duplicate
	if dedup.IsDuplicate("msg-2") {
		t.Error("different message should not be duplicate")
	}

	// After clear, should not be duplicate
	dedup.Clear()
	if dedup.IsDuplicate("msg-1") {
		t.Error("after clear, should not be duplicate")
	}
}

func TestMessageDeduplicator_Cleanup(t *testing.T) {
	// Test cleanup when at capacity
	dedup := NewMessageDeduplicator(time.Millisecond, 2)

	// Add first message
	dedup.IsDuplicate("msg-1")
	// Wait for it to expire
	time.Sleep(5 * time.Millisecond)

	// Add more messages to reach capacity and trigger cleanup
	dedup.IsDuplicate("msg-2")
	dedup.IsDuplicate("msg-3") // This should trigger cleanup of expired msg-1

	// msg-1 should have been cleaned up since it was expired
	if dedup.IsDuplicate("msg-1") {
		t.Error("expired msg-1 should not be duplicate after cleanup")
	}
}

func TestEventSubMiddleware(t *testing.T) {
	secret := "test-secret"
	maxAge := 10 * time.Minute

	middleware := EventSubMiddleware(secret, maxAge)

	t.Run("valid request", func(t *testing.T) {
		// Create a handler that will be called if signature is valid
		called := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(nextHandler)

		body := []byte(`{"subscription":{},"event":{}}`)
		messageID := "msg-123"
		timestamp := time.Now().UTC().Format(time.RFC3339)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, messageID)
		req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)

		// Create valid signature
		message := messageID + timestamp + string(body)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(message))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set(EventSubHeaderMessageSignature, signature)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if !called {
			t.Error("expected next handler to be called")
		}
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called for invalid signature")
		})

		handler := middleware(nextHandler)

		body := []byte(`{"subscription":{}}`)
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, "msg-123")
		req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
		req.Header.Set(EventSubHeaderMessageSignature, "sha256=invalid")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("expired timestamp", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called for expired timestamp")
		})

		handler := middleware(nextHandler)

		body := []byte(`{"subscription":{}}`)
		messageID := "msg-123"
		timestamp := time.Now().Add(-15 * time.Minute).UTC().Format(time.RFC3339)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, messageID)
		req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)

		// Create valid signature with expired timestamp
		message := messageID + timestamp + string(body)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(message))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set(EventSubHeaderMessageSignature, signature)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid timestamp format", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called for invalid timestamp")
		})

		handler := middleware(nextHandler)

		body := []byte(`{"subscription":{}}`)
		messageID := "msg-123"
		timestamp := "invalid-timestamp"

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, messageID)
		req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)

		// Create signature with invalid timestamp
		message := messageID + timestamp + string(body)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(message))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set(EventSubHeaderMessageSignature, signature)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestGetRevocationReason(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "authorization_revoked status",
			status:   "authorization_revoked",
			expected: RevocationReasonAuthorizationRevoked,
		},
		{
			name:     "authorization_revoked with details",
			status:   "authorization_revoked_by_user",
			expected: RevocationReasonAuthorizationRevoked,
		},
		{
			name:     "user_removed status",
			status:   "user_removed",
			expected: "user_removed",
		},
		{
			name:     "notification_failures_exceeded status",
			status:   "notification_failures_exceeded",
			expected: "notification_failures_exceeded",
		},
		{
			name:     "other status",
			status:   "some_other_reason",
			expected: "some_other_reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := EventSubSubscription{Status: tt.status}
			result := GetRevocationReason(subscription)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEventSubWebhookHandler_MethodNotAllowed(t *testing.T) {
	handler := NewEventSubWebhookHandler()

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_InvalidPayloadJSON(t *testing.T) {
	handler := NewEventSubWebhookHandler()

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_InvalidTimestamp(t *testing.T) {
	handler := NewEventSubWebhookHandler()

	body := []byte(`{"subscription":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, "invalid-timestamp")
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_UnknownMessageType(t *testing.T) {
	handler := NewEventSubWebhookHandler()

	body := []byte(`{"subscription":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, "unknown_type")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_VerificationRejected(t *testing.T) {
	handler := NewEventSubWebhookHandler(
		WithVerificationHandler(func(msg *EventSubWebhookMessage) bool {
			return false // Reject the subscription
		}),
	)

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:   "sub123",
			Type: EventSubTypeStreamOnline,
		},
		Challenge: "test-challenge",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for rejected verification, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_NotificationWithoutHandler(t *testing.T) {
	// Test notification without a handler configured
	handler := NewEventSubWebhookHandler()

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:   "sub123",
			Type: EventSubTypeStreamOnline,
		},
		Event: []byte(`{"id":"event123"}`),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_RevocationWithoutHandler(t *testing.T) {
	// Test revocation without a handler configured
	handler := NewEventSubWebhookHandler()

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:     "sub123",
			Type:   EventSubTypeStreamOnline,
			Status: "authorization_revoked",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeRevocation)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_MissingSignatureHeaders(t *testing.T) {
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret("test-secret"),
	)

	body := []byte(`{"subscription":{}}`)

	// Test missing message ID
	t.Run("missing message ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
		req.Header.Set(EventSubHeaderMessageSignature, "sha256=something")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	// Test missing timestamp
	t.Run("missing timestamp", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, "msg-123")
		req.Header.Set(EventSubHeaderMessageSignature, "sha256=something")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	// Test missing signature
	t.Run("missing signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, "msg-123")
		req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})
}

func TestParseEventSubEvent_Error(t *testing.T) {
	msg := &EventSubWebhookMessage{
		Event: []byte(`{invalid json}`),
	}

	_, err := ParseEventSubEvent[StreamOnlineEvent](msg)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestEventSubWebhookHandler_VerificationWithoutHandler(t *testing.T) {
	// Test verification without a handler - should accept by default
	handler := NewEventSubWebhookHandler()

	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:   "sub123",
			Type: EventSubTypeStreamOnline,
		},
		Challenge: "test-challenge",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test-challenge" {
		t.Errorf("expected challenge response, got %s", w.Body.String())
	}
}
