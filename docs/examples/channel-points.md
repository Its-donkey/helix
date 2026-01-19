---
layout: default
title: Channel Points Examples
description: Manage custom rewards and handle redemptions.
---

## Overview

Channel Points are a loyalty system where viewers earn points by watching and engaging with streams. Broadcasters can create custom rewards that viewers redeem with their points.

**What you can do**:
- Create, update, and delete custom rewards
- Get pending redemptions and update their status (fulfill/cancel)
- Listen for real-time redemption events via EventSub
- Build automated reward fulfillment systems

**Common use cases**:
- Song request systems
- TTS (text-to-speech) messages
- Game actions triggered by viewers
- VIP perks and special interactions

## Prerequisites

Channel Points require user authentication with these scopes:
- `channel:read:redemptions` - Read redemptions
- `channel:manage:redemptions` - Manage redemption status
- `channel:manage:rewards` - Create/manage custom rewards (broadcaster only)

## Create Custom Rewards

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Setup client with broadcaster token
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"

    // Create a simple reward
    reward, err := client.CreateCustomReward(ctx, &helix.CreateCustomRewardParams{
        BroadcasterID: broadcasterID,
        Title:         "Hydrate!",
        Cost:          100,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created reward: %s (ID: %s)\n", reward.Data[0].Title, reward.Data[0].ID)

    // Create a reward with all options
    fullReward, err := client.CreateCustomReward(ctx, &helix.CreateCustomRewardParams{
        BroadcasterID:                    broadcasterID,
        Title:                            "Song Request",
        Cost:                             500,
        Prompt:                           "Enter the song name and artist",
        IsEnabled:                        boolPtr(true),
        BackgroundColor:                  "#FF0000",
        IsUserInputRequired:              true,
        IsMaxPerStreamEnabled:            true,
        MaxPerStream:                     10,
        IsMaxPerUserPerStreamEnabled:     true,
        MaxPerUserPerStream:              2,
        IsGlobalCooldownEnabled:          true,
        GlobalCooldownSeconds:            300,
        ShouldRedemptionsSkipRequestQueue: false,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created reward: %s\n", fullReward.Data[0].Title)
}

func boolPtr(b bool) *bool { return &b }
```

## Get Custom Rewards

```go
// Get all custom rewards
rewards, err := client.GetCustomRewards(ctx, &helix.GetCustomRewardsParams{
    BroadcasterID: broadcasterID,
})
if err != nil {
    log.Fatal(err)
}

for _, reward := range rewards.Data {
    fmt.Printf("Reward: %s\n", reward.Title)
    fmt.Printf("  ID: %s\n", reward.ID)
    fmt.Printf("  Cost: %d\n", reward.Cost)
    fmt.Printf("  Enabled: %v\n", reward.IsEnabled)
    fmt.Printf("  Paused: %v\n", reward.IsPaused)
    fmt.Printf("  In Stock: %v\n", reward.IsInStock)
}

// Get specific rewards by ID
specificRewards, err := client.GetCustomRewards(ctx, &helix.GetCustomRewardsParams{
    BroadcasterID: broadcasterID,
    IDs:           []string{"reward-id-1", "reward-id-2"},
})

// Get only manageable rewards (created by your app)
manageableRewards, err := client.GetCustomRewards(ctx, &helix.GetCustomRewardsParams{
    BroadcasterID:      broadcasterID,
    OnlyManageableRewards: true,
})
```

## Update Custom Rewards

```go
// Update reward cost and title
updated, err := client.UpdateCustomReward(ctx, &helix.UpdateCustomRewardParams{
    BroadcasterID: broadcasterID,
    ID:            "reward-id",
    Title:         "Song Request (Updated)",
    Cost:          750,
})

// Pause a reward
paused, err := client.UpdateCustomReward(ctx, &helix.UpdateCustomRewardParams{
    BroadcasterID: broadcasterID,
    ID:            "reward-id",
    IsPaused:      boolPtr(true),
})

// Disable a reward
disabled, err := client.UpdateCustomReward(ctx, &helix.UpdateCustomRewardParams{
    BroadcasterID: broadcasterID,
    ID:            "reward-id",
    IsEnabled:     boolPtr(false),
})
```

## Delete Custom Rewards

```go
// Delete a reward (only rewards created by your app can be deleted)
err := client.DeleteCustomReward(ctx, broadcasterID, "reward-id")
if err != nil {
    log.Printf("Failed to delete reward: %v", err)
}
```

## Get Redemptions

```go
// Get unfulfilled redemptions
redemptions, err := client.GetCustomRewardRedemptions(ctx, &helix.GetCustomRewardRedemptionsParams{
    BroadcasterID: broadcasterID,
    RewardID:      "reward-id",
    Status:        "UNFULFILLED",
})
if err != nil {
    log.Fatal(err)
}

for _, redemption := range redemptions.Data {
    fmt.Printf("Redemption by %s\n", redemption.UserName)
    fmt.Printf("  ID: %s\n", redemption.ID)
    fmt.Printf("  User Input: %s\n", redemption.UserInput)
    fmt.Printf("  Redeemed At: %s\n", redemption.RedeemedAt)
    fmt.Printf("  Status: %s\n", redemption.Status)
}

// Get redemptions with pagination
allRedemptions, err := client.GetCustomRewardRedemptions(ctx, &helix.GetCustomRewardRedemptionsParams{
    BroadcasterID: broadcasterID,
    RewardID:      "reward-id",
    Status:        "UNFULFILLED",
    First:         50,
    Sort:          "NEWEST",
})
```

## Update Redemption Status

```go
// Fulfill a redemption
fulfilled, err := client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
    BroadcasterID: broadcasterID,
    RewardID:      "reward-id",
    IDs:           []string{"redemption-id"},
    Status:        "FULFILLED",
})

// Cancel a redemption (refunds points)
cancelled, err := client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
    BroadcasterID: broadcasterID,
    RewardID:      "reward-id",
    IDs:           []string{"redemption-id"},
    Status:        "CANCELED",
})

