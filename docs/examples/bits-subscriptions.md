---
layout: default
title: Bits & Subscriptions Examples
description: Work with Bits leaderboards, cheermotes, and channel subscriptions.
---

## Overview

This guide covers two monetization features for Twitch channels:

**Bits**: Virtual currency viewers use to "cheer" in chat
- View bits leaderboards (top cheerers)
- Get cheermote images for different bit amounts
- React to cheer events in real-time

**Subscriptions**: Recurring monthly support from viewers
- Get subscriber list with tier information
- Check if specific users are subscribed
- React to new subscriptions, gift subs, and resubs in real-time

## Prerequisites

- **Bits:** `bits:read` scope for leaderboard
- **Subscriptions:** `channel:read:subscriptions` scope

## Bits Leaderboard

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

func main() {
    ctx := context.Background()

    // Setup client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"

    // Get all-time bits leaderboard
    leaderboard, err := client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
        Count: 10,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== All-Time Bits Leaderboard ===")
    fmt.Printf("Period: %s to %s\n", leaderboard.DateRange.StartedAt, leaderboard.DateRange.EndedAt)
    fmt.Printf("Total: %d\n\n", leaderboard.Total)

    for _, entry := range leaderboard.Data {
        fmt.Printf("%d. %s - %d bits\n", entry.Rank, entry.UserName, entry.Score)
    }

    // Get weekly leaderboard
    weeklyLeaderboard, err := client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
        Period: "week",
        Count:  10,
    })

    // Get monthly leaderboard
    monthlyLeaderboard, err := client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
        Period: "month",
        Count:  10,
    })

    // Get leaderboard for a specific time period
    startTime := time.Now().AddDate(0, -1, 0) // 1 month ago
    customLeaderboard, err := client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
        Period:    "all",
        StartedAt: startTime,
        Count:     25,
    })

    // Get specific user's rank
    userLeaderboard, err := client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
        UserID: "user-id",
    })
    if len(userLeaderboard.Data) > 0 {
        fmt.Printf("User rank: %d with %d bits\n",
            userLeaderboard.Data[0].Rank, userLeaderboard.Data[0].Score)
    }
}
```

## Cheermotes

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

    // Setup client
    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Get global cheermotes
    globalCheermotes, err := client.GetCheermotes(ctx, &helix.GetCheermotesParams{})
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Global Cheermotes ===")
    for _, cheermote := range globalCheermotes.Data {
        fmt.Printf("Prefix: %s (Type: %s)\n", cheermote.Prefix, cheermote.Type)
        fmt.Printf("  Tiers: ")
        for _, tier := range cheermote.Tiers {
            fmt.Printf("%d+ ", tier.MinBits)
        }
        fmt.Println()
    }

    // Get channel-specific cheermotes
    channelCheermotes, err := client.GetCheermotes(ctx, &helix.GetCheermotesParams{
        BroadcasterID: "12345",
    })

    // Find cheermote URLs for different tiers
    for _, cheermote := range channelCheermotes.Data {
        if cheermote.Prefix == "Cheer" {
            for _, tier := range cheermote.Tiers {
                fmt.Printf("Cheer%d: %s\n", tier.MinBits, tier.Images.Dark.Animated["1"])
            }
        }
    }
}
```

## Handle Bits with EventSub

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

    // Subscribe to bits events
    ws.Subscribe(ctx, helix.EventSubTypeChannelCheer, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelCheerEvent](event)
            handleCheer(e)
        },
    )

    fmt.Println("Listening for bits...")
    select {}
}

func handleCheer(e *helix.ChannelCheerEvent) {
    fmt.Printf("=== CHEER ===\n")
    fmt.Printf("From: %s\n", e.UserName)
    fmt.Printf("Bits: %d\n", e.Bits)
    fmt.Printf("Message: %s\n", e.Message)
    fmt.Printf("Anonymous: %v\n", e.IsAnonymous)

    // Trigger actions based on bits amount
    switch {
    case e.Bits >= 10000:
        fmt.Println("MEGA CHEER! Playing special alert...")
    case e.Bits >= 1000:
        fmt.Println("Big cheer! Playing alert...")
    case e.Bits >= 100:
        fmt.Println("Nice cheer! Playing sound...")
    default:
        fmt.Println("Thanks for the bits!")
    }
}
```

## Get Subscriptions

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

    // Get all subscribers
    subs, err := client.GetBroadcasterSubscriptions(ctx, &helix.GetBroadcasterSubscriptionsParams{
        BroadcasterID: broadcasterID,
        First:         100,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("=== Subscribers ===\n")
    fmt.Printf("Total: %d\n", subs.Total)
    fmt.Printf("Subscriber Points: %d\n\n", subs.Points)

    for _, sub := range subs.Data {
        fmt.Printf("%s (Tier %s)\n", sub.UserName, sub.Tier)
        fmt.Printf("  Gifter: %s\n", sub.GifterName)
        fmt.Printf("  Gift: %v\n", sub.IsGift)
    }

    // Check if specific user is subscribed
    userSub, err := client.CheckUserSubscription(ctx, &helix.CheckUserSubscriptionParams{
        BroadcasterID: broadcasterID,
        UserID:        "user-to-check",
    })
    if err != nil {
        fmt.Println("User is not subscribed")
    } else {
        fmt.Printf("User is subscribed at tier %s\n", userSub.Data[0].Tier)
    }

    // Get subscribers with pagination
    var allSubs []helix.BroadcasterSubscription
    cursor := ""
    for {
        resp, err := client.GetBroadcasterSubscriptions(ctx, &helix.GetBroadcasterSubscriptionsParams{
            BroadcasterID: broadcasterID,
            First:         100,
            After:         cursor,
        })
        if err != nil {
            log.Fatal(err)
        }
        allSubs = append(allSubs, resp.Data...)
        if resp.Pagination == nil || resp.Pagination.Cursor == "" {
            break
        }
        cursor = resp.Pagination.Cursor
    }
    fmt.Printf("Fetched all %d subscribers\n", len(allSubs))
}
```

