# PubSub Migration Example

This example demonstrates how to use the PubSub compatibility layer to receive real-time events using familiar PubSub-style topic subscriptions.

> **Note:** Twitch PubSub was decommissioned on April 14, 2025. This compatibility layer uses EventSub WebSocket internally but provides a PubSub-like API for easier migration.

## Basic Usage

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/signal"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Create auth client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
        ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
    })

    // Get app access token
    token, err := authClient.GetAppAccessToken(ctx)
    if err != nil {
        log.Fatalf("Failed to get token: %v", err)
    }
    authClient.SetToken(token)

    // Create Helix client
    helixClient := helix.NewClient(os.Getenv("TWITCH_CLIENT_ID"), authClient)

    // Create PubSub client with handlers
    pubsub := helix.NewPubSubClient(helixClient,
        helix.WithPubSubMessageHandler(handleMessage),
        helix.WithPubSubErrorHandler(handleError),
        helix.WithPubSubConnectHandler(func() {
            log.Println("Connected to EventSub")
        }),
        helix.WithPubSubReconnectHandler(func() {
            log.Println("Reconnected to EventSub")
        }),
    )

    // Connect
    if err := pubsub.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer pubsub.Close(ctx)

    channelID := "12345" // Replace with actual channel ID

    // Listen to topics (familiar PubSub-style API)
    topics := []string{
        helix.BuildTopic("channel-points-channel-v1", channelID),
        helix.BuildTopic("channel-subscribe-events-v1", channelID),
        helix.BuildTopic("channel-bits-events-v2", channelID),
    }

    for _, topic := range topics {
        if err := pubsub.Listen(ctx, topic); err != nil {
            log.Printf("Failed to listen to %s: %v", topic, err)
        } else {
            log.Printf("Listening to: %s", topic)
        }
    }

    // Wait for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan

    log.Println("Shutting down...")
}

func handleMessage(topic string, message json.RawMessage) {
    // Parse the envelope
    var envelope helix.PubSubMessage
    if err := json.Unmarshal(message, &envelope); err != nil {
        log.Printf("Failed to parse envelope: %v", err)
        return
    }

    log.Printf("Event on %s (type: %s)", topic, envelope.Type)

    // Route to specific handler based on EventSub type
    switch envelope.Type {
    case helix.EventSubTypeChannelPointsRedemptionAdd:
        handleRedemption(envelope.Data)
    case helix.EventSubTypeChannelCheer:
        handleCheer(envelope.Data)
    case helix.EventSubTypeChannelSubscribe:
        handleSubscription(envelope.Data)
    case helix.EventSubTypeChannelSubscriptionGift:
        handleGiftSub(envelope.Data)
    case helix.EventSubTypeChannelSubscriptionMessage:
        handleResub(envelope.Data)
    default:
        log.Printf("Unhandled event type: %s", envelope.Type)
    }
}

func handleRedemption(data json.RawMessage) {
    var event helix.ChannelPointsRedemptionAddEvent
    if err := json.Unmarshal(data, &event); err != nil {
        log.Printf("Failed to parse redemption: %v", err)
        return
    }
    fmt.Printf("🎁 %s redeemed '%s' for %d points\n",
        event.UserName, event.Reward.Title, event.Reward.Cost)
}

func handleCheer(data json.RawMessage) {
    var event helix.ChannelCheerEvent
    if err := json.Unmarshal(data, &event); err != nil {
        log.Printf("Failed to parse cheer: %v", err)
        return
    }
    fmt.Printf("💎 %s cheered %d bits: %s\n",
        event.UserName, event.Bits, event.Message)
}

func handleSubscription(data json.RawMessage) {
    var event helix.ChannelSubscribeEvent
    if err := json.Unmarshal(data, &event); err != nil {
        log.Printf("Failed to parse subscription: %v", err)
        return
    }
    fmt.Printf("⭐ New subscriber: %s (Tier %s)\n",
        event.UserName, event.Tier)
}

func handleGiftSub(data json.RawMessage) {
    var event helix.ChannelSubscriptionGiftEvent
    if err := json.Unmarshal(data, &event); err != nil {
        log.Printf("Failed to parse gift sub: %v", err)
        return
    }
    fmt.Printf("🎉 %s gifted %d subs!\n",
        event.UserName, event.Total)
}

func handleResub(data json.RawMessage) {
    var event helix.ChannelSubscriptionMessageEvent
    if err := json.Unmarshal(data, &event); err != nil {
        log.Printf("Failed to parse resub: %v", err)
        return
    }
    fmt.Printf("🔄 %s resubscribed for %d months: %s\n",
        event.UserName, event.CumulativeMonths, event.Message.Text)
}

func handleError(err error) {
    log.Printf("PubSub error: %v", err)
}
```

## Multiple Channels

You can listen to multiple channels by creating multiple topic strings:

```go
channels := []string{"12345", "67890", "11111"}

for _, channelID := range channels {
    topic := helix.BuildTopic("channel-points-channel-v1", channelID)
    if err := pubsub.Listen(ctx, topic); err != nil {
        log.Printf("Failed to listen to channel %s: %v", channelID, err)
    }
}
```

## Dynamic Subscribe/Unsubscribe

You can add and remove subscriptions at runtime:

```go
// Add a new topic
newTopic := helix.BuildTopic("channel-bits-events-v1", "99999")
pubsub.Listen(ctx, newTopic)

// Remove a topic
pubsub.Unlisten(ctx, newTopic)

// Check current topics
for _, topic := range pubsub.Topics() {
    fmt.Println("Active:", topic)
}
```

## Supported Topics Reference

```go
// Check what topics are supported
for _, pattern := range helix.SupportedTopics() {
    fmt.Println(pattern)
}

// Check what EventSub types a topic maps to
types := helix.TopicEventSubTypes("channel-subscribe-events-v1.12345")
fmt.Println(types) // [channel.subscribe, channel.subscription.gift, channel.subscription.message]
```

## Migration Checklist

When migrating from old PubSub code:

1. ✅ Replace direct PubSub connection with `helix.NewPubSubClient`
2. ✅ Add `context.Context` to all calls
3. ✅ Update message handler to parse `PubSubMessage` envelope
4. ✅ Update event parsing to use EventSub types (slightly different field names)
5. ✅ Handle that some topics create multiple EventSub subscriptions
6. ✅ Test with actual Twitch events

## EventSub Event Types

For reference, here are the EventSub event types that PubSub topics map to:

| Topic Pattern | EventSub Types |
|---------------|----------------|
| `channel-bits-events-*` | `channel.cheer` |
| `channel-points-channel-v1` | `channel.channel_points_custom_reward_redemption.add` |
| `channel-subscribe-events-v1` | `channel.subscribe`, `channel.subscription.gift`, `channel.subscription.message` |
| `automod-queue` | `automod.message.hold` |
| `chat_moderator_actions` | `channel.moderate` |
| `whispers` | `user.whisper.message` |
