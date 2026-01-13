# PubSub Compatibility Layer

The PubSub compatibility layer provides a familiar PubSub-style API (`Listen`/`Unlisten` with topic strings) while internally using EventSub WebSocket. This enables a smooth migration path for developers transitioning from the deprecated Twitch PubSub system.

> **Note:** Twitch PubSub was fully decommissioned on April 14, 2025. This compatibility layer uses EventSub under the hood, which is Twitch's current real-time event system.

## Quick Start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    // Create auth client and get token
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    token, _ := authClient.GetAppAccessToken(context.Background())
    authClient.SetToken(token)

    // Create Helix client
    helixClient := helix.NewClient("your-client-id", authClient)

    // Create PubSub client
    pubsub := helix.NewPubSubClient(helixClient,
        helix.WithPubSubMessageHandler(func(topic string, message json.RawMessage) {
            fmt.Printf("Received on %s: %s\n", topic, string(message))
        }),
        helix.WithPubSubErrorHandler(func(err error) {
            log.Printf("Error: %v\n", err)
        }),
    )

    // Connect
    ctx := context.Background()
    if err := pubsub.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer pubsub.Close(ctx)

    // Listen to topics (familiar PubSub-style!)
    pubsub.Listen(ctx, "channel-points-channel-v1.12345")
    pubsub.Listen(ctx, "channel-subscribe-events-v1.12345")

    // Block forever (in real app, use proper signal handling)
    select {}
}
```

## Supported Topics

The following PubSub topics are supported and automatically mapped to their EventSub equivalents:

| PubSub Topic | EventSub Type(s) |
|--------------|------------------|
| `channel-bits-events-v1.<channel_id>` | `channel.cheer` |
| `channel-bits-events-v2.<channel_id>` | `channel.cheer` |
| `channel-bits-badge-unlocks.<channel_id>` | `channel.chat.notification` |
| `channel-points-channel-v1.<channel_id>` | `channel.channel_points_custom_reward_redemption.add` |
| `channel-subscribe-events-v1.<channel_id>` | `channel.subscribe`, `channel.subscription.gift`, `channel.subscription.message` |
| `automod-queue.<moderator_id>.<channel_id>` | `automod.message.hold` |
| `chat_moderator_actions.<user_id>.<channel_id>` | `channel.moderate` |
| `whispers.<user_id>` | `user.whisper.message` |

## API Reference

### NewPubSubClient

Creates a new PubSub compatibility client.

```go
pubsub := helix.NewPubSubClient(helixClient, opts...)
```

**Parameters:**
- `helixClient` (*Client): The Helix API client for creating subscriptions
- `opts` (...PubSubOption): Optional configuration functions

### Options

```go
// Set message handler for all topics
helix.WithPubSubMessageHandler(func(topic string, message json.RawMessage) {
    // Handle messages
})

// Set error handler
helix.WithPubSubErrorHandler(func(err error) {
    // Handle errors
})

// Set connection handler
helix.WithPubSubConnectHandler(func() {
    // Called when connected
})

// Set reconnection handler
helix.WithPubSubReconnectHandler(func() {
    // Called after successful reconnection
})

// Set custom WebSocket URL (for testing)
helix.WithPubSubWSURL("wss://custom.example.com/ws")
```

### Connect

Establishes the WebSocket connection to EventSub.

```go
err := pubsub.Connect(ctx)
```

### Listen

Subscribes to a PubSub topic. The topic is automatically translated to the equivalent EventSub subscription(s).

```go
err := pubsub.Listen(ctx, "channel-points-channel-v1.12345")
```

**Notes:**
- Some topics map to multiple EventSub subscriptions (e.g., `channel-subscribe-events-v1` creates 3 subscriptions)
- Listening to the same topic twice is idempotent (no error, no duplicate subscriptions)
- Returns `ErrPubSubNotConnected` if not connected
- Returns `ErrPubSubInvalidTopic` for unrecognized topic formats
- Returns `ErrPubSubUnsupportedTopic` for valid but unsupported topic types

### Unlisten

Unsubscribes from a PubSub topic.

```go
err := pubsub.Unlisten(ctx, "channel-points-channel-v1.12345")
```

### Close

Closes the connection and cleans up all subscriptions.

```go
err := pubsub.Close(ctx)
```

### IsConnected

Returns whether the client is currently connected.

```go
if pubsub.IsConnected() {
    // Connected
}
```

### Topics

Returns the list of topics currently being listened to.

```go
topics := pubsub.Topics()
for _, topic := range topics {
    fmt.Println(topic)
}
```

### SessionID

Returns the EventSub session ID.

```go
sessionID := pubsub.SessionID()
```

## Helper Functions

### ParseTopic

Parses a PubSub topic string into its components.

```go
parsed, err := helix.ParseTopic("channel-points-channel-v1.12345")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Type: %s, ChannelID: %s\n", parsed.Type, parsed.ChannelID)
```

### BuildTopic

Constructs a PubSub topic string from components.

```go
topic := helix.BuildTopic("channel-points-channel-v1", "12345")
// Returns: "channel-points-channel-v1.12345"

