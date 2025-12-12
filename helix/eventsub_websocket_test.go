package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockWSServer creates a test WebSocket server
type mockWSServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	conn     *websocket.Conn
	mu       sync.Mutex
}

func newMockWSServer(handler func(*websocket.Conn)) *mockWSServer {
	mock := &mockWSServer{
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

func (m *mockWSServer) URL() string {
	return "ws" + strings.TrimPrefix(m.server.URL, "http")
}

func (m *mockWSServer) Close() {
	m.mu.Lock()
	if m.conn != nil {
		_ = m.conn.Close()
	}
	m.mu.Unlock()
	m.server.Close()
}

func TestNewEventSubWebSocketClient(t *testing.T) {
	client := NewEventSubWebSocketClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.url != EventSubWebSocketURL {
		t.Errorf("expected default URL %s, got %s", EventSubWebSocketURL, client.url)
	}
}

func TestEventSubWebSocketClient_WithOptions(t *testing.T) {
	customURL := "wss://custom.example.com/ws"
	welcomeCalled := false
	notifCalled := false
	revokeCalled := false
	reconnectCalled := false
	errorCalled := false
	keepaliveCalled := false

	client := NewEventSubWebSocketClient(
		WithWSURL(customURL),
		WithWSWelcomeHandler(func(*WebSocketSession) { welcomeCalled = true }),
		WithWSNotificationHandler(func(*EventSubSubscription, json.RawMessage) { notifCalled = true }),
		WithWSRevocationHandler(func(*EventSubSubscription) { revokeCalled = true }),
		WithWSReconnectHandler(func(string) { reconnectCalled = true }),
		WithWSErrorHandler(func(error) { errorCalled = true }),
		WithWSKeepaliveHandler(func() { keepaliveCalled = true }),
	)

	if client.url != customURL {
		t.Errorf("expected URL %s, got %s", customURL, client.url)
	}

	// Test handlers are set (call them to verify)
	if client.onWelcome != nil {
		client.onWelcome(nil)
	}
	if client.onNotification != nil {
		client.onNotification(nil, nil)
	}
	if client.onRevocation != nil {
		client.onRevocation(nil)
	}
	if client.onReconnect != nil {
		client.onReconnect("")
	}
	if client.onError != nil {
		client.onError(nil)
	}
	if client.onKeepalive != nil {
		client.onKeepalive()
	}

	if !welcomeCalled || !notifCalled || !revokeCalled || !reconnectCalled || !errorCalled || !keepaliveCalled {
		t.Error("not all handlers were called")
	}
}

func TestEventSubWebSocketClient_Connect(t *testing.T) {
	sessionID := "test-session-123"

	mock := newMockWSServer(func(conn *websocket.Conn) {
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

		// Keep connection alive for a bit
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	var welcomeSession *WebSocketSession
	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSWelcomeHandler(func(session *WebSocketSession) {
			welcomeSession = session
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gotSessionID, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if gotSessionID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, gotSessionID)
	}

	if !client.IsConnected() {
		t.Error("expected client to be connected")
	}

	if client.SessionID() != sessionID {
		t.Errorf("expected SessionID() to return %s, got %s", sessionID, client.SessionID())
	}

	if welcomeSession == nil {
		t.Error("welcome handler was not called")
	} else if welcomeSession.ID != sessionID {
		t.Errorf("welcome session ID mismatch: %s vs %s", welcomeSession.ID, sessionID)
	}

	// Close client
	if err := client.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("expected client to be disconnected after Close")
	}
}

func TestEventSubWebSocketClient_HandleNotification(t *testing.T) {
	sessionID := "test-session-456"
	var receivedSub *EventSubSubscription
	var receivedEvent json.RawMessage
	notifReceived := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
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

		// Give time for connection setup
		time.Sleep(50 * time.Millisecond)

		// Send notification
		eventData := mustMarshal(map[string]interface{}{
			"broadcaster_user_id":    "12345",
			"broadcaster_user_login": "testuser",
			"broadcaster_user_name":  "TestUser",
			"type":                   "live",
			"started_at":             time.Now().Format(time.RFC3339),
		})

		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:           "notif-1",
				MessageType:         WSMessageTypeNotification,
				MessageTimestamp:    time.Now(),
				SubscriptionType:    EventSubTypeStreamOnline,
				SubscriptionVersion: "1",
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:      "sub-123",
					Type:    EventSubTypeStreamOnline,
					Status:  "enabled",
					Version: "1",
				},
				Event: eventData,
			}),
		}
		_ = conn.WriteJSON(notification)

		// Keep alive
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSNotificationHandler(func(sub *EventSubSubscription, event json.RawMessage) {
			receivedSub = sub
			receivedEvent = event
			close(notifReceived)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for notification
	select {
	case <-notifReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for notification")
	}

	if receivedSub == nil {
		t.Fatal("no subscription received")
	}
	if receivedSub.Type != EventSubTypeStreamOnline {
		t.Errorf("expected subscription type %s, got %s", EventSubTypeStreamOnline, receivedSub.Type)
	}
	if receivedEvent == nil {
		t.Fatal("no event received")
	}
}

func TestEventSubWebSocketClient_HandleKeepalive(t *testing.T) {
	sessionID := "test-session-789"
	keepaliveCalled := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Send welcome
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

		time.Sleep(50 * time.Millisecond)

		// Send keepalive
		keepalive := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "keepalive-1",
				MessageType:      WSMessageTypeKeepalive,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(struct{}{}),
		}
		_ = conn.WriteJSON(keepalive)

		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSKeepaliveHandler(func() {
			select {
			case <-keepaliveCalled:
			default:
				close(keepaliveCalled)
			}
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case <-keepaliveCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("keepalive handler was not called")
	}
}

func TestEventSubWebSocketClient_HandleReconnect(t *testing.T) {
	sessionID := "test-session-reconnect"
	reconnectURL := "wss://new-server.twitch.tv/ws"
	var receivedReconnectURL string
	reconnectCalled := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Send welcome
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

		time.Sleep(50 * time.Millisecond)

		// Send reconnect
		reconnect := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "reconnect-1",
				MessageType:      WSMessageTypeReconnect,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketReconnectPayload{
				Session: WebSocketSession{
					ID:           sessionID,
					Status:       "reconnecting",
					ReconnectURL: reconnectURL,
				},
			}),
		}
		_ = conn.WriteJSON(reconnect)

		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSReconnectHandler(func(url string) {
			receivedReconnectURL = url
			close(reconnectCalled)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case <-reconnectCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("reconnect handler was not called")
	}

	if receivedReconnectURL != reconnectURL {
		t.Errorf("expected reconnect URL %s, got %s", reconnectURL, receivedReconnectURL)
	}
}

