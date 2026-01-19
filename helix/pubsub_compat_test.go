package helix

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestParseTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantType  string
		wantChan  string
		wantUser  string
		wantMod   string
		wantError bool
	}{
		{
			name:     "channel-bits-events-v1",
			topic:    "channel-bits-events-v1.12345",
			wantType: "channel-bits-events",
			wantChan: "12345",
		},
		{
			name:     "channel-bits-events-v2",
			topic:    "channel-bits-events-v2.67890",
			wantType: "channel-bits-events",
			wantChan: "67890",
		},
		{
			name:     "channel-bits-badge-unlocks",
			topic:    "channel-bits-badge-unlocks.12345",
			wantType: "channel-bits-badge",
			wantChan: "12345",
		},
		{
			name:     "channel-points-channel-v1",
			topic:    "channel-points-channel-v1.12345",
			wantType: "channel-points-channel",
			wantChan: "12345",
		},
		{
			name:     "channel-subscribe-events-v1",
			topic:    "channel-subscribe-events-v1.12345",
			wantType: "channel-subscribe",
			wantChan: "12345",
		},
		{
			name:     "automod-queue",
			topic:    "automod-queue.11111.22222",
			wantType: "automod-queue",
			wantMod:  "11111",
			wantChan: "22222",
		},
		{
			name:     "chat_moderator_actions",
			topic:    "chat_moderator_actions.11111.22222",
			wantType: "chat-moderator-actions",
			wantUser: "11111",
			wantChan: "22222",
		},
		{
			name:     "whispers",
			topic:    "whispers.12345",
			wantType: "whispers",
			wantUser: "12345",
		},
		{
			name:      "invalid topic",
			topic:     "invalid-topic.123",
			wantError: true,
		},
		{
			name:      "empty topic",
			topic:     "",
			wantError: true,
		},
		{
			name:      "missing channel id",
			topic:     "channel-bits-events-v1.",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseTopic(tt.topic)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if parsed.Type != tt.wantType {
				t.Errorf("Type: expected %q, got %q", tt.wantType, parsed.Type)
			}
			if parsed.ChannelID != tt.wantChan {
				t.Errorf("ChannelID: expected %q, got %q", tt.wantChan, parsed.ChannelID)
			}
			if parsed.UserID != tt.wantUser {
				t.Errorf("UserID: expected %q, got %q", tt.wantUser, parsed.UserID)
			}
			if parsed.ModeratorID != tt.wantMod {
				t.Errorf("ModeratorID: expected %q, got %q", tt.wantMod, parsed.ModeratorID)
			}
		})
	}
}

func TestTopicEventSubTypes(t *testing.T) {
	tests := []struct {
		topic    string
		expected []string
	}{
		{
			topic:    "channel-bits-events-v1.12345",
			expected: []string{EventSubTypeChannelCheer},
		},
		{
			topic:    "channel-points-channel-v1.12345",
			expected: []string{EventSubTypeChannelPointsRedemptionAdd},
		},
		{
			topic: "channel-subscribe-events-v1.12345",
			expected: []string{
				EventSubTypeChannelSubscribe,
				EventSubTypeChannelSubscriptionGift,
				EventSubTypeChannelSubscriptionMessage,
			},
		},
		{
			topic:    "whispers.12345",
			expected: []string{EventSubTypeUserWhisperMessage},
		},
		{
			topic:    "invalid-topic",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			result := TopicEventSubTypes(tt.topic)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d types, got %d", len(tt.expected), len(result))
				return
			}

			for i, typ := range tt.expected {
				if result[i] != typ {
					t.Errorf("index %d: expected %q, got %q", i, typ, result[i])
				}
			}
		})
	}
}

