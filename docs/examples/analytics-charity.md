---
layout: default
title: Analytics, Charity & Miscellaneous Examples
description: Extension analytics, game analytics, charity campaigns, teams, and more.
---

## Overview

This guide covers various specialized Twitch features:

**Analytics**: View performance data for extensions and games you own
- Extension usage statistics and reports
- Game viewership analytics

**Charity Campaigns**: Support charitable giving through your channel
- View current campaign progress
- Track donations in real-time via EventSub

**Teams**: Groups of streamers who collaborate together
- View team membership and details
- Get team rosters

**Additional Features**:
- Guest Star: Bring guests onto your stream
- Content Classification Labels: View available content labels
- Ingest Servers: Get streaming server endpoints
- Drops Entitlements: Manage game drops
- Conduits: Manage EventSub at scale

## Prerequisites

- **Analytics:** `analytics:read:extensions` or `analytics:read:games`
- **Charity:** `channel:read:charity`
- **Teams:** No scope needed (public data)
- **Guest Star:** `channel:read:guest_star` or `channel:manage:guest_star`

## Extension Analytics

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

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Get extension analytics
    analytics, err := client.GetExtensionAnalytics(ctx, &helix.GetExtensionAnalyticsParams{
        First: 20,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Extension Analytics ===")
    for _, ext := range analytics.Data {
        fmt.Printf("\nExtension ID: %s\n", ext.ExtensionID)
        fmt.Printf("  Type: %s\n", ext.Type)
        fmt.Printf("  Date Range: %s to %s\n", ext.DateRange.StartedAt, ext.DateRange.EndedAt)
        fmt.Printf("  Report URL: %s\n", ext.URL)
    }

    // Get analytics for specific extension
    extAnalytics, err := client.GetExtensionAnalytics(ctx, &helix.GetExtensionAnalyticsParams{
        ExtensionID: "extension-id",
    })

    // Get analytics for a time period
    startTime := time.Now().AddDate(0, -1, 0) // 1 month ago
    endTime := time.Now()
    periodAnalytics, err := client.GetExtensionAnalytics(ctx, &helix.GetExtensionAnalyticsParams{
        StartedAt: &startTime,
        EndedAt:   &endTime,
        Type:      "overview_v2",
    })
}
```

## Game Analytics

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

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    // Get game analytics (for games you own)
    analytics, err := client.GetGameAnalytics(ctx, &helix.GetGameAnalyticsParams{
        First: 20,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Game Analytics ===")
    for _, game := range analytics.Data {
        fmt.Printf("\nGame ID: %s\n", game.GameID)
        fmt.Printf("  Type: %s\n", game.Type)
        fmt.Printf("  Date Range: %s to %s\n", game.DateRange.StartedAt, game.DateRange.EndedAt)
        fmt.Printf("  Report URL: %s\n", game.URL)
    }

    // Get analytics for specific game
    gameAnalytics, err := client.GetGameAnalytics(ctx, &helix.GetGameAnalyticsParams{
        GameID: "game-id",
    })

    // Get analytics for a time period
    startTime := time.Now().AddDate(0, 0, -7) // Last week
    endTime := time.Now()
    weeklyAnalytics, err := client.GetGameAnalytics(ctx, &helix.GetGameAnalyticsParams{
        GameID:    "game-id",
        StartedAt: &startTime,
        EndedAt:   &endTime,
    })
}
```

## Charity Campaigns

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

    // Get current charity campaign
    campaign, err := client.GetCharityCampaign(ctx, broadcasterID)
    if err != nil {
        log.Fatal(err)
    }

    if len(campaign.Data) == 0 {
        fmt.Println("No active charity campaign")
        return
    }

    c := campaign.Data[0]
    fmt.Printf("=== Charity Campaign ===\n")
    fmt.Printf("Charity: %s\n", c.CharityName)
    fmt.Printf("Description: %s\n", c.CharityDescription)
    fmt.Printf("Website: %s\n", c.CharityWebsite)
    fmt.Printf("Logo: %s\n", c.CharityLogo)
    fmt.Printf("\nCurrent Amount: %d %s\n", c.CurrentAmount.Value, c.CurrentAmount.Currency)
    fmt.Printf("Target Amount: %d %s\n", c.TargetAmount.Value, c.TargetAmount.Currency)

    progress := float64(c.CurrentAmount.Value) / float64(c.TargetAmount.Value) * 100
    fmt.Printf("Progress: %.1f%%\n", progress)

    // Get charity donations
    donations, err := client.GetCharityCampaignDonations(ctx, &helix.GetCharityCampaignDonationsParams{
        BroadcasterID: broadcasterID,
        First:         20,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n=== Recent Donations ===\n")
    for _, d := range donations.Data {
        fmt.Printf("%s donated %d %s\n", d.UserName, d.Amount.Value, d.Amount.Currency)
    }
}
```

## Track Charity Events with EventSub

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

    // Subscribe to charity donations
    ws.Subscribe(ctx, helix.EventSubTypeCharityDonation, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.CharityDonationEvent](event)
            fmt.Printf("💝 %s donated %d %s to %s!\n",
                e.UserName, e.Amount.Value, e.Amount.Currency, e.CharityName)
        },
    )

    // Subscribe to campaign start
    ws.Subscribe(ctx, helix.EventSubTypeCharityCampaignStart, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.CharityCampaignEvent](event)
            fmt.Printf("🎗️ Charity campaign started for %s!\n", e.CharityName)
            fmt.Printf("   Goal: %d %s\n", e.TargetAmount.Value, e.TargetAmount.Currency)
        },
    )

    // Subscribe to campaign progress
    ws.Subscribe(ctx, helix.EventSubTypeCharityCampaignProgress, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.CharityCampaignEvent](event)
            progress := float64(e.CurrentAmount.Value) / float64(e.TargetAmount.Value) * 100
            fmt.Printf("📊 Campaign progress: %d/%d %s (%.1f%%)\n",
                e.CurrentAmount.Value, e.TargetAmount.Value, e.CurrentAmount.Currency, progress)
        },
    )

    // Subscribe to campaign end
    ws.Subscribe(ctx, helix.EventSubTypeCharityCampaignStop, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.CharityCampaignEvent](event)
            fmt.Printf("🎉 Charity campaign ended!\n")
            fmt.Printf("   Final amount: %d %s\n", e.CurrentAmount.Value, e.CurrentAmount.Currency)
        },
    )

    fmt.Println("Listening for charity events...")
    select {}
}
```

## Teams

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

    // Get teams a channel belongs to
    channelTeams, err := client.GetChannelTeams(ctx, "broadcaster-id")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Channel Teams ===")
    for _, team := range channelTeams.Data {
        fmt.Printf("\n%s\n", team.TeamDisplayName)
        fmt.Printf("  ID: %s\n", team.ID)
        fmt.Printf("  Name: %s\n", team.TeamName)
        fmt.Printf("  Info: %s\n", team.Info)
        fmt.Printf("  Created: %s\n", team.CreatedAt)
    }

    // Get team details by ID
    teamByID, err := client.GetTeams(ctx, &helix.GetTeamsParams{
        ID: "team-id",
    })
    if err != nil {
        log.Fatal(err)
    }

    if len(teamByID.Data) > 0 {
        team := teamByID.Data[0]
        fmt.Printf("\n=== Team: %s ===\n", team.TeamDisplayName)
        fmt.Printf("Members: %d\n", len(team.Users))
        for _, user := range team.Users {
            fmt.Printf("  - %s\n", user.UserName)
        }
    }

    // Get team by name
    teamByName, err := client.GetTeams(ctx, &helix.GetTeamsParams{
        Name: "team-name",
    })
}
```

## Guest Star

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

    // Get Guest Star channel settings
    settings, err := client.GetChannelGuestStarSettings(ctx, broadcasterID, broadcasterID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Guest Star Settings ===")
    fmt.Printf("Enabled: %v\n", settings.Data.IsModeratorSendLiveEnabled)
    fmt.Printf("Group Layout: %s\n", settings.Data.GroupLayout)
    fmt.Printf("Regenerate Browser Sources: %v\n", settings.Data.BrowserSourceToken)

    // Get Guest Star session
    session, err := client.GetGuestStarSession(ctx, broadcasterID, broadcasterID)
    if err != nil {
        // No active session
        fmt.Println("No active Guest Star session")
    } else {
        fmt.Printf("\n=== Active Session ===\n")
        fmt.Printf("ID: %s\n", session.Data[0].ID)
        fmt.Printf("Guests:\n")
        for _, guest := range session.Data[0].Guests {
            fmt.Printf("  - %s (Slot: %s)\n", guest.UserDisplayName, guest.SlotID)
        }
    }

    // Get Guest Star invites
    invites, err := client.GetGuestStarInvites(ctx, broadcasterID, broadcasterID, "session-id")
    if err == nil {
        fmt.Printf("\n=== Pending Invites ===\n")
        for _, invite := range invites.Data {
            fmt.Printf("  - %s (Invited: %s)\n", invite.UserDisplayName, invite.InvitedAt)
        }
    }
}
```

## Content Classification Labels

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

    // Get all content classification labels
    labels, err := client.GetContentClassificationLabels(ctx, "en") // locale
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Content Classification Labels ===")
    for _, label := range labels.Data {
        fmt.Printf("\n%s (%s)\n", label.Name, label.ID)
        fmt.Printf("  Description: %s\n", label.Description)
    }
}
```

## Ingest Servers

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

    // Get ingest servers
    servers, err := client.GetIngestServers(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Ingest Servers ===")
    for _, server := range servers.Ingests {
        fmt.Printf("\n%s\n", server.Name)
        fmt.Printf("  URL Template: %s\n", server.URLTemplate)
        fmt.Printf("  Default: %v\n", server.Default)
        fmt.Printf("  Availability: %.2f\n", server.Availability)
        fmt.Printf("  Priority: %d\n", server.Priority)
    }
}
```

