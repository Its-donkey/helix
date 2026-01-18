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

// Official Twitch EventSub WebSocket example values from documentation
// https://dev.twitch.tv/docs/eventsub/handling-websocket-events
const (
	twitchWSExampleWelcomeMessageID    = "96a3f3b5-5dec-4eed-908e-e11ee657416c"
	twitchWSExampleSessionID           = "AQoQILE98gtqShGmLD7AM6yJThAB"
	twitchWSExampleKeepaliveMessageID  = "84c1e79a-2a4b-4c13-ba0b-4312293e9308"
	twitchWSExampleNotifMessageID      = "befa7b53-d79d-478f-86b9-120f112b044e"
	twitchWSExampleSubscriptionID      = "f1c2a387-161a-49f9-a165-0f21d7a4e1c4"
	twitchWSExampleBroadcasterUserID   = "12826"
	twitchWSExampleUserID              = "1337"
	twitchWSExampleUserLogin           = "awesome_user"
	twitchWSExampleUserName            = "Awesome_User"
	twitchWSExampleBroadcasterLogin    = "twitch"
	twitchWSExampleBroadcasterName     = "Twitch"
	twitchWSExampleTransportSessionID  = "AQoQexAWVYKSTIu4ec_2VAxyuhAB"
	twitchWSExampleReconnectURL        = "wss://eventsub.wss.twitch.tv?..."
	twitchWSExampleWelcomeTimestamp    = "2023-07-19T14:56:51.634234626Z"
	twitchWSExampleConnectedAt         = "2023-07-19T14:56:51.616329898Z"
	twitchWSExampleNotifTimestamp      = "2022-11-16T10:11:12.464757833Z"
	twitchWSExampleFollowedAt          = "2023-07-15T18:16:11.17106713Z"
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
		defer func() { _ = conn.Close() }() // Ensure connection is closed when handler returns
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
					KeepaliveTimeoutSeconds: 1,
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
					KeepaliveTimeoutSeconds: 1,
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
					KeepaliveTimeoutSeconds: 1,
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
					KeepaliveTimeoutSeconds: 1,
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
					KeepaliveTimeoutSeconds: 1,
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
					KeepaliveTimeoutSeconds: 1,
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

func TestEventSubWebSocketClient_Reconnect(t *testing.T) {
	oldSessionID := twitchWSExampleSessionID
	newSessionID := "new-session-after-reconnect"

	// Start old server
	oldServer := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      oldSessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(200 * time.Millisecond)
	})
	defer oldServer.Close()

	// Start new server for reconnect
	newServer := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "new-welcome-id",
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      newSessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(200 * time.Millisecond)
	})
	defer newServer.Close()

	client := NewEventSubWebSocketClient(WithWSURL(oldServer.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Reconnect to new server
	gotSessionID, err := client.Reconnect(ctx, newServer.URL())
	if err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}

	if gotSessionID != newSessionID {
		t.Errorf("expected session ID %s, got %s", newSessionID, gotSessionID)
	}

	if !client.IsConnected() {
		t.Error("expected client to be connected after reconnect")
	}

	_ = client.Close()
}

func TestEventSubWebSocketClient_Reconnect_DialError(t *testing.T) {
	sessionID := twitchWSExampleSessionID

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Try to reconnect to invalid URL
	_, err = client.Reconnect(ctx, "ws://invalid.invalid:9999")
	if err == nil {
		t.Error("expected error for invalid reconnect URL")
	}
}

func TestEventSubWebSocketClient_ConnectAlreadyConnecting(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	connectStarted := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Signal that connection started
		close(connectStarted)
		// Wait a bit before sending welcome
		time.Sleep(100 * time.Millisecond)

		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))
	ctx := context.Background()

	// Start first connection in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = client.Connect(ctx)
	}()

	// Wait for connection to start, then try second connect
	<-connectStarted
	_, err := client.Connect(ctx)
	if !errors.Is(err, ErrAlreadyConnecting) {
		t.Errorf("expected ErrAlreadyConnecting, got %v", err)
	}

	wg.Wait()
	_ = client.Close()
}

