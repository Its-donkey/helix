# EventSub WebSocket

Real-time event streaming without requiring a public endpoint.

## Low-Level Client

Full control over WebSocket connection and event handling:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    ctx := context.Background()

    // Create auth and helix clients
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    _, _ = authClient.GetAppAccessToken(ctx)

    helixClient := helix.NewClient("your-client-id", authClient)

    // Create WebSocket client
    wsClient := helix.NewEventSubWebSocketClient(
        helix.WithWSNotificationHandler(func(sub *helix.EventSubSubscription, event json.RawMessage) {
            switch sub.Type {
            case helix.EventSubTypeStreamOnline:
                e, _ := helix.ParseWSEvent[helix.StreamOnlineEvent](event)
                fmt.Printf("Stream online: %s\n", e.BroadcasterUserName)
            case helix.EventSubTypeStreamOffline:
                e, _ := helix.ParseWSEvent[helix.StreamOfflineEvent](event)
                fmt.Printf("Stream offline: %s\n", e.BroadcasterUserName)
            case helix.EventSubTypeChannelFollow:
                e, _ := helix.ParseWSEvent[helix.ChannelFollowEvent](event)
                fmt.Printf("New follower: %s\n", e.UserName)
            }
        }),
        helix.WithWSReconnectHandler(func(reconnectURL string) {
            log.Printf("Reconnecting to: %s", reconnectURL)
            wsClient.Reconnect(ctx, reconnectURL)
        }),
        helix.WithWSErrorHandler(func(err error) {
            log.Printf("WebSocket error: %v", err)
        }),
    )

    // Connect and get session ID
    sessionID, err := wsClient.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer wsClient.Close()

    fmt.Printf("Connected with session ID: %s\n", sessionID)

    // Create EventSub subscriptions using the session ID
    _, err = helixClient.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
        Type:    helix.EventSubTypeStreamOnline,
        Version: "1",
        Condition: map[string]string{
            "broadcaster_user_id": "12345",
        },
        Transport: helix.CreateEventSubTransport{
            Method:    "websocket",
            SessionID: sessionID,
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Keep running to receive events
    select {}
}
```

## High-Level Wrapper

Simplified interface that manages subscriptions automatically:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    ctx := context.Background()

    // Create auth and helix clients
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    _, _ = authClient.GetAppAccessToken(ctx)

    helixClient := helix.NewClient("your-client-id", authClient)

    // Create high-level WebSocket wrapper
    ws := helix.NewEventSubWebSocket(helixClient)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Subscribe to events with automatic handler registration
    ws.Subscribe(ctx, helix.EventSubTypeStreamOnline, "1",
        map[string]string{"broadcaster_user_id": "12345"},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.StreamOnlineEvent](event)
            fmt.Printf("Stream online: %s\n", e.BroadcasterUserName)
        },
    )

    ws.Subscribe(ctx, helix.EventSubTypeChannelFollow, "2",
        map[string]string{
            "broadcaster_user_id": "12345",
            "moderator_user_id":   "12345",
        },
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelFollowEvent](event)
            fmt.Printf("New follower: %s\n", e.UserName)
        },
    )

    // Keep running to receive events
    select {}
}
```

## Handling Reconnection

Twitch may send a reconnect message before maintenance. Handle it gracefully:

```go
wsClient := helix.NewEventSubWebSocketClient(
    helix.WithWSNotificationHandler(handleNotification),
    helix.WithWSReconnectHandler(func(reconnectURL string) {
        log.Printf("Twitch requested reconnection")

        // Reconnect to the new URL
        if err := wsClient.Reconnect(ctx, reconnectURL); err != nil {
            log.Printf("Reconnection failed: %v", err)
            // Implement retry logic or fallback
        }
    }),
    helix.WithWSKeepaliveHandler(func() {
        log.Printf("Keepalive received")
    }),
)
```

## Error Handling

```go
wsClient := helix.NewEventSubWebSocketClient(
    helix.WithWSNotificationHandler(handleNotification),
    helix.WithWSErrorHandler(func(err error) {
        log.Printf("WebSocket error: %v", err)

        // Implement reconnection logic
        go func() {
            time.Sleep(5 * time.Second)
            if _, err := wsClient.Connect(ctx); err != nil {
                log.Printf("Reconnection failed: %v", err)
            }
        }()
    }),
)
```

Note: Expected close errors (like "use of closed network connection" during shutdown) are automatically filtered and won't trigger the error handler.

## Supported Event Types

See [EventSub documentation](../eventsub.md) for a complete list of event types.
