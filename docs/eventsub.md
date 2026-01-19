---
layout: default
title: EventSub API
description: Manage Twitch EventSub subscriptions for real-time event notifications.
---

## GetEventSubSubscriptions

Get a list of your EventSub subscriptions.

**Requires:** App access token

```go
// Get all subscriptions
resp, err := client.GetEventSubSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{})

// Filter by status (enabled, webhook_callback_verification_pending, etc.)
resp, err = client.GetEventSubSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    Status: "enabled",
})

// Filter by subscription type
resp, err = client.GetEventSubSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    Type: helix.EventSubTypeStreamOnline,
})

// Filter by user ID
resp, err = client.GetEventSubSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    UserID: "12345",
})

// With pagination
resp, err = client.GetEventSubSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    PaginationParams: &helix.PaginationParams{
        First: 100,
        After: "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19",
    },
})

fmt.Printf("Total: %d, Total Cost: %d, Max Total Cost: %d\n",
    resp.Total, resp.TotalCost, resp.MaxTotalCost)

for _, sub := range resp.Data {
    fmt.Printf("Subscription: %s (Type: %s, Status: %s)\n",
        sub.ID, sub.Type, sub.Status)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
      "status": "enabled",
      "type": "stream.online",
      "version": "1",
      "condition": {
        "broadcaster_user_id": "12345"
      },
      "created_at": "2021-07-15T17:16:03.17106713Z",
      "transport": {
        "method": "webhook",
        "callback": "https://example.com/webhooks/eventsub"
      },
      "cost": 1
    },
    {
      "id": "a8e1c4f3-2d1b-49e9-b865-1f31c7a2e4b5",
      "status": "enabled",
      "type": "channel.follow",
      "version": "2",
      "condition": {
        "broadcaster_user_id": "12345",
        "moderator_user_id": "12345"
      },
      "created_at": "2021-07-15T18:22:14.23456789Z",
      "transport": {
        "method": "websocket",
        "session_id": "AgoQHR3s6bRQjsoiS4vB7gRRc4msQB57iQ",
        "connected_at": "2021-07-15T18:22:10.123456789Z"
      },
      "cost": 1
    }
  ],
  "total": 2,
  "total_cost": 2,
  "max_total_cost": 10000,
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## CreateEventSubSubscription

Create a new EventSub subscription.

**Requires:** App access token

```go
// Create a webhook subscription for stream online events
resp, err := client.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
    Type:    helix.EventSubTypeStreamOnline,
    Version: "1",
    Condition: map[string]string{
        "broadcaster_user_id": "12345",
    },
    Transport: helix.EventSubTransport{
        Method:   "webhook",
        Callback: "https://example.com/webhooks/eventsub",
        Secret:   "your-webhook-secret",
    },
})

// Create a WebSocket subscription for channel follow events
resp, err = client.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
    Type:    helix.EventSubTypeChannelFollow,
    Version: "2",
    Condition: map[string]string{
        "broadcaster_user_id": "12345",
        "moderator_user_id":   "12345",
    },
    Transport: helix.EventSubTransport{
        Method:    "websocket",
        SessionID: "websocket-session-id",
    },
})

// Create a conduit subscription
resp, err = client.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
    Type:    helix.EventSubTypeChannelUpdate,
    Version: "2",
    Condition: map[string]string{
        "broadcaster_user_id": "12345",
    },
    Transport: helix.EventSubTransport{
        Method:    "conduit",
        ConduitID: "conduit-id",
    },
})