func TestBuildTopic(t *testing.T) {
	tests := []struct {
		name      string
		topicType string
		ids       []string
		expected  string
	}{
		{
			name:      "channel-bits-events-v1",
			topicType: "channel-bits-events-v1",
			ids:       []string{"12345"},
			expected:  "channel-bits-events-v1.12345",
		},
		{
			name:      "channel-bits-events-v2",
			topicType: "channel-bits-events-v2",
			ids:       []string{"12345"},
			expected:  "channel-bits-events-v2.12345",
		},
		{
			name:      "channel-points-channel-v1",
			topicType: "channel-points-channel-v1",
			ids:       []string{"12345"},
			expected:  "channel-points-channel-v1.12345",
		},
		{
			name:      "automod-queue",
			topicType: "automod-queue",
			ids:       []string{"11111", "22222"},
			expected:  "automod-queue.11111.22222",
		},
		{
			name:      "chat_moderator_actions",
			topicType: "chat_moderator_actions",
			ids:       []string{"11111", "22222"},
			expected:  "chat_moderator_actions.11111.22222",
		},
		{
			name:      "whispers",
			topicType: "whispers",
			ids:       []string{"12345"},
			expected:  "whispers.12345",
		},
		{
			name:      "unknown type",
			topicType: "unknown",
			ids:       []string{"12345"},
			expected:  "",
		},
		{
			name:      "missing ids",
			topicType: "channel-bits-events-v1",
			ids:       []string{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildTopic(tt.topicType, tt.ids...)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSupportedTopics(t *testing.T) {
	topics := SupportedTopics()
	if len(topics) == 0 {
		t.Error("expected at least one supported topic")
	}

	// Verify expected topics are present
	expectedPatterns := []string{
		"channel-bits-events-v1",
		"channel-points-channel-v1",
		"whispers",
	}

	for _, pattern := range expectedPatterns {
		found := false
		for _, topic := range topics {
			if strings.Contains(topic, pattern) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected pattern %q to be in supported topics", pattern)
		}
	}
}

func TestNewPubSubClient(t *testing.T) {
	client := &Client{} // Mock client
	pubsub := NewPubSubClient(client)

	if pubsub == nil {
		t.Fatal("expected non-nil PubSubClient")
	}
	if pubsub.client != client {
		t.Error("client not set correctly")
	}
	if pubsub.topics == nil {
		t.Error("topics map not initialized")
	}
	if pubsub.subToTopic == nil {
		t.Error("subToTopic map not initialized")
	}
}

func TestNewPubSubClient_WithOptions(t *testing.T) {
	client := &Client{}
	customURL := "wss://custom.example.com/ws"
	messageCalled := false
	errorCalled := false
	connectCalled := false
	reconnectCalled := false

	pubsub := NewPubSubClient(client,
		WithPubSubWSURL(customURL),
		WithPubSubMessageHandler(func(topic string, msg json.RawMessage) {
			messageCalled = true
		}),
		WithPubSubErrorHandler(func(err error) {
			errorCalled = true
		}),
		WithPubSubConnectHandler(func() {
			connectCalled = true
		}),
		WithPubSubReconnectHandler(func() {
			reconnectCalled = true
		}),
	)

	if pubsub.wsURL != customURL {
		t.Errorf("expected wsURL %q, got %q", customURL, pubsub.wsURL)
	}

	// Test handlers are set
	if pubsub.onMessage != nil {
		pubsub.onMessage("test", nil)
	}
	if pubsub.onError != nil {
		pubsub.onError(nil)
	}
	if pubsub.onConnect != nil {
		pubsub.onConnect()
	}
	if pubsub.onReconnect != nil {
		pubsub.onReconnect()
	}

	if !messageCalled || !errorCalled || !connectCalled || !reconnectCalled {
		t.Error("not all handlers were called")
	}
}

func TestPubSubClient_IsConnected(t *testing.T) {
	client := &Client{}
	pubsub := NewPubSubClient(client)

	if pubsub.IsConnected() {
		t.Error("expected not connected before Connect()")
	}
}

func TestPubSubClient_Topics(t *testing.T) {
	client := &Client{}
	pubsub := NewPubSubClient(client)

	topics := pubsub.Topics()
	if len(topics) != 0 {
		t.Errorf("expected 0 topics, got %d", len(topics))
	}
}

func TestPubSubClient_ListenNotConnected(t *testing.T) {
	client := &Client{}
	pubsub := NewPubSubClient(client)

	err := pubsub.Listen(context.Background(), "channel-bits-events-v1.12345")
	if err != ErrPubSubNotConnected {
		t.Errorf("expected ErrPubSubNotConnected, got %v", err)
	}
}

// mockPubSubWSServer creates a test WebSocket server for PubSub testing
type mockPubSubWSServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	conn     *websocket.Conn
	mu       sync.Mutex
}

func newMockPubSubWSServer(handler func(*websocket.Conn)) *mockPubSubWSServer {
	mock := &mockPubSubWSServer{
		upgrader: websocket.Upgrader{},
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := mock.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		mock.mu.Lock()
		mock.conn = conn
		mock.mu.Unlock()
		handler(conn)
	}))

	return mock
}

func (m *mockPubSubWSServer) URL() string {
	return "ws" + strings.TrimPrefix(m.server.URL, "http")
}

func (m *mockPubSubWSServer) Close() {
	m.mu.Lock()
	if m.conn != nil {
		_ = m.conn.Close()
	}
	m.mu.Unlock()
	m.server.Close()
}

// mockHelixServer creates a mock Helix API server for subscription management
type mockHelixServer struct {
	server        *httptest.Server
	subscriptions map[string]*EventSubSubscription
	mu            sync.Mutex
	subIDCounter  int
}

func newMockHelixServer() *mockHelixServer {
	mock := &mockHelixServer{
		subscriptions: make(map[string]*EventSubSubscription),
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Create subscription
			var params CreateEventSubSubscriptionParams
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			mock.mu.Lock()
			mock.subIDCounter++
			subID := "sub-" + string(rune('0'+mock.subIDCounter))
			sub := &EventSubSubscription{
				ID:        subID,
				Status:    EventSubStatusEnabled,
				Type:      params.Type,
				Version:   params.Version,
				Condition: params.Condition,
				Transport: EventSubTransport{
					Method:    params.Transport.Method,
					SessionID: params.Transport.SessionID,
				},
			}
			mock.subscriptions[subID] = sub
			mock.mu.Unlock()

			resp := EventSubResponse{
				Data:  []EventSubSubscription{*sub},
				Total: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)

		case http.MethodDelete:
			// Delete subscription
			subID := r.URL.Query().Get("id")
			mock.mu.Lock()
			delete(mock.subscriptions, subID)
			mock.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	return mock
}

func (m *mockHelixServer) Close() {
	m.server.Close()
}

func TestPubSubClient_Connect(t *testing.T) {
	sessionID := "pubsub-session-123"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		// Send welcome message
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(100 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	connectCalled := false
	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubConnectHandler(func() {
			connectCalled = true
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !pubsub.IsConnected() {
		t.Error("expected to be connected")
	}

	if pubsub.SessionID() != sessionID {
		t.Errorf("expected session ID %q, got %q", sessionID, pubsub.SessionID())
	}

	if !connectCalled {
		t.Error("connect handler was not called")
	}

	// Test Close
	err = pubsub.Close(ctx)
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestPubSubClient_Connect_NoDeadlock verifies that the onConnect handler can
// call back into PubSubClient methods without deadlocking. This tests the fix
// where Connect() releases the lock before calling the onConnect callback.
func TestPubSubClient_Connect_NoDeadlock(t *testing.T) {
	sessionID := "pubsub-session-nodeadlock"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		// Send welcome message
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(100 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	callbackCompleted := make(chan struct{})
	var callbackSessionID string
	var pubsubPtr *PubSubClient

	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubConnectHandler(func() {
			// This would deadlock if the lock wasn't released before calling onConnect
			// Calling methods that acquire the lock from within the callback
			_ = pubsubPtr.IsConnected()
			_ = pubsubPtr.Topics()
			callbackSessionID = pubsubPtr.SessionID()
			close(callbackCompleted)
		}),
	)
	pubsubPtr = pubsub

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Wait for callback with timeout (would hang forever if deadlocked)
	select {
	case <-callbackCompleted:
		// Success - callback completed without deadlocking
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: onConnect handler appears to be deadlocked")
	}

	if callbackSessionID != sessionID {
		t.Errorf("callback got session ID %q, want %q", callbackSessionID, sessionID)
	}
}

func TestPubSubClient_ListenAndUnlisten(t *testing.T) {
	sessionID := "pubsub-session-listen"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(500 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	pubsub := NewPubSubClient(helixClient, WithPubSubWSURL(wsMock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Listen to a topic
	topic := "channel-points-channel-v1.12345"
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Verify topic is tracked
	topics := pubsub.Topics()
	if len(topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(topics))
	}
	if topics[0] != topic {
		t.Errorf("expected topic %q, got %q", topic, topics[0])
	}

	// Listen again should be idempotent
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Second Listen failed: %v", err)
	}
	if len(pubsub.Topics()) != 1 {
		t.Error("duplicate listen should not add topic")
	}

	// Unlisten
	err = pubsub.Unlisten(ctx, topic)
	if err != nil {
		t.Fatalf("Unlisten failed: %v", err)
	}

	if len(pubsub.Topics()) != 0 {
		t.Error("expected 0 topics after unlisten")
	}

	// Unlisten again should be safe
	err = pubsub.Unlisten(ctx, topic)
	if err != nil {
		t.Fatalf("Second Unlisten failed: %v", err)
	}
}

func TestPubSubClient_ListenInvalidTopic(t *testing.T) {
	sessionID := "pubsub-session-invalid"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(100 * time.Millisecond)
	})
	defer wsMock.Close()

	helixClient := NewClient("test-client-id", nil)
	pubsub := NewPubSubClient(helixClient, WithPubSubWSURL(wsMock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Listen to invalid topic
	err = pubsub.Listen(ctx, "invalid-topic-format")
	if err == nil {
		t.Error("expected error for invalid topic")
	}
}

func TestPubSubClient_MultipleSubscriptionsPerTopic(t *testing.T) {
	sessionID := "pubsub-session-multi"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(500 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	pubsub := NewPubSubClient(helixClient, WithPubSubWSURL(wsMock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Listen to subscribe topic (maps to 3 EventSub types)
	topic := "channel-subscribe-events-v1.12345"
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Verify 3 subscriptions were created
	helixMock.mu.Lock()
	subCount := len(helixMock.subscriptions)
	helixMock.mu.Unlock()

	if subCount != 3 {
		t.Errorf("expected 3 subscriptions, got %d", subCount)
	}
}

func TestPubSubClient_NotificationRouting(t *testing.T) {
	sessionID := "pubsub-session-notif"
	var receivedTopic string
	var receivedMessage json.RawMessage
	notifReceived := make(chan struct{})

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)

		// Wait a bit then send notification
		time.Sleep(100 * time.Millisecond)

		// Send notification with subscription ID that matches what we'll create
		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "notif-1",
				MessageType:      WSMessageTypeNotification,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:   "sub-1", // First subscription ID from mock server
					Type: EventSubTypeChannelPointsRedemptionAdd,
				},
				Event: mustMarshal(map[string]string{
					"user_name": "testuser",
					"reward":    "Test Reward",
				}),
			}),
		}
		_ = conn.WriteJSON(notification)

		time.Sleep(200 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubMessageHandler(func(topic string, msg json.RawMessage) {
			receivedTopic = topic
			receivedMessage = msg
			close(notifReceived)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	topic := "channel-points-channel-v1.12345"
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Wait for notification
	select {
	case <-notifReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for notification")
	}

	if receivedTopic != topic {
		t.Errorf("expected topic %q, got %q", topic, receivedTopic)
	}

	if receivedMessage == nil {
		t.Error("expected message to be non-nil")
	}

	// Verify message envelope
	var envelope PubSubMessage
	if err := json.Unmarshal(receivedMessage, &envelope); err != nil {
		t.Fatalf("failed to parse message envelope: %v", err)
	}

	if envelope.Type != EventSubTypeChannelPointsRedemptionAdd {
		t.Errorf("expected type %q, got %q", EventSubTypeChannelPointsRedemptionAdd, envelope.Type)
	}
}

func TestPubSubClient_Revocation(t *testing.T) {
	sessionID := "pubsub-session-revoke"
	errorReceived := make(chan error, 1)

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)

		// Wait for subscription to be created
		time.Sleep(150 * time.Millisecond)

		// Send revocation message
		revocation := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "revoke-1",
				MessageType:      WSMessageTypeRevocation,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:     "sub-1",
					Type:   EventSubTypeChannelPointsRedemptionAdd,
					Status: "authorization_revoked",
				},
			}),
		}
		_ = conn.WriteJSON(revocation)

		time.Sleep(200 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Listen to a topic
	topic := "channel-points-channel-v1.12345"
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Wait for revocation error
	select {
	case err := <-errorReceived:
		if err == nil {
			t.Error("expected error from revocation")
		}
		// Verify error contains expected text
		if !strings.Contains(err.Error(), "subscription revoked") {
			t.Errorf("expected revocation error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for revocation error")
	}

	// Verify topic was removed
	time.Sleep(100 * time.Millisecond)
	topics := pubsub.Topics()
	if len(topics) != 0 {
		t.Errorf("expected 0 topics after revocation, got %d", len(topics))
	}
}

func TestPubSubClient_Reconnect(t *testing.T) {
	sessionID := "pubsub-session-reconnect"
	newSessionID := "pubsub-session-new"
	reconnectCalled := make(chan struct{})

	// First server for initial connection
	firstServer := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(500 * time.Millisecond)
	})
	defer firstServer.Close()

	// Second server for reconnection
	secondServer := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-2",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      newSessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(500 * time.Millisecond)
	})
	defer secondServer.Close()

	helixClient := NewClient("test-client-id", nil)

	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(firstServer.URL()),
		WithPubSubReconnectHandler(func() {
			close(reconnectCalled)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Manually trigger reconnect handler (simulating server request)
	pubsub.handleReconnect(secondServer.URL())

	// Wait for reconnect handler to be called
	select {
	case <-reconnectCalled:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for reconnect handler")
	}

	// Give time for session ID to be updated
	time.Sleep(100 * time.Millisecond)

	if pubsub.SessionID() != newSessionID {
		t.Errorf("expected new session ID %q, got %q", newSessionID, pubsub.SessionID())
	}
}

func TestPubSubClient_ReconnectError(t *testing.T) {
	sessionID := "pubsub-session-reconnect-err"
	errorReceived := make(chan error, 1)

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(500 * time.Millisecond)
	})
	defer wsMock.Close()

	helixClient := NewClient("test-client-id", nil)

	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Trigger reconnect to invalid URL
	pubsub.handleReconnect("ws://invalid-url-that-does-not-exist:12345")

	// Wait for error
	select {
	case err := <-errorReceived:
		if err == nil {
			t.Error("expected error from failed reconnect")
		}
		if !strings.Contains(err.Error(), "reconnecting") {
			t.Errorf("expected reconnecting error, got: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for reconnect error")
	}
}

func TestPubSubClient_HandleErrorNoHandler(t *testing.T) {
	// Test that handleError doesn't panic when no error handler is set
	client := &Client{}
	pubsub := NewPubSubClient(client)

	// This should not panic
	pubsub.handleError(errors.New("test error"))
}

func TestPubSubClient_NotificationUnknownSubscription(t *testing.T) {
	sessionID := "pubsub-session-unknown"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)

		// Send notification for unknown subscription
		time.Sleep(100 * time.Millisecond)
		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "notif-unknown",
				MessageType:      WSMessageTypeNotification,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:   "unknown-sub-id",
					Type: EventSubTypeChannelCheer,
				},
				Event: mustMarshal(map[string]string{"test": "data"}),
			}),
		}
		_ = conn.WriteJSON(notification)

		time.Sleep(200 * time.Millisecond)
	})
	defer wsMock.Close()

	messageCalled := false
	helixClient := NewClient("test-client-id", nil)
	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubMessageHandler(func(topic string, msg json.RawMessage) {
			messageCalled = true
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Wait for notification to be processed
	time.Sleep(500 * time.Millisecond)

	// Message handler should NOT have been called for unknown subscription
	if messageCalled {
		t.Error("message handler should not be called for unknown subscription")
	}
}

func TestPubSubClient_NotificationNoHandler(t *testing.T) {
	sessionID := "pubsub-session-no-handler"

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)

		time.Sleep(100 * time.Millisecond)

		// Send notification
		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "notif-1",
				MessageType:      WSMessageTypeNotification,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:   "sub-1",
					Type: EventSubTypeChannelPointsRedemptionAdd,
				},
				Event: mustMarshal(map[string]string{"test": "data"}),
			}),
		}
		_ = conn.WriteJSON(notification)

		time.Sleep(200 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	// No message handler set
	pubsub := NewPubSubClient(helixClient, WithPubSubWSURL(wsMock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	topic := "channel-points-channel-v1.12345"
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Wait for notification to be processed - should not panic
	time.Sleep(500 * time.Millisecond)
}

func TestPubSubClient_RevocationCleanupMultipleSubscriptions(t *testing.T) {
	sessionID := "pubsub-session-revoke-multi"
	errorCount := 0
	var errorMu sync.Mutex

	wsMock := newMockPubSubWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "welcome-1",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 10,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)

		// Wait for subscriptions to be created (3 for subscribe events)
		time.Sleep(200 * time.Millisecond)

		// Revoke only one of the three subscriptions
		revocation := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "revoke-1",
				MessageType:      WSMessageTypeRevocation,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:     "sub-1",
					Type:   EventSubTypeChannelSubscribe,
					Status: "authorization_revoked",
				},
			}),
		}
		_ = conn.WriteJSON(revocation)

		time.Sleep(300 * time.Millisecond)
	})
	defer wsMock.Close()

	helixMock := newMockHelixServer()
	defer helixMock.Close()

	helixClient := NewClient("test-client-id", nil)
	helixClient.baseURL = helixMock.server.URL

	pubsub := NewPubSubClient(helixClient,
		WithPubSubWSURL(wsMock.URL()),
		WithPubSubErrorHandler(func(err error) {
			errorMu.Lock()
			errorCount++
			errorMu.Unlock()
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pubsub.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = pubsub.Close(ctx) }()

	// Listen to subscribe events (maps to 3 EventSub types)
	topic := "channel-subscribe-events-v1.12345"
	err = pubsub.Listen(ctx, topic)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	// Verify 3 subscriptions were created
	helixMock.mu.Lock()
	initialSubCount := len(helixMock.subscriptions)
	helixMock.mu.Unlock()

	if initialSubCount != 3 {
		t.Errorf("expected 3 subscriptions, got %d", initialSubCount)
	}

	// Wait for revocation to be processed
	time.Sleep(600 * time.Millisecond)

	// Topic should still exist (2 subs remaining)
	topics := pubsub.Topics()
	if len(topics) != 1 {
		t.Errorf("expected 1 topic (with 2 remaining subs), got %d", len(topics))
	}

	// Should have received exactly one error
	errorMu.Lock()
	finalErrorCount := errorCount
	errorMu.Unlock()

	if finalErrorCount != 1 {
		t.Errorf("expected 1 error from revocation, got %d", finalErrorCount)
	}
}

func TestPubSubClient_UnlistenNotConnected(t *testing.T) {
	client := &Client{}
	pubsub := NewPubSubClient(client)

	// Unlisten when not connected should be safe (no-op)
	err := pubsub.Unlisten(context.Background(), "channel-bits-events-v1.12345")
	if err != nil {
		t.Errorf("expected no error for unlisten when not listening, got %v", err)
	}
}

func TestPubSubClient_CloseNotConnected(t *testing.T) {
	client := &Client{}
	pubsub := NewPubSubClient(client)

	// Close when not connected should be safe
	err := pubsub.Close(context.Background())
	if err != nil {
		t.Errorf("expected no error for close when not connected, got %v", err)
	}
}

func TestNewPubSubClient_NilClient(t *testing.T) {
	pubsub := NewPubSubClient(nil)
	if pubsub != nil {
		t.Error("expected nil PubSubClient when helixClient is nil")
	}
}

func TestTopicEventSubTypes_AllTopics(t *testing.T) {
	tests := []struct {
		topic    string
		expected []string
	}{
		{
			topic:    "channel-bits-badge-unlocks.12345",
			expected: []string{EventSubTypeChannelChatNotification},
		},
		{
			topic:    "automod-queue.11111.22222",
			expected: []string{EventSubTypeAutomodMessageHold},
		},
		{
			topic:    "chat_moderator_actions.11111.22222",
			expected: []string{EventSubTypeChannelModerate},
		},
	}

	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			result := TopicEventSubTypes(tt.topic)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d types, got %d", len(tt.expected), len(result))
				return
			}

			for i, typ := range tt.expected {
				if result[i] != typ {
					t.Errorf("index %d: expected %q, got %q", i, typ, result[i])
				}
			}
		})
	}
}

func TestPubSubClient_HandleErrorWithHandler(t *testing.T) {
	client := &Client{}
	var receivedErr error
	pubsub := NewPubSubClient(client, WithPubSubErrorHandler(func(err error) {
		receivedErr = err
	}))

	testErr := errors.New("test error")
	pubsub.handleError(testErr)

	if receivedErr != testErr {
		t.Errorf("expected error %v, got %v", testErr, receivedErr)
	}
}