func TestEventSubWebSocketClient_HandleRevocation(t *testing.T) {
	sessionID := "test-session-revoke"
	var receivedSub *EventSubSubscription
	revokeCalled := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Send welcome
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

		time.Sleep(50 * time.Millisecond)

		// Send revocation
		revocation := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "revoke-1",
				MessageType:      WSMessageTypeRevocation,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:     "sub-revoked",
					Type:   EventSubTypeStreamOnline,
					Status: "authorization_revoked",
				},
			}),
		}
		_ = conn.WriteJSON(revocation)

		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSRevocationHandler(func(sub *EventSubSubscription) {
			receivedSub = sub
			close(revokeCalled)
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case <-revokeCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("revocation handler was not called")
	}

	if receivedSub == nil {
		t.Fatal("no subscription received")
	}
	if receivedSub.Status != "authorization_revoked" {
		t.Errorf("expected status authorization_revoked, got %s", receivedSub.Status)
	}
}

func TestParseWSEvent(t *testing.T) {
	eventData := json.RawMessage(`{
		"id": "stream123",
		"broadcaster_user_id": "12345",
		"broadcaster_user_login": "testuser",
		"broadcaster_user_name": "TestUser",
		"type": "live",
		"started_at": "2024-01-01T00:00:00Z"
	}`)

	event, err := ParseWSEvent[StreamOnlineEvent](eventData)
	if err != nil {
		t.Fatalf("ParseWSEvent failed: %v", err)
	}

	if event.ID != "stream123" {
		t.Errorf("expected ID stream123, got %s", event.ID)
	}
	if event.BroadcasterUserID != "12345" {
		t.Errorf("expected broadcaster ID 12345, got %s", event.BroadcasterUserID)
	}
	if event.Type != "live" {
		t.Errorf("expected type live, got %s", event.Type)
	}
}

func TestParseWSEvent_InvalidJSON(t *testing.T) {
	eventData := json.RawMessage(`{invalid json}`)

	_, err := ParseWSEvent[StreamOnlineEvent](eventData)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestEventSubWebSocketClient_ConnectAlreadyConnected(t *testing.T) {
	sessionID := "test-session-already"

	mock := newMockWSServer(func(conn *websocket.Conn) {
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
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("first Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Second connect should return existing session ID
	gotSessionID, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("second Connect failed: %v", err)
	}
	if gotSessionID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, gotSessionID)
	}
}

// mustMarshal marshals v to JSON and panics on error.
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