if len(resp.Data) > 0 {
    fmt.Printf("Created subscription: %s (Status: %s)\n",
        resp.Data[0].ID, resp.Data[0].Status)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "f1c2a387-161a-49f9-a165-0f21d7a4e1c4",
      "status": "webhook_callback_verification_pending",
      "type": "stream.online",
      "version": "1",
      "condition": {
        "broadcaster_user_id": "12345"
      },
      "created_at": "2021-07-15T17:16:03.17106713Z",
      "transport": {
        "method": "webhook",
        "callback": "https://example.com/webhooks/eventsub"
      },
      "cost": 1
    }
  ],
  "total": 1,
  "total_cost": 1,
  "max_total_cost": 10000
}
```

## DeleteEventSubSubscription

Delete an EventSub subscription.

**Requires:** App access token

```go
err := client.DeleteEventSubSubscription(ctx, "subscription-id")
if err != nil {
    fmt.Printf("Failed to delete subscription: %v\n", err)
}
```

**Parameters:**
- `subscriptionID` (string): The ID of the subscription to delete

**Returns:**
- No response body. Returns an error if the deletion fails.

## EventSub Subscription Types

The following constants are available for EventSub subscription types:

### Channel Events

- `EventSubTypeChannelUpdate` - Channel update events (title, category, etc.)
- `EventSubTypeChannelFollow` - New follower events
- `EventSubTypeChannelSubscribe` - New subscription events
- `EventSubTypeChannelSubscriptionEnd` - Subscription end events
- `EventSubTypeChannelSubscriptionGift` - Gift subscription events
- `EventSubTypeChannelSubscriptionMessage` - Resubscription message events
- `EventSubTypeChannelCheer` - Bits cheer events
- `EventSubTypeChannelRaid` - Channel raid events
- `EventSubTypeChannelBan` - User ban events
- `EventSubTypeChannelUnban` - User unban events
- `EventSubTypeChannelModeratorAdd` - Moderator added events
- `EventSubTypeChannelModeratorRemove` - Moderator removed events

### Channel Points Events

- `EventSubTypeChannelPointsRewardAdd` - Custom reward created
- `EventSubTypeChannelPointsRewardUpdate` - Custom reward updated
- `EventSubTypeChannelPointsRewardRemove` - Custom reward removed
- `EventSubTypeChannelPointsRedemptionAdd` - Custom reward redeemed
- `EventSubTypeChannelPointsRedemptionUpdate` - Redemption status updated

### Poll Events

- `EventSubTypeChannelPollBegin` - Poll started
- `EventSubTypeChannelPollProgress` - Poll progress update
- `EventSubTypeChannelPollEnd` - Poll ended

### Prediction Events

- `EventSubTypeChannelPredictionBegin` - Prediction started
- `EventSubTypeChannelPredictionProgress` - Prediction progress update
- `EventSubTypeChannelPredictionLock` - Prediction locked
- `EventSubTypeChannelPredictionEnd` - Prediction ended

### Hype Train Events

- `EventSubTypeChannelHypeTrainBegin` - Hype Train started
- `EventSubTypeChannelHypeTrainProgress` - Hype Train progress update
- `EventSubTypeChannelHypeTrainEnd` - Hype Train ended

**Note:** Hype Train v1 is deprecated by Twitch. This library defaults to v2.

```go
resp, err := client.CreateEventSubSubscription(ctx, &helix.CreateEventSubSubscriptionParams{
    Type:    helix.EventSubTypeChannelHypeTrainBegin,
    Version: helix.EventSubVersionHypeTrainV2, // or omit for default v2
    // ...
})
```

**V2 Fields:**
- `Type` - Hype train type: `regular`, `golden_kappa`, or `shared`
- `IsSharedTrain` - Whether this is a shared hype train across multiple channels
- `SharedTrainParticipants` - List of participating broadcasters (for shared trains)
- `AllTimeHighLevel` - Channel's all-time highest hype train level
- `AllTimeHighTotal` - Channel's all-time highest hype train total

**Migration from V1:** The library automatically converts v1 fields to v2 during JSON unmarshaling to ease migration:
- `IsGoldenKappaTrain=true` → `Type` is set to `golden_kappa`
- `IsGoldenKappaTrain=false` → `Type` is set to `regular`

This allows existing code using `IsGoldenKappaTrain` to continue working while you migrate to the `Type` field.

### Stream Events

- `EventSubTypeStreamOnline` - Stream went online
- `EventSubTypeStreamOffline` - Stream went offline

### User Events

- `EventSubTypeUserUpdate` - User profile updated

## Helper Functions

These convenience functions simplify common EventSub operations.

### Condition Builders

Build condition maps for EventSub subscriptions:

```go
// For channel events (broadcaster_user_id)
cond := helix.BroadcasterCondition("12345")
// Returns: map[string]string{"broadcaster_user_id": "12345"}

