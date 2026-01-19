package helix

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Official Twitch EventSub webhook test values from:
// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/
const (
	// Official Twitch example challenge
	twitchExampleChallenge = "pogchamp-kappa-360noscope-vohiyo"
	// Official Twitch example subscription ID
	twitchExampleSubscriptionID = "f1c2a387-161a-49f9-a165-0f21d7a4e1c4"
	// Official Twitch example broadcaster user ID
	twitchExampleBroadcasterUserID = "12826"
	// Official Twitch example user ID (follower)
	twitchExampleUserID = "1337"
	// Official Twitch example user login
	twitchExampleUserLogin = "awesome_user"
	// Official Twitch example user name
	twitchExampleUserName = "Awesome_User"
	// Official Twitch example broadcaster login
	twitchExampleBroadcasterLogin = "twitch"
	// Official Twitch example broadcaster name
	twitchExampleBroadcasterName = "Twitch"
	// Official Twitch example created_at timestamp
	twitchExampleCreatedAt = "2019-11-16T10:11:12.634234626Z"
	// Official Twitch example followed_at timestamp
	twitchExampleFollowedAt = "2020-07-15T18:16:11.17106713Z"
	// Official Twitch example webhook callback URL
	twitchExampleCallbackURL = "https://example.com/webhooks/callback"
)

func TestEventSubWebhookHandler_Verification(t *testing.T) {
	// Using official Twitch example secret format (64-char hex from crypto.randomBytes(32))
	secret := "5f1a6e7cd2e7137ccf9e15b2f43fe63949eb84b1db83c1d5a867dc93429de4e4"

	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithVerificationHandler(func(msg *EventSubWebhookMessage) bool {
			receivedMsg = msg
			return true
		}),
	)

	// Create verification payload using official Twitch example values
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/
	createdAt, _ := time.Parse(time.RFC3339Nano, twitchExampleCreatedAt)
	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:      twitchExampleSubscriptionID,
			Type:    "channel.follow",
			Version: "1",
			Status:  "webhook_callback_verification_pending",
			Condition: map[string]string{
				"broadcaster_user_id": twitchExampleBroadcasterUserID,
			},
			Transport: EventSubTransport{
				Method:   "webhook",
				Callback: twitchExampleCallbackURL,
			},
			CreatedAt: createdAt,
			Cost:      1,
		},
		Challenge: twitchExampleChallenge,
	}
	body, _ := json.Marshal(payload)

	// Create request with proper headers
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	messageID := "e76c6bd4-55c9-4987-8304-da1588d8988b"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)
	req.Header.Set(EventSubHeaderSubscriptionType, "channel.follow")
	req.Header.Set(EventSubHeaderSubscriptionVersion, "1")

	// Sign the message (official Twitch algorithm: HMAC-SHA256 of message_id + timestamp + body)
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	// Execute request
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response - Twitch expects the challenge echoed back
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != twitchExampleChallenge {
		t.Errorf("expected challenge %s, got %s", twitchExampleChallenge, w.Body.String())
	}
	if receivedMsg == nil {
		t.Error("verification handler was not called")
	}
	if receivedMsg.Challenge != twitchExampleChallenge {
		t.Errorf("expected challenge in message, got %s", receivedMsg.Challenge)
	}
	if receivedMsg.Subscription.ID != twitchExampleSubscriptionID {
		t.Errorf("expected subscription ID %s, got %s", twitchExampleSubscriptionID, receivedMsg.Subscription.ID)
	}
}