func TestEventSubWebSocketClient_ConnectDialError(t *testing.T) {
	client := NewEventSubWebSocketClient(WithWSURL("ws://invalid.invalid:9999"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestEventSubWebSocketClient_WaitForWelcome_WrongMessageType(t *testing.T) {
	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Send keepalive instead of welcome
		msg := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleKeepaliveMessageID,
				MessageType:      WSMessageTypeKeepalive,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(struct{}{}),
		}
		_ = conn.WriteJSON(msg)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for wrong message type")
	}
	if !strings.Contains(err.Error(), "expected welcome message") {
		t.Errorf("expected error about welcome message, got: %v", err)
	}
}

func TestEventSubWebSocketClient_WaitForWelcome_InvalidJSON(t *testing.T) {
	mock := newMockWSServer(func(conn *websocket.Conn) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parsing welcome message") {
		t.Errorf("expected parsing error, got: %v", err)
	}
}

func TestEventSubWebSocketClient_WaitForWelcome_InvalidPayload(t *testing.T) {
	mock := newMockWSServer(func(conn *websocket.Conn) {
		msg := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: json.RawMessage(`"not an object"`),
		}
		_ = conn.WriteJSON(msg)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for invalid payload")
	}
	if !strings.Contains(err.Error(), "parsing welcome payload") {
		t.Errorf("expected payload parsing error, got: %v", err)
	}
}

func TestEventSubWebSocketClient_HandleMessage_InvalidJSON(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	errorReceived := make(chan error, 1)

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send invalid JSON
		_ = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case err := <-errorReceived:
		if !strings.Contains(err.Error(), "parsing message") {
			t.Errorf("expected parsing error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestEventSubWebSocketClient_HandleMessage_PanicRecovery(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	errorReceived := make(chan error, 1)

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send keepalive to trigger handler
		keepalive := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleKeepaliveMessageID,
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
			panic("test panic")
		}),
		WithWSErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case err := <-errorReceived:
		if !strings.Contains(err.Error(), "handler panic") {
			t.Errorf("expected panic error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for panic error")
	}
}

func TestEventSubWebSocketClient_HandleNotification_InvalidPayload(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	errorReceived := make(chan error, 1)

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send notification with invalid payload
		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleNotifMessageID,
				MessageType:      WSMessageTypeNotification,
				MessageTimestamp: time.Now(),
			},
			Payload: json.RawMessage(`"invalid"`),
		}
		_ = conn.WriteJSON(notification)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSNotificationHandler(func(sub *EventSubSubscription, event json.RawMessage) {
			// This shouldn't be called
		}),
		WithWSErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case err := <-errorReceived:
		if !strings.Contains(err.Error(), "parsing notification") {
			t.Errorf("expected notification parsing error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestEventSubWebSocketClient_HandleReconnect_InvalidPayload(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	errorReceived := make(chan error, 1)

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send reconnect with invalid payload
		reconnect := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "reconnect-1",
				MessageType:      WSMessageTypeReconnect,
				MessageTimestamp: time.Now(),
			},
			Payload: json.RawMessage(`"invalid"`),
		}
		_ = conn.WriteJSON(reconnect)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case err := <-errorReceived:
		if !strings.Contains(err.Error(), "parsing reconnect") {
			t.Errorf("expected reconnect parsing error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestEventSubWebSocketClient_HandleRevocation_InvalidPayload(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	errorReceived := make(chan error, 1)

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send revocation with invalid payload
		revocation := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "revoke-1",
				MessageType:      WSMessageTypeRevocation,
				MessageTimestamp: time.Now(),
			},
			Payload: json.RawMessage(`"invalid"`),
		}
		_ = conn.WriteJSON(revocation)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSRevocationHandler(func(sub *EventSubSubscription) {
			// This shouldn't be called
		}),
		WithWSErrorHandler(func(err error) {
			select {
			case errorReceived <- err:
			default:
			}
		}),
	)

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	select {
	case err := <-errorReceived:
		if !strings.Contains(err.Error(), "parsing revocation") {
			t.Errorf("expected revocation parsing error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for error")
	}
}

func TestEventSubWebSocketClient_HandleNotification_NoHandler(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	notificationSent := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send notification (no handler set)
		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleNotifMessageID,
				MessageType:      WSMessageTypeNotification,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:   twitchWSExampleSubscriptionID,
					Type: EventSubTypeChannelFollow,
				},
			}),
		}
		_ = conn.WriteJSON(notification)
		close(notificationSent)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	// No notification handler set
	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for notification to be sent
	select {
	case <-notificationSent:
		// Success - no crash when no handler
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for notification to be sent")
	}
}

func TestEventSubWebSocketClient_HandleRevocation_NoHandler(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	revocationSent := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send revocation (no handler set)
		revocation := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "revoke-1",
				MessageType:      WSMessageTypeRevocation,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:     twitchWSExampleSubscriptionID,
					Status: "authorization_revoked",
				},
			}),
		}
		_ = conn.WriteJSON(revocation)
		close(revocationSent)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	// No revocation handler set
	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for revocation to be sent
	select {
	case <-revocationSent:
		// Success - no crash when no handler
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for revocation to be sent")
	}
}

