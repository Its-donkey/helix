# EventSub API

Manage Twitch EventSub subscriptions for real-time event notifications.

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
