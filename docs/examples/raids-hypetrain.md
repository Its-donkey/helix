---
layout: default
title: Raids & Hype Train Examples
description: Manage channel raids and track hype train events.
---

## Overview

**Raids**: Send your viewers to another channel at the end of your stream
- Programmatically start and cancel raids
- React to incoming raids from other channels
- Build auto-raid systems for friends

**Hype Trains**: Community engagement events triggered by bits and subs
- Track hype train progress in real-time
- Display hype train status on overlays
- React to hype train milestones (level ups)
- Support for regular, golden kappa, and shared hype trains

## Prerequisites

- **Raids:** `channel:manage:raids` scope
- **Hype Train:** `channel:read:hype_train` scope

## Start a Raid

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

    fromBroadcasterID := "12345" // Your channel
    toBroadcasterID := "67890"   // Target channel

    // Start a raid
    raid, err := client.StartRaid(ctx, &helix.StartRaidParams{
        FromBroadcasterID: fromBroadcasterID,
        ToBroadcasterID:   toBroadcasterID,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Raid started!\n")
    fmt.Printf("  Created at: %s\n", raid.Data[0].CreatedAt)
    fmt.Printf("  Raid is mature: %v\n", raid.Data[0].IsMature)
}
```

## Cancel a Raid

```go
// Cancel an in-progress raid
err := client.CancelRaid(ctx, broadcasterID)
if err != nil {
    log.Printf("Failed to cancel raid: %v", err)
} else {
    fmt.Println("Raid cancelled")
}
```

## Handle Raid Events with EventSub

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

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"

    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Subscribe to incoming raids (someone raiding you)
    ws.Subscribe(ctx, helix.EventSubTypeChannelRaid, "1",
        map[string]string{"to_broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelRaidEvent](event)
            fmt.Printf("=== INCOMING RAID ===\n")
            fmt.Printf("From: %s\n", e.FromBroadcasterUserName)
            fmt.Printf("Viewers: %d\n", e.Viewers)

            // Welcome the raiders
            welcomeRaiders(ctx, client, broadcasterID, e)
        },
    )

    // Subscribe to outgoing raids (you raiding someone)
    ws.Subscribe(ctx, helix.EventSubTypeChannelRaid, "1",
        map[string]string{"from_broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelRaidEvent](event)
            fmt.Printf("Raid sent to %s with %d viewers\n",
                e.ToBroadcasterUserName, e.Viewers)
        },
    )

    fmt.Println("Listening for raid events...")
    select {}
}

func welcomeRaiders(ctx context.Context, client *helix.Client, broadcasterID string, raid *helix.ChannelRaidEvent) {
    // Send a shoutout
    err := client.SendShoutout(ctx, &helix.SendShoutoutParams{
        FromBroadcasterID: broadcasterID,
        ToBroadcasterID:   raid.FromBroadcasterUserID,
        ModeratorID:       broadcasterID,
    })
    if err != nil {
        log.Printf("Failed to send shoutout: %v", err)
    }

    // Send welcome message
    message := fmt.Sprintf("Welcome raiders from %s! Thanks for bringing %d viewers!",
        raid.FromBroadcasterUserName, raid.Viewers)
    _, _ = client.SendChatMessage(ctx, &helix.SendChatMessageParams{
        BroadcasterID: broadcasterID,
        SenderID:      broadcasterID,
        Message:       message,
    })
}
```

## Get Hype Train Events

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

    // Get hype train events
    hypeTrain, err := client.GetHypeTrainEvents(ctx, &helix.GetHypeTrainEventsParams{
        BroadcasterID: broadcasterID,
        First:         5,
    })
    if err != nil {
        log.Fatal(err)
    }

    if len(hypeTrain.Data) == 0 {
        fmt.Println("No hype train events found")
        return
    }

    for _, event := range hypeTrain.Data {
        fmt.Printf("=== Hype Train ===\n")
        fmt.Printf("ID: %s\n", event.ID)
        fmt.Printf("Type: %s\n", event.Type) // regular, golden_kappa, shared
        fmt.Printf("Level: %d\n", event.Level)
        fmt.Printf("Total: %d\n", event.Total)
        fmt.Printf("Progress: %d / %d\n", event.Progress, event.Goal)
        fmt.Printf("Started: %s\n", event.StartedAt)
        fmt.Printf("Expires: %s\n", event.ExpiresAt)

        if event.IsSharedTrain {
            fmt.Printf("Shared train with: ")
            for _, p := range event.SharedTrainParticipants {
                fmt.Printf("%s ", p.BroadcasterUserName)
            }
            fmt.Println()
        }

        fmt.Printf("\nTop Contributors:\n")
        for _, contrib := range event.TopContributions {
            fmt.Printf("  %s: %d (%s)\n", contrib.UserName, contrib.Total, contrib.Type)
        }
        fmt.Println()
    }
}
```

## Get Current Hype Train Status

```go
// Get the current/most recent hype train status
status, err := client.GetHypeTrainStatus(ctx, broadcasterID)
if err != nil {
    log.Fatal(err)
}