// For events requiring moderator (follows v2, chat settings, etc.)
cond := helix.BroadcasterModeratorCondition("12345", "67890")
// Returns: map[string]string{"broadcaster_user_id": "12345", "moderator_user_id": "67890"}

// For user events (user_id)
cond := helix.UserCondition("12345")
// Returns: map[string]string{"user_id": "12345"}

// For channel points with optional reward filter
cond := helix.ChannelPointsCondition("12345", "reward-uuid")
// Returns: map[string]string{"broadcaster_user_id": "12345", "reward_id": "reward-uuid"}
```

### SubscribeToChannel

Subscribe to channel events with automatic version selection.

```go
transport := helix.CreateEventSubTransport{
    Method:    "websocket",
    SessionID: wsClient.SessionID(),
}

// Subscribe to stream online events
sub, err := client.SubscribeToChannel(ctx, helix.EventSubTypeStreamOnline, "12345", transport)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Subscribed: %s\n", sub.ID)

// Subscribe to raid events
sub, err = client.SubscribeToChannel(ctx, helix.EventSubTypeChannelRaid, "12345", transport)
```

### SubscribeToChannelWithModerator

Subscribe to channel events that require moderator authorization (e.g., follows v2, chat settings).

```go
transport := helix.CreateEventSubTransport{
    Method:    "websocket",
    SessionID: wsClient.SessionID(),
}

// Subscribe to follow events (requires moderator)
sub, err := client.SubscribeToChannelWithModerator(ctx,
    helix.EventSubTypeChannelFollow,
    "12345",  // broadcaster ID
    "67890",  // moderator ID (often same as broadcaster)
    transport,
)

// Subscribe to chat settings updates
sub, err = client.SubscribeToChannelWithModerator(ctx,
    helix.EventSubTypeChannelChatSettingsUpdate,
    "12345",
    "67890",
    transport,
)
```

### SubscribeToUser

Subscribe to user-specific events.

```go
transport := helix.CreateEventSubTransport{
    Method:    "websocket",
    SessionID: wsClient.SessionID(),
}

// Subscribe to user update events
sub, err := client.SubscribeToUser(ctx, helix.EventSubTypeUserUpdate, "12345", transport)

// Subscribe to whisper events
sub, err = client.SubscribeToUser(ctx, helix.EventSubTypeWhisperReceived, "12345", transport)
```

### GetAllSubscriptions

Retrieve all EventSub subscriptions with automatic pagination.

```go
// Get all subscriptions
subs, err := client.GetAllSubscriptions(ctx, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Total subscriptions: %d\n", len(subs))

// Get all enabled subscriptions
subs, err = client.GetAllSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    Status: "enabled",
})

// Get all subscriptions of a specific type
subs, err = client.GetAllSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    Type: helix.EventSubTypeStreamOnline,
})

// Get all subscriptions for a user
subs, err = client.GetAllSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    UserID: "12345",
})
```

### DeleteAllSubscriptions

Delete all EventSub subscriptions matching a filter.

```go
// Delete all subscriptions (use with caution!)
deleted, err := client.DeleteAllSubscriptions(ctx, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Deleted %d subscriptions\n", deleted)

// Delete all subscriptions of a specific type
deleted, err = client.DeleteAllSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    Type: helix.EventSubTypeStreamOnline,
})

// Delete all failed subscriptions
deleted, err = client.DeleteAllSubscriptions(ctx, &helix.GetEventSubSubscriptionsParams{
    Status: "webhook_callback_verification_failed",
})
```

### GetEventSubVersion

Get the current default version for an EventSub subscription type.

```go
version := helix.GetEventSubVersion(helix.EventSubTypeChannelFollow)
// Returns "2" for channel.follow

version = helix.GetEventSubVersion(helix.EventSubTypeChannelHypeTrainBegin)
// Returns "2" for hype train events (v1 deprecated)

version = helix.GetEventSubVersion(helix.EventSubTypeStreamOnline)
// Returns "1" for most events
```

