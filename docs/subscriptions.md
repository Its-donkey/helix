---
layout: default
title: Subscriptions API
description: Retrieve subscriber information for Twitch channels.
---

## GetBroadcasterSubscriptions

Get a list of subscribers for a broadcaster's channel.

**Requires:** `channel:read:subscriptions`

```go
// Get all subscribers for a channel
resp, err := client.GetBroadcasterSubscriptions(ctx, &helix.GetBroadcasterSubscriptionsParams{
    BroadcasterID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

// Get specific subscribers by user IDs (max 100)
resp, err = client.GetBroadcasterSubscriptions(ctx, &helix.GetBroadcasterSubscriptionsParams{
    BroadcasterID: "12345",
    UserIDs:       []string{"67890", "11111", "22222"},
})

for _, sub := range resp.Data {
    if sub.IsGift {
        fmt.Printf("%s (gifted by %s) - Tier: %s, Plan: %s\n",
            sub.UserName, sub.GifterName, sub.Tier, sub.PlanName)
    } else {
        fmt.Printf("%s - Tier: %s, Plan: %s\n",
            sub.UserName, sub.Tier, sub.PlanName)
    }
}

fmt.Printf("Total subscribers: %d\n", resp.Total)
fmt.Printf("Subscriber points: %d\n", resp.Points)
```

**Parameters:**
- `BroadcasterID` (string, required): The broadcaster's user ID
- `UserIDs` ([]string, optional): Filter to specific user IDs (max 100)
- `PaginationParams` (optional): Pagination options

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "gifter_id": "98765",
      "gifter_login": "giftgiver123",
      "gifter_name": "GiftGiver123",
      "is_gift": true,
      "plan_name": "Channel Subscription (twitchdev)",
      "tier": "1000",
      "user_id": "67890",
      "user_login": "subscriber1",
      "user_name": "Subscriber1"
    },
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "is_gift": false,
      "plan_name": "Channel Subscription (twitchdev): $9.99 Sub",
      "tier": "2000",
      "user_id": "11111",
      "user_login": "subscriber2",
      "user_name": "Subscriber2"
    },
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "is_gift": false,
      "plan_name": "Channel Subscription (twitchdev): $24.99 Sub",
      "tier": "3000",
      "user_id": "22222",
      "user_login": "subscriber3",
      "user_name": "Subscriber3"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjoiMTUwMzQ0MTc3NjQyNDQyMjAwMCJ9"
  },
  "total": 150,
  "points": 350
}
```

**Response:**
- `Data` ([]Subscription): Array of subscription objects
  - `BroadcasterID` (string): Broadcaster's user ID
  - `BroadcasterLogin` (string): Broadcaster's login name
  - `BroadcasterName` (string): Broadcaster's display name
  - `GifterID` (string): Gift giver's user ID (if gifted)
  - `GifterLogin` (string): Gift giver's login name (if gifted)
  - `GifterName` (string): Gift giver's display name (if gifted)
  - `IsGift` (bool): Whether the subscription is a gift
  - `PlanName` (string): Subscription plan name
  - `Tier` (string): Subscription tier (`1000`, `2000`, or `3000`)
  - `UserID` (string): Subscriber's user ID
  - `UserLogin` (string): Subscriber's login name
  - `UserName` (string): Subscriber's display name
- `Total` (int): Total number of subscribers
- `Points` (int): Subscriber points based on subscription tiers

## CheckUserSubscription

Check if a specific user is subscribed to a broadcaster.

**Requires:** `user:read:subscriptions`

```go
// Check if a user is subscribed to a broadcaster
sub, err := client.CheckUserSubscription(ctx, "12345", "67890")
if err != nil {
    fmt.Printf("Error checking subscription: %v\n", err)
    return
}

if sub != nil {
    fmt.Printf("User is subscribed - Tier: %s, Plan: %s\n", sub.Tier, sub.PlanName)
    if sub.IsGift {
        fmt.Printf("Subscription was gifted by: %s\n", sub.GifterName)
    }
} else {
    fmt.Println("User is not subscribed")
}
```

**Parameters:**
- `broadcasterID` (string, required): The broadcaster's user ID
- `userID` (string, required): The user ID to check

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "gifter_id": "98765",
      "gifter_login": "giftgiver123",
      "gifter_name": "GiftGiver123",
      "is_gift": true,
      "tier": "1000"
    }
  ]
}
```

**Response:**
- Returns `UserSubscription` object if subscribed, `nil` if not subscribed