func TestEventSubWebSocketClient_CloseNotConnected(t *testing.T) {
	client := NewEventSubWebSocketClient()

	// Should not error when not connected
	err := client.Close()
	if err != nil {
		t.Errorf("expected no error for close when not connected, got: %v", err)
	}
}

func TestEventSubWebSocketClient_ReadLoop_NilConnection(t *testing.T) {
	client := NewEventSubWebSocketClient()

	// Manually start readLoop with nil connection (edge case)
	client.wg.Add(1)
	client.stopChan = make(chan struct{})
	go client.readLoop()

	// Wait for readLoop to exit
	client.wg.Wait()

	// Should have exited gracefully
	if client.IsConnected() {
		t.Error("expected client to not be connected")
	}
}

// High-level EventSubWebSocket tests

func TestNewEventSubWebSocket(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient)

	ws := NewEventSubWebSocket(helixClient)
	if ws == nil {
		t.Fatal("expected non-nil EventSubWebSocket")
	}
	if ws.client != helixClient {
		t.Error("expected client to be set")
	}
	if ws.handlers == nil {
		t.Error("expected handlers map to be initialized")
	}
}

func TestEventSubWebSocket_Close_NoConnection(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient)

	ws := NewEventSubWebSocket(helixClient)

	// Should not error when ws.ws is nil
	err := ws.Close()
	if err != nil {
		t.Errorf("expected no error for close with nil ws, got: %v", err)
	}
}

func TestEventSubWebSocket_Connect(t *testing.T) {
	sessionID := twitchWSExampleSessionID

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient)

	ws := NewEventSubWebSocket(helixClient)
	// Inject the mock URL by accessing the internal ws after Connect creates it
	// We need to set the URL before Connect, so we'll use a workaround

	// Create the internal client manually to set the URL
	ws.ws = NewEventSubWebSocketClient(
		WithWSURL(mock.URL()),
		WithWSNotificationHandler(func(sub *EventSubSubscription, event json.RawMessage) {
			ws.mu.RLock()
			handler, ok := ws.handlers[sub.Type]
			ws.mu.RUnlock()
			if ok {
				handler(event)
			}
		}),
	)

	ctx := context.Background()
	sid, err := ws.ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	ws.sessionID = sid

	if ws.sessionID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, ws.sessionID)
	}

	// Close
	err = ws.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestEventSubWebSocket_ConnectError(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient)

	ws := NewEventSubWebSocket(helixClient)
	// Set invalid URL to force connection error
	ws.ws = NewEventSubWebSocketClient(WithWSURL("ws://invalid.invalid:9999"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := ws.ws.Connect(ctx)
	if err == nil {
		t.Error("expected connect error")
	}
}

func TestEventSubWebSocket_Subscribe(t *testing.T) {
	sessionID := twitchWSExampleSessionID

	// Create mock WebSocket server
	wsMock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(500 * time.Millisecond)
	})
	defer wsMock.Close()

	// Create mock API server for subscription creation
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/eventsub/subscriptions" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			resp := EventSubResponse{
				Data: []EventSubSubscription{{
					ID:      twitchWSExampleSubscriptionID,
					Type:    EventSubTypeChannelFollow,
					Version: "2",
					Status:  "enabled",
				}},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer apiServer.Close()

	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient, WithBaseURL(apiServer.URL))

	ws := NewEventSubWebSocket(helixClient)
	ws.ws = NewEventSubWebSocketClient(
		WithWSURL(wsMock.URL()),
		WithWSNotificationHandler(func(sub *EventSubSubscription, event json.RawMessage) {
			ws.mu.RLock()
			handler, ok := ws.handlers[sub.Type]
			ws.mu.RUnlock()
			if ok {
				handler(event)
			}
		}),
	)

	ctx := context.Background()
	sid, err := ws.ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	ws.sessionID = sid

	// Subscribe
	handlerCalled := false
	err = ws.Subscribe(ctx, EventSubTypeChannelFollow, "2", map[string]string{
		"broadcaster_user_id": twitchWSExampleBroadcasterUserID,
		"moderator_user_id":   twitchWSExampleUserID,
	}, func(event json.RawMessage) {
		handlerCalled = true
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	// Check handler is registered
	ws.mu.RLock()
	_, ok := ws.handlers[EventSubTypeChannelFollow]
	ws.mu.RUnlock()
	if !ok {
		t.Error("expected handler to be registered")
	}

	// Simulate calling handler
	ws.handlers[EventSubTypeChannelFollow](json.RawMessage(`{}`))
	if !handlerCalled {
		t.Error("expected handler to be called")
	}

	_ = ws.Close()
}

func TestEventSubWebSocketClient_HandleReconnect_NoHandler(t *testing.T) {
	sessionID := twitchWSExampleSessionID
	reconnectSent := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send reconnect (no handler set)
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
					ReconnectURL: twitchWSExampleReconnectURL,
				},
			}),
		}
		_ = conn.WriteJSON(reconnect)
		close(reconnectSent)
		time.Sleep(100 * time.Millisecond)
	})
	defer mock.Close()

	// No reconnect handler set
	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Wait for reconnect to be sent
	select {
	case <-reconnectSent:
		// Give time for message to be processed
		time.Sleep(50 * time.Millisecond)
		// Check that reconnectURL was stored even without handler
		client.mu.RLock()
		url := client.reconnectURL
		client.mu.RUnlock()
		if url != twitchWSExampleReconnectURL {
			t.Errorf("expected reconnectURL to be stored, got: %s", url)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for reconnect to be sent")
	}
}