## Handle Subscriptions with EventSub

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

    // Subscribe to new subscriptions
    ws.Subscribe(ctx, helix.EventSubTypeChannelSubscribe, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelSubscribeEvent](event)
            fmt.Printf("New sub: %s (Tier %s, Gift: %v)\n", e.UserName, e.Tier, e.IsGift)
        },
    )

    // Subscribe to gift subs
    ws.Subscribe(ctx, helix.EventSubTypeChannelSubscriptionGift, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelSubscriptionGiftEvent](event)
            fmt.Printf("%s gifted %d subs! (Tier %s, Total: %d)\n",
                e.UserName, e.Total, e.Tier, e.CumulativeTotal)
        },
    )

    // Subscribe to resub messages
    ws.Subscribe(ctx, helix.EventSubTypeChannelSubscriptionMessage, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelSubscriptionMessageEvent](event)
            fmt.Printf("Resub: %s - %d months! Message: %s\n",
                e.UserName, e.CumulativeMonths, e.Message.Text)
            if e.StreakMonths > 0 {
                fmt.Printf("  Streak: %d months\n", e.StreakMonths)
            }
        },
    )

    // Subscribe to sub ends
    ws.Subscribe(ctx, helix.EventSubTypeChannelSubscriptionEnd, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelSubscriptionEndEvent](event)
            fmt.Printf("Sub ended: %s (Tier %s, Gift: %v)\n", e.UserName, e.Tier, e.IsGift)
        },
    )

    fmt.Println("Listening for subscription events...")
    select {}
}
```

## Subscription Tier Analysis

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

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"

    // Get all subscribers
    var allSubs []helix.BroadcasterSubscription
    cursor := ""
    for {
        resp, err := client.GetBroadcasterSubscriptions(ctx, &helix.GetBroadcasterSubscriptionsParams{
            BroadcasterID: broadcasterID,
            First:         100,
            After:         cursor,
        })
        if err != nil {
            log.Fatal(err)
        }
        allSubs = append(allSubs, resp.Data...)
        if resp.Pagination == nil || resp.Pagination.Cursor == "" {
            break
        }
        cursor = resp.Pagination.Cursor
    }

    // Analyze tiers
    tiers := map[string]int{"1000": 0, "2000": 0, "3000": 0}
    gifted := 0

    for _, sub := range allSubs {
        tiers[sub.Tier]++
        if sub.IsGift {
            gifted++
        }
    }

    fmt.Println("=== Subscription Analysis ===")
    fmt.Printf("Total Subscribers: %d\n", len(allSubs))
    fmt.Printf("Tier 1 (1000): %d (%.1f%%)\n", tiers["1000"], float64(tiers["1000"])/float64(len(allSubs))*100)
    fmt.Printf("Tier 2 (2000): %d (%.1f%%)\n", tiers["2000"], float64(tiers["2000"])/float64(len(allSubs))*100)
    fmt.Printf("Tier 3 (3000): %d (%.1f%%)\n", tiers["3000"], float64(tiers["3000"])/float64(len(allSubs))*100)
    fmt.Printf("Gifted: %d (%.1f%%)\n", gifted, float64(gifted)/float64(len(allSubs))*100)

    // Calculate estimated monthly revenue (rough estimate)
    // Tier 1 = $4.99, Tier 2 = $9.99, Tier 3 = $24.99 (50% split assumed)
    revenue := float64(tiers["1000"])*2.50 + float64(tiers["2000"])*5.00 + float64(tiers["3000"])*12.50
    fmt.Printf("Estimated Monthly Revenue: $%.2f\n", revenue)
}
```