func TestEventSubWebhookHandler_Notification(t *testing.T) {
	// Using official Twitch example secret format
	secret := "5f1a6e7cd2e7137ccf9e15b2f43fe63949eb84b1db83c1d5a867dc93429de4e4"

	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
		WithNotificationHandler(func(msg *EventSubWebhookMessage) {
			receivedMsg = msg
		}),
	)

	// Create notification payload using official Twitch example values
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/
	// Using channel.follow event structure from Twitch docs
	followEvent := map[string]any{
		"user_id":               twitchExampleUserID,
		"user_login":            twitchExampleUserLogin,
		"user_name":             twitchExampleUserName,
		"broadcaster_user_id":   twitchExampleBroadcasterUserID,
		"broadcaster_user_login": twitchExampleBroadcasterLogin,
		"broadcaster_user_name": twitchExampleBroadcasterName,
		"followed_at":           twitchExampleFollowedAt,
	}
	eventJSON, _ := json.Marshal(followEvent)

	createdAt, _ := time.Parse(time.RFC3339Nano, twitchExampleCreatedAt)
	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:      twitchExampleSubscriptionID,
			Type:    "channel.follow",
			Version: "1",
			Status:  "enabled",
			Condition: map[string]string{
				"broadcaster_user_id": twitchExampleBroadcasterUserID,
			},
			Transport: EventSubTransport{
				Method:   "webhook",
				Callback: twitchExampleCallbackURL,
			},
			CreatedAt: createdAt,
			Cost:      1,
		},
		Event: eventJSON,
	}
	body, _ := json.Marshal(payload)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	messageID := "befa7b53-d79d-478f-86b9-120f112b044e"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)
	req.Header.Set(EventSubHeaderSubscriptionType, "channel.follow")
	req.Header.Set(EventSubHeaderSubscriptionVersion, "1")

	// Sign the message (official Twitch HMAC-SHA256 algorithm)
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	// Execute request
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify response - Twitch expects 2xx for successful notification handling
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if receivedMsg == nil {
		t.Fatal("notification handler was not called")
	}
	if receivedMsg.SubscriptionType != "channel.follow" {
		t.Errorf("expected subscription type channel.follow, got %s", receivedMsg.SubscriptionType)
	}
	if receivedMsg.Subscription.ID != twitchExampleSubscriptionID {
		t.Errorf("expected subscription ID %s, got %s", twitchExampleSubscriptionID, receivedMsg.Subscription.ID)
	}

	// Parse the event data
	var parsedEvent map[string]any
	if err := json.Unmarshal(receivedMsg.Event, &parsedEvent); err != nil {
		t.Fatalf("failed to parse event: %v", err)
	}
	if parsedEvent["user_id"] != twitchExampleUserID {
		t.Errorf("expected user_id %s, got %v", twitchExampleUserID, parsedEvent["user_id"])
	}
	if parsedEvent["broadcaster_user_id"] != twitchExampleBroadcasterUserID {
		t.Errorf("expected broadcaster_user_id %s, got %v", twitchExampleBroadcasterUserID, parsedEvent["broadcaster_user_id"])
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

func TestEventSubWebhookHandler_FutureTimestamp(t *testing.T) {
	secret := "test-secret"
	handler := NewEventSubWebhookHandler(
		WithWebhookSecret(secret),
	)

	body := []byte(`{"subscription":{},"challenge":"test"}`)
	messageID := "msg-123"
	// Timestamp 5 minutes in the future (beyond 1-minute tolerance)
	timestamp := time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, messageID)
	req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeVerification)

	// Sign with future timestamp
	message := messageID + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	req.Header.Set(EventSubHeaderMessageSignature, signature)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for future timestamp, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("future")) {
		t.Errorf("expected error message about future timestamp, got %s", w.Body.String())
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

func TestMessageDeduplicator_MaxSizeEnforcement(t *testing.T) {
	// Test that maxSize is enforced even when entries are not expired
	dedup := NewMessageDeduplicator(time.Hour, 3) // 1 hour maxAge, maxSize 3

	// Fill up to capacity
	dedup.IsDuplicate("msg-1")
	dedup.IsDuplicate("msg-2")
	dedup.IsDuplicate("msg-3")

	// Check that we have exactly 3 entries (at capacity)
	if len(dedup.seen) != 3 {
		t.Errorf("expected 3 entries, got %d", len(dedup.seen))
	}

	// Verify msg-1, msg-2, msg-3 are in the map
	if _, ok := dedup.seen["msg-1"]; !ok {
		t.Error("msg-1 should be in the map")
	}
	if _, ok := dedup.seen["msg-2"]; !ok {
		t.Error("msg-2 should be in the map")
	}
	if _, ok := dedup.seen["msg-3"]; !ok {
		t.Error("msg-3 should be in the map")
	}

	// Add another message - should evict the oldest (msg-1)
	dedup.IsDuplicate("msg-4")

	// Should still have exactly 3 entries (maxSize enforced)
	if len(dedup.seen) != 3 {
		t.Errorf("expected 3 entries after eviction, got %d", len(dedup.seen))
	}

	// msg-1 should have been evicted as the oldest
	if _, ok := dedup.seen["msg-1"]; ok {
		t.Error("msg-1 should have been evicted")
	}

	// msg-2, msg-3, msg-4 should still be in the map
	if _, ok := dedup.seen["msg-2"]; !ok {
		t.Error("msg-2 should still be in the map")
	}
	if _, ok := dedup.seen["msg-3"]; !ok {
		t.Error("msg-3 should still be in the map")
	}
	if _, ok := dedup.seen["msg-4"]; !ok {
		t.Error("msg-4 should be in the map")
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

	t.Run("future timestamp", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler should not be called for future timestamp")
		})

		handler := middleware(nextHandler)

		body := []byte(`{"subscription":{}}`)
		messageID := "msg-123"
		// Timestamp 5 minutes in the future (beyond 1-minute tolerance)
		timestamp := time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(EventSubHeaderMessageID, messageID)
		req.Header.Set(EventSubHeaderMessageTimestamp, timestamp)

		// Create signature with future timestamp
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
		if !bytes.Contains(w.Body.Bytes(), []byte("future")) {
			t.Errorf("expected error message about future timestamp, got %s", w.Body.String())
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

func TestParseEventSubEvent_Success(t *testing.T) {
	// Test successful parsing using official Twitch example event structure
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/
	followEvent := map[string]string{
		"user_id":                twitchExampleUserID,
		"user_login":             twitchExampleUserLogin,
		"user_name":              twitchExampleUserName,
		"broadcaster_user_id":    twitchExampleBroadcasterUserID,
		"broadcaster_user_login": twitchExampleBroadcasterLogin,
		"broadcaster_user_name":  twitchExampleBroadcasterName,
		"followed_at":            twitchExampleFollowedAt,
	}
	eventJSON, _ := json.Marshal(followEvent)

	msg := &EventSubWebhookMessage{
		Event: eventJSON,
	}

	// Parse into a struct type matching Twitch's channel.follow event
	type FollowEvent struct {
		UserID               string `json:"user_id"`
		UserLogin            string `json:"user_login"`
		UserName             string `json:"user_name"`
		BroadcasterUserID    string `json:"broadcaster_user_id"`
		BroadcasterUserLogin string `json:"broadcaster_user_login"`
		BroadcasterUserName  string `json:"broadcaster_user_name"`
		FollowedAt           string `json:"followed_at"`
	}

	event, err := ParseEventSubEvent[FollowEvent](msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.UserID != twitchExampleUserID {
		t.Errorf("expected user_id %s, got %s", twitchExampleUserID, event.UserID)
	}
	if event.BroadcasterUserID != twitchExampleBroadcasterUserID {
		t.Errorf("expected broadcaster_user_id %s, got %s", twitchExampleBroadcasterUserID, event.BroadcasterUserID)
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

// errorBodyReader simulates a broken request body that fails on Read
type errorBodyReader struct{}

func (e *errorBodyReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated body read error")
}

func (e *errorBodyReader) Close() error {
	return nil
}

func TestEventSubWebhookHandler_BodyReadError(t *testing.T) {
	// Test the io.ReadAll error path in ServeHTTP
	handler := NewEventSubWebhookHandler()

	req := httptest.NewRequest(http.MethodPost, "/webhook", &errorBodyReader{})
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for body read error, got %d", w.Code)
	}
}

func TestEventSubMiddleware_BodyReadError(t *testing.T) {
	// Test the io.ReadAll error path in EventSubMiddleware
	secret := "test-secret"
	maxAge := 10 * time.Minute

	middleware := EventSubMiddleware(secret, maxAge)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called when body read fails")
	})

	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/webhook", &errorBodyReader{})
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for body read error, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_BodyTooLarge(t *testing.T) {
	handler := NewEventSubWebhookHandler()

	// Create body larger than EventSubMaxBodySize (1MB)
	largeBody := make([]byte, EventSubMaxBodySize+100)
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(largeBody))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeNotification)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", w.Code)
	}
}

func TestEventSubMiddleware_BodyTooLarge(t *testing.T) {
	secret := "test-secret"
	maxAge := 10 * time.Minute

	middleware := EventSubMiddleware(secret, maxAge)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called when body is too large")
	})

	handler := middleware(nextHandler)

	// Create body larger than EventSubMaxBodySize (1MB)
	largeBody := make([]byte, EventSubMaxBodySize+100)
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(largeBody))
	req.Header.Set(EventSubHeaderMessageID, "msg-123")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageSignature, "sha256=invalid")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", w.Code)
	}
}

