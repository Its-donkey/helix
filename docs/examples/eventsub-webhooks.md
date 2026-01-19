---
layout: default
title: EventSub Webhooks
description: Handle EventSub webhook notifications with built-in signature verification.
---

## Overview

EventSub Webhooks allow Twitch to push real-time notifications to your server. Unlike WebSocket, webhooks require a publicly accessible HTTPS endpoint.

**When to use Webhooks**:
- Server-side applications with a public endpoint
- When you need events even when your app isn't actively connected
- For high-reliability scenarios (Twitch retries failed deliveries)

**When to use WebSocket instead**:
- Client-side or local applications
- When you don't have a public HTTPS endpoint
- For simpler setup without SSL certificate management

**How it works**:
1. Your server exposes an HTTPS endpoint
2. You create subscriptions telling Twitch where to send events
3. Twitch verifies your endpoint ownership via a challenge request
4. Twitch sends signed notifications to your endpoint
5. Your handler verifies signatures and processes events

## Basic Setup

The handler automatically verifies HMAC signatures to ensure notifications are from Twitch.

```go
handler := helix.NewEventSubWebhookHandler(
    helix.WithWebhookSecret("your-webhook-secret"),
    helix.WithNotificationHandler(func(msg *helix.EventSubWebhookMessage) {
        switch msg.SubscriptionType {
        case helix.EventSubTypeStreamOnline:
            event, _ := helix.ParseEventSubEvent[helix.StreamOnlineEvent](msg)
            fmt.Printf("Stream online: %s\n", event.BroadcasterUserName)
        case helix.EventSubTypeChannelFollow:
            event, _ := helix.ParseEventSubEvent[helix.ChannelFollowEvent](msg)
            fmt.Printf("New follower: %s\n", event.UserName)
        }
    }),
    helix.WithVerificationHandler(func(msg *helix.EventSubWebhookMessage) bool {
        return true // Accept all subscription verifications
    }),
)

http.Handle("/webhook", handler)
```

## Complete Example

A full webhook server handling multiple event types. The handler uses separate callbacks for:
- **Notifications**: Actual events (follows, subscriptions, stream status, etc.)
- **Verification**: Challenge requests when creating subscriptions
- **Revocation**: When Twitch revokes a subscription (e.g., token expired)

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/Its-donkey/helix/helix"
)

func main() {
    handler := helix.NewEventSubWebhookHandler(
        helix.WithWebhookSecret("your-webhook-secret"),
        helix.WithNotificationHandler(handleNotification),
        helix.WithVerificationHandler(handleVerification),
        helix.WithRevocationHandler(handleRevocation),
    )

    http.Handle("/webhook", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleNotification(msg *helix.EventSubWebhookMessage) {
    switch msg.SubscriptionType {
    case helix.EventSubTypeStreamOnline:
        event, err := helix.ParseEventSubEvent[helix.StreamOnlineEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("Stream online: %s started streaming\n", event.BroadcasterUserName)

    case helix.EventSubTypeStreamOffline:
        event, err := helix.ParseEventSubEvent[helix.StreamOfflineEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("Stream offline: %s stopped streaming\n", event.BroadcasterUserName)

    case helix.EventSubTypeChannelFollow:
        event, err := helix.ParseEventSubEvent[helix.ChannelFollowEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("New follower: %s followed %s\n", event.UserName, event.BroadcasterUserName)

    case helix.EventSubTypeChannelSubscribe:
        event, err := helix.ParseEventSubEvent[helix.ChannelSubscribeEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("New subscriber: %s subscribed with tier %s\n", event.UserName, event.Tier)

    case helix.EventSubTypeChannelCheer:
        event, err := helix.ParseEventSubEvent[helix.ChannelCheerEvent](msg)
        if err != nil {
            log.Printf("Failed to parse event: %v", err)
            return
        }
        fmt.Printf("Cheer: %s cheered %d bits\n", event.UserName, event.Bits)

    default:
        log.Printf("Unhandled event type: %s", msg.SubscriptionType)
    }
}

func handleVerification(msg *helix.EventSubWebhookMessage) bool {
    log.Printf("Verifying subscription: %s", msg.SubscriptionType)
    return true // Accept all verifications
}

func handleRevocation(msg *helix.EventSubWebhookMessage) {
    log.Printf("Subscription revoked: %s (reason: %s)",
        msg.SubscriptionType, msg.Subscription.Status)
}
```

## Creating Subscriptions

After setting up your webhook handler, create subscriptions to tell Twitch which events to send.

**Important**:
- The callback URL must be HTTPS (HTTP is not accepted)
- The secret must match between subscription creation and webhook handler
- Twitch will send a verification challenge immediately after subscription creation

```go
ctx := context.Background()

// Create subscription for stream online events
_, err := client.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
    Type:    helix.EventSubTypeStreamOnline,
    Version: "1",
    Condition: map[string]string{
        "broadcaster_user_id": "12345",
    },
    Transport: helix.CreateEventSubTransport{
        Method:   "webhook",
        Callback: "https://your-domain.com/webhook",
        Secret:   "your-webhook-secret",
    },
})
if err != nil {
    log.Fatal(err)
}
```

## Supported Event Types
See [EventSub documentation](eventsub.md) for a complete list of event types.