func TestEventSubWebSocketClient_Reconnect_WelcomeError(t *testing.T) {
	sessionID := twitchWSExampleSessionID

	// Old server sends welcome properly
	oldServer := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(200 * time.Millisecond)
	})
	defer oldServer.Close()

	// New server sends invalid welcome
	newServer := newMockWSServer(func(conn *websocket.Conn) {
		// Send keepalive instead of welcome
		msg := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        "msg-1",
				MessageType:      WSMessageTypeKeepalive,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(struct{}{}),
		}
		_ = conn.WriteJSON(msg)
	})
	defer newServer.Close()

	client := NewEventSubWebSocketClient(WithWSURL(oldServer.URL()))

	ctx := context.Background()
	_, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Reconnect should fail due to wrong welcome message
	_, err = client.Reconnect(ctx, newServer.URL())
	if err == nil {
		t.Error("expected error for invalid welcome on reconnect")
	}
	if !strings.Contains(err.Error(), "expected welcome message") {
		t.Errorf("expected welcome error, got: %v", err)
	}
}

func TestEventSubWebSocketClient_WaitForWelcome_ReadError(t *testing.T) {
	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Close connection immediately without sending anything
		_ = conn.Close()
	})
	defer mock.Close()

	client := NewEventSubWebSocketClient(WithWSURL(mock.URL()))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Connect(ctx)
	if err == nil {
		t.Error("expected error for read failure")
	}
	if !strings.Contains(err.Error(), "reading welcome message") {
		t.Errorf("expected read error, got: %v", err)
	}
}

func TestEventSubWebSocket_Connect_CancelledContext(t *testing.T) {
	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient)

	ws := NewEventSubWebSocket(helixClient)

	// Use a cancelled context to trigger an error
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := ws.Connect(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
		_ = ws.Close()
	}
}

func TestEventSubWebSocket_Connect_Real(t *testing.T) {
	sessionID := twitchWSExampleSessionID

	mock := newMockWSServer(func(conn *websocket.Conn) {
		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: time.Now(),
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      sessionID,
					Status:                  "connected",
					ConnectedAt:             time.Now(),
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(200 * time.Millisecond)
	})
	defer mock.Close()

	authClient := NewAuthClient(AuthConfig{
		ClientID: "test-client-id",
	})
	helixClient := NewClient("test-client-id", authClient)

	ws := NewEventSubWebSocket(helixClient)

	// We can't directly test Connect because it uses the default URL
	// Instead, test the internal mechanics by creating the ws client manually
	// and verifying the notification handler works

	ws.ws = NewEventSubWebSocketClient(WithWSURL(mock.URL()))
	// Set up notification handler like Connect does
	ws.ws.onNotification = func(sub *EventSubSubscription, event json.RawMessage) {
		ws.mu.RLock()
		handler, ok := ws.handlers[sub.Type]
		ws.mu.RUnlock()
		if ok {
			handler(event)
		}
	}

	ctx := context.Background()
	sid, err := ws.ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	ws.sessionID = sid

	if ws.sessionID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, ws.sessionID)
	}

	// Register a handler and verify it gets called
	handlerCalled := make(chan struct{})
	ws.mu.Lock()
	ws.handlers["test.event"] = func(event json.RawMessage) {
		close(handlerCalled)
	}
	ws.mu.Unlock()

	// Simulate notification
	if ws.ws.onNotification != nil {
		ws.ws.onNotification(&EventSubSubscription{Type: "test.event"}, json.RawMessage(`{}`))
	}

	select {
	case <-handlerCalled:
		// Success
	case <-time.After(time.Second):
		t.Error("handler was not called")
	}

	_ = ws.Close()
}