if len(status.Data) > 0 {
    ht := status.Data[0]
    fmt.Printf("Current Hype Train Level: %d\n", ht.Level)
    fmt.Printf("Progress: %d/%d\n", ht.Progress, ht.Goal)
    fmt.Printf("Expires: %s\n", ht.ExpiresAt)
} else {
    fmt.Println("No active hype train")
}
```

## Track Hype Train with EventSub

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

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"

    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Subscribe to hype train start (using v2)
    ws.Subscribe(ctx, helix.EventSubTypeChannelHypeTrainBegin, helix.EventSubVersionHypeTrainV2,
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.HypeTrainBeginEvent](event)
            fmt.Printf("🚂 HYPE TRAIN STARTED!\n")
            fmt.Printf("Type: %s\n", e.Type) // regular, golden_kappa, shared
            fmt.Printf("Level: %d, Goal: %d\n", e.Level, e.Goal)
            fmt.Printf("Expires: %s\n", e.ExpiresAt)

            if e.IsSharedTrain {
                fmt.Printf("Shared with: ")
                for _, p := range e.SharedTrainParticipants {
                    fmt.Printf("%s ", p.BroadcasterUserName)
                }
                fmt.Println()
            }
        },
    )

    // Subscribe to hype train progress
    ws.Subscribe(ctx, helix.EventSubTypeChannelHypeTrainProgress, helix.EventSubVersionHypeTrainV2,
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.HypeTrainProgressEvent](event)
            progress := float64(e.Progress) / float64(e.Goal) * 100
            fmt.Printf("🚂 Level %d: %d/%d (%.1f%%)\n", e.Level, e.Progress, e.Goal, progress)
            fmt.Printf("   Total: %d\n", e.Total)

            // Show latest contribution
            if e.LastContribution.UserName != "" {
                fmt.Printf("   Latest: %s contributed %d via %s\n",
                    e.LastContribution.UserName,
                    e.LastContribution.Total,
                    e.LastContribution.Type)
            }
        },
    )

    // Subscribe to hype train end
    ws.Subscribe(ctx, helix.EventSubTypeChannelHypeTrainEnd, helix.EventSubVersionHypeTrainV2,
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.HypeTrainEndEvent](event)
            fmt.Printf("🚂 HYPE TRAIN ENDED!\n")
            fmt.Printf("Final Level: %d\n", e.Level)
            fmt.Printf("Total Points: %d\n", e.Total)
            fmt.Printf("Ended: %s\n", e.EndedAt)

            // Check if this was a new record
            if e.Level > e.AllTimeHighLevel {
                fmt.Printf("🎉 NEW ALL-TIME HIGH! Previous: Level %d\n", e.AllTimeHighLevel)
            }

            fmt.Printf("\nTop Contributors:\n")
            for i, contrib := range e.TopContributions {
                fmt.Printf("  %d. %s: %d (%s)\n", i+1, contrib.UserName, contrib.Total, contrib.Type)
            }
        },
    )

    fmt.Println("Listening for hype train events...")
    select {}
}
```