// Batch update multiple redemptions
batchUpdated, err := client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
    BroadcasterID: broadcasterID,
    RewardID:      "reward-id",
    IDs:           []string{"id-1", "id-2", "id-3"},
    Status:        "FULFILLED",
})
```

## Real-Time Redemptions with EventSub

Handle redemptions in real-time:

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
    ctx := context.Background()

    // Setup clients
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"

    // Create WebSocket client
    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Subscribe to all redemptions
    ws.Subscribe(ctx, helix.EventSubTypeChannelPointsRedemptionAdd, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelPointsRedemptionEvent](event)
            handleRedemption(ctx, client, e)
        },
    )

    // Subscribe to specific reward redemptions
    ws.Subscribe(ctx, helix.EventSubTypeChannelPointsRedemptionAdd, "1",
        map[string]string{
            "broadcaster_user_id": broadcasterID,
            "reward_id":           "specific-reward-id",
        },
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelPointsRedemptionEvent](event)
            fmt.Printf("Specific reward redeemed by %s\n", e.UserName)
        },
    )

    fmt.Println("Listening for redemptions...")
    select {}
}

func handleRedemption(ctx context.Context, client *helix.Client, e *helix.ChannelPointsRedemptionEvent) {
    fmt.Printf("Redemption: %s by %s\n", e.Reward.Title, e.UserName)

    switch e.Reward.Title {
    case "Song Request":
        // Process song request
        fmt.Printf("Song requested: %s\n", e.UserInput)
        // Fulfill after processing
        client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
            BroadcasterID: e.BroadcasterUserID,
            RewardID:      e.Reward.ID,
            IDs:           []string{e.ID},
            Status:        "FULFILLED",
        })

    case "Hydrate!":
        // Auto-fulfill simple rewards
        client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
            BroadcasterID: e.BroadcasterUserID,
            RewardID:      e.Reward.ID,
            IDs:           []string{e.ID},
            Status:        "FULFILLED",
        })

    default:
        fmt.Printf("Unhandled reward: %s\n", e.Reward.Title)
    }
}
```

## Reward Queue Manager

Complete example for managing a reward queue:

```go
package main

import (
    "context"
    "fmt"
    "sync"

    "github.com/Its-donkey/kappopher/helix"
)

type RewardQueue struct {
    client        *helix.Client
    broadcasterID string
    queue         []helix.ChannelPointsRedemptionEvent
    mu            sync.Mutex
}

func NewRewardQueue(client *helix.Client, broadcasterID string) *RewardQueue {
    return &RewardQueue{
        client:        client,
        broadcasterID: broadcasterID,
        queue:         make([]helix.ChannelPointsRedemptionEvent, 0),
    }
}

func (q *RewardQueue) Add(redemption helix.ChannelPointsRedemptionEvent) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.queue = append(q.queue, redemption)
    fmt.Printf("Added to queue: %s by %s (queue size: %d)\n",
        redemption.Reward.Title, redemption.UserName, len(q.queue))
}

func (q *RewardQueue) ProcessNext(ctx context.Context) *helix.ChannelPointsRedemptionEvent {
    q.mu.Lock()
    if len(q.queue) == 0 {
        q.mu.Unlock()
        return nil
    }
    redemption := q.queue[0]
    q.queue = q.queue[1:]
    q.mu.Unlock()

    return &redemption
}

func (q *RewardQueue) Fulfill(ctx context.Context, redemption *helix.ChannelPointsRedemptionEvent) error {
    _, err := q.client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
        BroadcasterID: q.broadcasterID,
        RewardID:      redemption.Reward.ID,
        IDs:           []string{redemption.ID},
        Status:        "FULFILLED",
    })
    return err
}

func (q *RewardQueue) Cancel(ctx context.Context, redemption *helix.ChannelPointsRedemptionEvent) error {
    _, err := q.client.UpdateCustomRewardRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
        BroadcasterID: q.broadcasterID,
        RewardID:      redemption.Reward.ID,
        IDs:           []string{redemption.ID},
        Status:        "CANCELED",
    })
    return err
}

func (q *RewardQueue) Size() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    return len(q.queue)
}
```