func TestEventSubWebhookHandler_Revocation_OfficialExample(t *testing.T) {
	// Test revocation using official Twitch example values
	// https://dev.twitch.tv/docs/eventsub/handling-webhook-events/
	var receivedMsg *EventSubWebhookMessage
	handler := NewEventSubWebhookHandler(
		WithRevocationHandler(func(msg *EventSubWebhookMessage) {
			receivedMsg = msg
		}),
	)

	createdAt, _ := time.Parse(time.RFC3339Nano, twitchExampleCreatedAt)
	payload := EventSubWebhookPayload{
		Subscription: EventSubSubscription{
			ID:      twitchExampleSubscriptionID,
			Type:    "channel.follow",
			Version: "1",
			Status:  "authorization_revoked", // Official Twitch revocation status
			Condition: map[string]string{
				"broadcaster_user_id": twitchExampleBroadcasterUserID,
			},
			Transport: EventSubTransport{
				Method:   "webhook",
				Callback: twitchExampleCallbackURL,
			},
			CreatedAt: createdAt,
			Cost:      1,
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set(EventSubHeaderMessageID, "84c1e79a-2a4b-4c13-ba0b-4312293e9308")
	req.Header.Set(EventSubHeaderMessageTimestamp, time.Now().UTC().Format(time.RFC3339))
	req.Header.Set(EventSubHeaderMessageType, EventSubMessageTypeRevocation)
	req.Header.Set(EventSubHeaderSubscriptionType, "channel.follow")
	req.Header.Set(EventSubHeaderSubscriptionVersion, "1")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if receivedMsg == nil {
		t.Error("revocation handler was not called")
	}
	if receivedMsg.Subscription.ID != twitchExampleSubscriptionID {
		t.Errorf("expected subscription ID %s, got %s", twitchExampleSubscriptionID, receivedMsg.Subscription.ID)
	}
	if receivedMsg.Subscription.Status != "authorization_revoked" {
		t.Errorf("expected status authorization_revoked, got %s", receivedMsg.Subscription.Status)
	}
}