func TestEventSubWebSocketClient_WithOfficialTwitchExamples(t *testing.T) {
	// Test using official Twitch example values
	var receivedSub *EventSubSubscription
	var receivedEvent json.RawMessage
	notifReceived := make(chan struct{})

	mock := newMockWSServer(func(conn *websocket.Conn) {
		// Send welcome using official Twitch example format
		welcomeTimestamp, _ := time.Parse(time.RFC3339Nano, twitchWSExampleWelcomeTimestamp)
		connectedAt, _ := time.Parse(time.RFC3339Nano, twitchWSExampleConnectedAt)

		welcome := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:        twitchWSExampleWelcomeMessageID,
				MessageType:      WSMessageTypeWelcome,
				MessageTimestamp: welcomeTimestamp,
			},
			Payload: mustMarshal(WebSocketWelcomePayload{
				Session: WebSocketSession{
					ID:                      twitchWSExampleSessionID,
					Status:                  "connected",
					ConnectedAt:             connectedAt,
					KeepaliveTimeoutSeconds: 1,
				},
			}),
		}
		_ = conn.WriteJSON(welcome)
		time.Sleep(50 * time.Millisecond)

		// Send notification using official Twitch example format
		notifTimestamp, _ := time.Parse(time.RFC3339Nano, twitchWSExampleNotifTimestamp)
		createdAt, _ := time.Parse(time.RFC3339Nano, twitchWSExampleNotifTimestamp)

		notification := WebSocketMessage{
			Metadata: WebSocketMetadata{
				MessageID:           twitchWSExampleNotifMessageID,
				MessageType:         WSMessageTypeNotification,
				MessageTimestamp:    notifTimestamp,
				SubscriptionType:    EventSubTypeChannelFollow,
				SubscriptionVersion: "1",
			},
			Payload: mustMarshal(WebSocketNotificationPayload{
				Subscription: EventSubSubscription{
					ID:      twitchWSExampleSubscriptionID,
					Status:  "enabled",
					Type:    EventSubTypeChannelFollow,
					Version: "1",
					Cost:    1,
					Condition: map[string]string{
						"broadcaster_user_id": twitchWSExampleBroadcasterUserID,
					},
					Transport: EventSubTransport{
						Method:    "websocket",
						SessionID: twitchWSExampleTransportSessionID,
					},
					CreatedAt: createdAt,
				},
				Event: mustMarshal(map[string]interface{}{
					"user_id":               twitchWSExampleUserID,
					"user_login":            twitchWSExampleUserLogin,
					"user_name":             twitchWSExampleUserName,
					"broadcaster_user_id":   twitchWSExampleBroadcasterUserID,
					"broadcaster_user_login": twitchWSExampleBroadcasterLogin,
					"broadcaster_user_name": twitchWSExampleBroadcasterName,
					"followed_at":           twitchWSExampleFollowedAt,
				}),
			}),
		}
		_ = conn.WriteJSON(notification)
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

	ctx := context.Background()
	gotSessionID, err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Verify session ID matches official Twitch example
	if gotSessionID != twitchWSExampleSessionID {
		t.Errorf("expected session ID %s, got %s", twitchWSExampleSessionID, gotSessionID)
	}

	// Wait for notification
	select {
	case <-notifReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for notification")
	}

	// Verify subscription matches official Twitch example
	if receivedSub.ID != twitchWSExampleSubscriptionID {
		t.Errorf("expected subscription ID %s, got %s", twitchWSExampleSubscriptionID, receivedSub.ID)
	}
	if receivedSub.Type != EventSubTypeChannelFollow {
		t.Errorf("expected type %s, got %s", EventSubTypeChannelFollow, receivedSub.Type)
	}
	if receivedSub.Transport.SessionID != twitchWSExampleTransportSessionID {
		t.Errorf("expected transport session ID %s, got %s", twitchWSExampleTransportSessionID, receivedSub.Transport.SessionID)
	}

	// Verify event data
	if receivedEvent == nil {
		t.Fatal("expected event data")
	}
}