## Drops Entitlements

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

    // Get drops entitlements
    entitlements, err := client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
        First: 20,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Drops Entitlements ===")
    for _, ent := range entitlements.Data {
        fmt.Printf("\nEntitlement: %s\n", ent.ID)
        fmt.Printf("  User: %s\n", ent.UserID)
        fmt.Printf("  Game: %s\n", ent.GameID)
        fmt.Printf("  Benefit: %s\n", ent.BenefitID)
        fmt.Printf("  Status: %s\n", ent.FulfillmentStatus)
        fmt.Printf("  Timestamp: %s\n", ent.Timestamp)
    }

    // Get entitlements for specific user
    userEntitlements, err := client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
        UserID: "user-id",
    })

    // Get entitlements for specific game
    gameEntitlements, err := client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
        GameID: "game-id",
    })

    // Update entitlement fulfillment status
    updated, err := client.UpdateDropsEntitlements(ctx, &helix.UpdateDropsEntitlementsParams{
        EntitlementIDs:    []string{"entitlement-id-1", "entitlement-id-2"},
        FulfillmentStatus: "FULFILLED",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Updated %d entitlements\n", len(updated.Data))
}
```

## Conduits (EventSub)

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

    // Create a conduit
    conduit, err := client.CreateConduit(ctx, 5) // 5 shards
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created conduit: %s with %d shards\n", conduit.Data[0].ID, conduit.Data[0].ShardCount)

    // Get conduits
    conduits, err := client.GetConduits(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\n=== Conduits ===")
    for _, c := range conduits.Data {
        fmt.Printf("ID: %s, Shards: %d\n", c.ID, c.ShardCount)
    }

    // Get conduit shards
    shards, err := client.GetConduitShards(ctx, &helix.GetConduitShardsParams{
        ConduitID: conduit.Data[0].ID,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("\n=== Shards for %s ===\n", conduit.Data[0].ID)
    for _, shard := range shards.Data {
        fmt.Printf("Shard %s: Status=%s\n", shard.ID, shard.Status)
    }

    // Update conduit shards
    _, err = client.UpdateConduitShards(ctx, &helix.UpdateConduitShardsParams{
        ConduitID: conduit.Data[0].ID,
        Shards: []helix.ConduitShardUpdate{
            {
                ID: "0",
                Transport: helix.ConduitShardTransport{
                    Method:   "websocket",
                    SessionID: "websocket-session-id",
                },
            },
        },
    })

    // Delete conduit
    err = client.DeleteConduit(ctx, conduit.Data[0].ID)
    if err != nil {
        log.Printf("Failed to delete conduit: %v", err)
    }
}
```