## Hype Train Overlay Example

Build a hype train display overlay:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"

    "github.com/Its-donkey/kappopher/helix"
)

type HypeTrainTracker struct {
    Active    bool
    Level     int
    Progress  int
    Goal      int
    Total     int
    Type      string
    ExpiresAt string
    TopContributions []helix.HypeTrainContribution
    mu        sync.RWMutex
}

func (h *HypeTrainTracker) Update(e *helix.HypeTrainProgressEvent) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.Active = true
    h.Level = e.Level
    h.Progress = e.Progress
    h.Goal = e.Goal
    h.Total = e.Total
    h.Type = e.Type
    h.ExpiresAt = e.ExpiresAt
    h.TopContributions = e.TopContributions
}

func (h *HypeTrainTracker) End() {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.Active = false
}

func (h *HypeTrainTracker) GetStatus() string {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if !h.Active {
        return "No active hype train"
    }

    progress := float64(h.Progress) / float64(h.Goal) * 100
    emoji := "🚂"
    if h.Type == "golden_kappa" {
        emoji = "✨🚂✨"
    } else if h.Type == "shared" {
        emoji = "🚂🤝🚂"
    }

    return fmt.Sprintf("%s HYPE TRAIN Level %d: %d/%d (%.1f%%) | Total: %d",
        emoji, h.Level, h.Progress, h.Goal, progress, h.Total)
}

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    tracker := &HypeTrainTracker{}
    broadcasterID := "12345"

    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    ws.Subscribe(ctx, helix.EventSubTypeChannelHypeTrainBegin, helix.EventSubVersionHypeTrainV2,
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.HypeTrainBeginEvent](event)
            tracker.Update(&helix.HypeTrainProgressEvent{
                Level:    e.Level,
                Progress: e.Progress,
                Goal:     e.Goal,
                Total:    e.Total,
                Type:     e.Type,
            })
            fmt.Println(tracker.GetStatus())
        },
    )

    ws.Subscribe(ctx, helix.EventSubTypeChannelHypeTrainProgress, helix.EventSubVersionHypeTrainV2,
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.HypeTrainProgressEvent](event)
            tracker.Update(e)
            fmt.Println(tracker.GetStatus())
        },
    )

    ws.Subscribe(ctx, helix.EventSubTypeChannelHypeTrainEnd, helix.EventSubVersionHypeTrainV2,
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            tracker.End()
            fmt.Println("Hype train ended!")
        },
    )

    fmt.Println("Hype train tracker running...")
    select {}
}
```

## Auto-Raid Feature

Automatically raid a friend when going offline:

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

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    myBroadcasterID := "12345"
    friendBroadcasterID := "67890" // Auto-raid target

    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Subscribe to stream offline event
    ws.Subscribe(ctx, helix.EventSubTypeStreamOffline, "1",
        map[string]string{"broadcaster_user_id": myBroadcasterID},
        func(event json.RawMessage) {
            fmt.Println("Stream went offline, checking if friend is live...")

            // Check if friend is live
            streams, err := client.GetStreams(ctx, &helix.GetStreamsParams{
                UserIDs: []string{friendBroadcasterID},
            })
            if err != nil {
                log.Printf("Failed to check friend's stream: %v", err)
                return
            }

            if len(streams.Data) > 0 {
                fmt.Printf("Friend %s is live! Starting raid...\n", streams.Data[0].UserName)

                _, err := client.StartRaid(ctx, &helix.StartRaidParams{
                    FromBroadcasterID: myBroadcasterID,
                    ToBroadcasterID:   friendBroadcasterID,
                })
                if err != nil {
                    log.Printf("Failed to start raid: %v", err)
                } else {
                    fmt.Println("Raid started!")
                }
            } else {
                fmt.Println("Friend is not live, no auto-raid")
            }
        },
    )

    fmt.Println("Auto-raid bot running...")
    select {}
}
```