topic = helix.BuildTopic("automod-queue", "mod_id", "channel_id")
// Returns: "automod-queue.mod_id.channel_id"
```

### TopicEventSubTypes

Returns the EventSub types that a PubSub topic maps to.

```go
types := helix.TopicEventSubTypes("channel-subscribe-events-v1.12345")
// Returns: ["channel.subscribe", "channel.subscription.gift", "channel.subscription.message"]
```

### SupportedTopics

Returns a list of all supported topic patterns.

```go
patterns := helix.SupportedTopics()
for _, pattern := range patterns {
    fmt.Println(pattern)
}
```

## Message Format

Messages are delivered to your handler wrapped in a `PubSubMessage` envelope:

```go
type PubSubMessage struct {
    Type string          `json:"type"` // EventSub subscription type
    Data json.RawMessage `json:"data"` // EventSub event payload
}
```

Example handler that parses the message:

```go
helix.WithPubSubMessageHandler(func(topic string, message json.RawMessage) {
    var envelope helix.PubSubMessage
    json.Unmarshal(message, &envelope)

    switch envelope.Type {
    case helix.EventSubTypeChannelPointsRedemptionAdd:
        var event helix.ChannelPointsRedemptionAddEvent
        json.Unmarshal(envelope.Data, &event)
        fmt.Printf("Redemption: %s redeemed %s\n",
            event.UserName, event.Reward.Title)

    case helix.EventSubTypeChannelCheer:
        var event helix.ChannelCheerEvent
        json.Unmarshal(envelope.Data, &event)
        fmt.Printf("Cheer: %s cheered %d bits\n",
            event.UserName, event.Bits)
    }
})
```

## Error Handling

The client handles several error scenarios:

```go
helix.WithPubSubErrorHandler(func(err error) {
    switch {
    case errors.Is(err, helix.ErrPubSubNotConnected):
        // Not connected to WebSocket
    case errors.Is(err, helix.ErrPubSubInvalidTopic):
        // Topic format not recognized
    case errors.Is(err, helix.ErrPubSubUnsupportedTopic):
        // Topic type not supported
    default:
        // Other errors (subscription revocation, reconnection failures, etc.)
        log.Printf("PubSub error: %v", err)
    }
})
```

## Reconnection

The client automatically handles EventSub reconnection requests. When the server requests a reconnection:

1. The client connects to the new URL provided by Twitch
2. A new session is established
3. Your `onReconnect` handler is called
4. All existing topic subscriptions remain active (EventSub handles this automatically)

```go
helix.WithPubSubReconnectHandler(func() {
    log.Println("Successfully reconnected to EventSub")
})
```

## Migration from PubSub

If you're migrating from the old PubSub system, the API is intentionally similar:

**Old PubSub code:**
```go
// Old: Direct PubSub (no longer works)
pubsub.Listen("channel-points-channel-v1.12345", token)
```

**New Kappopher code:**
```go
// New: PubSub compatibility layer (uses EventSub)
pubsub := helix.NewPubSubClient(helixClient, ...)
pubsub.Connect(ctx)
pubsub.Listen(ctx, "channel-points-channel-v1.12345")
```

Key differences:
1. Requires a Helix client (for creating EventSub subscriptions via API)
2. Uses `context.Context` for all operations
3. Messages are wrapped in a `PubSubMessage` envelope with the EventSub type
4. Event payloads use EventSub format (not old PubSub format)
