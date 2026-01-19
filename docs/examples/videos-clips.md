---
layout: default
title: Videos & Clips Examples
description: Work with VODs, highlights, and clips.
---

## Overview

Manage video content on Twitch, including:

**Videos (VODs)**: Recorded broadcasts, highlights, and uploads
- Retrieve video metadata (title, duration, views)
- Delete videos from your channel
- Filter by type (archive, highlight, upload), game, and time period

**Clips**: Short highlight moments from streams
- Create clips from live streams
- Create clips from VODs at specific timestamps
- Retrieve clips by broadcaster, game, or clip ID
- Build automated clip compilation systems

**Stream Markers**: Timestamps you can place during a stream for VOD editing
- Create markers with descriptions
- Retrieve markers for a video

## Prerequisites

- **Videos:** `channel:manage:videos` to delete, no scope needed to read
- **Clips:** `clips:edit` to create clips

## Get Videos (VODs)

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

    // Get videos by user ID
    videos, err := client.GetVideos(ctx, &helix.GetVideosParams{
        UserID: "12345",
        First:  10,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Recent Videos ===")
    for _, video := range videos.Data {
        fmt.Printf("\n%s\n", video.Title)
        fmt.Printf("  ID: %s\n", video.ID)
        fmt.Printf("  Type: %s\n", video.Type) // archive, highlight, upload
        fmt.Printf("  Duration: %s\n", video.Duration)
        fmt.Printf("  Views: %d\n", video.ViewCount)
        fmt.Printf("  Created: %s\n", video.CreatedAt)
        fmt.Printf("  URL: %s\n", video.URL)
    }

    // Get specific videos by ID
    specificVideos, err := client.GetVideos(ctx, &helix.GetVideosParams{
        IDs: []string{"video-id-1", "video-id-2"},
    })

    // Get videos by game
    gameVideos, err := client.GetVideos(ctx, &helix.GetVideosParams{
        GameID: "509658", // Just Chatting
        First:  20,
        Sort:   "views", // time, trending, views
        Type:   "all",   // all, archive, highlight, upload
    })

    // Filter by language and period
    filteredVideos, err := client.GetVideos(ctx, &helix.GetVideosParams{
        GameID:   "509658",
        Language: "en",
        Period:   "week", // all, day, week, month
    })
}
```

## Delete Videos

```go
// Delete specific videos
deleted, err := client.DeleteVideos(ctx, []string{"video-id-1", "video-id-2"})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deleted %d videos\n", len(deleted.Data))
for _, id := range deleted.Data {
    fmt.Printf("  Deleted: %s\n", id)
}
```

## Get Clips

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

    broadcasterID := "12345"

    // Get clips by broadcaster
    clips, err := client.GetClips(ctx, &helix.GetClipsParams{
        BroadcasterID: broadcasterID,
        First:         20,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("=== Top Clips ===")
    for _, clip := range clips.Data {
        fmt.Printf("\n%s\n", clip.Title)
        fmt.Printf("  ID: %s\n", clip.ID)
        fmt.Printf("  Creator: %s\n", clip.CreatorName)
        fmt.Printf("  Views: %d\n", clip.ViewCount)
        fmt.Printf("  Duration: %.1fs\n", clip.Duration)
        fmt.Printf("  Created: %s\n", clip.CreatedAt)
        fmt.Printf("  URL: %s\n", clip.URL)
        fmt.Printf("  Thumbnail: %s\n", clip.ThumbnailURL)
    }

    // Get clips by game
    gameClips, err := client.GetClips(ctx, &helix.GetClipsParams{
        GameID: "509658",
        First:  10,
    })

    // Get specific clips by ID
    specificClips, err := client.GetClips(ctx, &helix.GetClipsParams{
        IDs: []string{"clip-id-1", "clip-id-2"},
    })

    // Get clips from a time range
    startTime := time.Now().AddDate(0, 0, -7) // Last 7 days
    endTime := time.Now()
    recentClips, err := client.GetClips(ctx, &helix.GetClipsParams{
        BroadcasterID: broadcasterID,
        StartedAt:     &startTime,
        EndedAt:       &endTime,
        First:         50,
    })

    // Get featured clips
    featuredClips, err := client.GetClips(ctx, &helix.GetClipsParams{
        BroadcasterID: broadcasterID,
        IsFeatured:    true,
    })
}
```

## Create Clips

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

    broadcasterID := "12345"

    // Create a clip from live stream
    clip, err := client.CreateClip(ctx, &helix.CreateClipParams{
        BroadcasterID: broadcasterID,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Clip created!\n")
    fmt.Printf("  ID: %s\n", clip.Data[0].ID)
    fmt.Printf("  Edit URL: %s\n", clip.Data[0].EditURL)

    // Wait for clip to be processed
    fmt.Println("Waiting for clip to process...")
    time.Sleep(15 * time.Second)

    // Get the processed clip
    processedClip, err := client.GetClips(ctx, &helix.GetClipsParams{
        IDs: []string{clip.Data[0].ID},
    })
    if err == nil && len(processedClip.Data) > 0 {
        fmt.Printf("Clip ready: %s\n", processedClip.Data[0].URL)
    }

    // Create clip with delay (captures content from 30 seconds ago)
    delayedClip, err := client.CreateClip(ctx, &helix.CreateClipParams{
        BroadcasterID: broadcasterID,
        HasDelay:      true, // Adds a delay to capture content from slightly before
    })
}
```

## Create Clip from VOD

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

    // Create a clip from a VOD
    clip, err := client.CreateClipFromVOD(ctx, &helix.CreateClipFromVODParams{
        VideoID:      "video-id",
        OffsetSeconds: 3600,  // Start at 1 hour mark
        Duration:      30,    // 30 second clip (5-60 seconds allowed)
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("VOD clip created!\n")
    fmt.Printf("  ID: %s\n", clip.Data[0].ID)
    fmt.Printf("  Edit URL: %s\n", clip.Data[0].EditURL)
}
```

## Stream Markers

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

    // Create a stream marker
    marker, err := client.CreateStreamMarker(ctx, &helix.CreateStreamMarkerParams{
        UserID:      broadcasterID,
        Description: "Epic moment!",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Marker created at %d seconds\n", marker.Data[0].PositionSeconds)

    // Get stream markers
    markers, err := client.GetStreamMarkers(ctx, &helix.GetStreamMarkersParams{
        UserID: broadcasterID,
        First:  20,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, video := range markers.Data {
        fmt.Printf("\nVideo: %s\n", video.VideoID)
        for _, m := range video.Markers {
            fmt.Printf("  [%d:%02d] %s (ID: %s)\n",
                m.PositionSeconds/60, m.PositionSeconds%60,
                m.Description, m.ID)
        }
    }

    // Get markers for specific video
    videoMarkers, err := client.GetStreamMarkers(ctx, &helix.GetStreamMarkersParams{
        VideoID: "video-id",
    })
}
```

## Clip Compilation Bot

Auto-clip highlights based on chat activity:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/Its-donkey/kappopher/helix"
)

type ClipBot struct {
    client        *helix.Client
    broadcasterID string
    chatActivity  int
    threshold     int
    cooldown      time.Duration
    lastClip      time.Time
    mu            sync.Mutex
}

func NewClipBot(client *helix.Client, broadcasterID string) *ClipBot {
    return &ClipBot{
        client:        client,
        broadcasterID: broadcasterID,
        threshold:     50,            // Clip when 50 messages in 10 seconds
        cooldown:      30 * time.Second,
    }
}

func (b *ClipBot) TrackMessage() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.chatActivity++
}

func (b *ClipBot) CheckAndClip(ctx context.Context) {
    b.mu.Lock()
    activity := b.chatActivity
    b.chatActivity = 0 // Reset counter
    canClip := time.Since(b.lastClip) > b.cooldown
    b.mu.Unlock()

    if activity >= b.threshold && canClip {
        fmt.Printf("High activity detected (%d messages)! Creating clip...\n", activity)

        clip, err := b.client.CreateClip(ctx, &helix.CreateClipParams{
            BroadcasterID: b.broadcasterID,
            HasDelay:      true, // Capture the moment that just happened
        })
        if err != nil {
            log.Printf("Failed to create clip: %v", err)
            return
        }

        b.mu.Lock()
        b.lastClip = time.Now()
        b.mu.Unlock()

        fmt.Printf("Clip created: %s\n", clip.Data[0].EditURL)
    }
}

func main() {
    ctx := context.Background()

    authClient := helix.NewAuthClient(helix.AuthConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
    })
    client := helix.NewClient("your-client-id", authClient)

    broadcasterID := "12345"
    bot := NewClipBot(client, broadcasterID)

    ws := helix.NewEventSubWebSocket(client)
    if err := ws.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Track chat messages
    ws.Subscribe(ctx, helix.EventSubTypeChannelChatMessage, "1",
        map[string]string{
            "broadcaster_user_id": broadcasterID,
            "user_id":             broadcasterID,
        },
        func(event json.RawMessage) {
            bot.TrackMessage()
        },
    )

    // Check activity every 10 seconds
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            bot.CheckAndClip(ctx)
        }
    }()

    fmt.Println("Clip bot running...")
    select {}
}
```

## Video Stats Dashboard

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sort"

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

    // Get all videos
    var allVideos []helix.Video
    cursor := ""
    for {
        resp, err := client.GetVideos(ctx, &helix.GetVideosParams{
            UserID: broadcasterID,
            First:  100,
            After:  cursor,
        })
        if err != nil {
            log.Fatal(err)
        }
        allVideos = append(allVideos, resp.Data...)
        if resp.Pagination == nil || resp.Pagination.Cursor == "" {
            break
        }
        cursor = resp.Pagination.Cursor
    }

    // Get all clips
    var allClips []helix.Clip
    cursor = ""
    for {
        resp, err := client.GetClips(ctx, &helix.GetClipsParams{
            BroadcasterID: broadcasterID,
            First:         100,
            After:         cursor,
        })
        if err != nil {
            log.Fatal(err)
        }
        allClips = append(allClips, resp.Data...)
        if resp.Pagination == nil || resp.Pagination.Cursor == "" {
            break
        }
        cursor = resp.Pagination.Cursor
    }

    // Calculate stats
    fmt.Println("=== Video & Clip Statistics ===\n")

    // Video stats
    totalVideoViews := 0
    videoTypes := map[string]int{}
    for _, v := range allVideos {
        totalVideoViews += v.ViewCount
        videoTypes[v.Type]++
    }

    fmt.Printf("Videos: %d\n", len(allVideos))
    fmt.Printf("  Archives: %d\n", videoTypes["archive"])
    fmt.Printf("  Highlights: %d\n", videoTypes["highlight"])
    fmt.Printf("  Uploads: %d\n", videoTypes["upload"])
    fmt.Printf("  Total Views: %d\n", totalVideoViews)

    // Top videos
    sort.Slice(allVideos, func(i, j int) bool {
        return allVideos[i].ViewCount > allVideos[j].ViewCount
    })
    fmt.Printf("\nTop 5 Videos:\n")
    for i := 0; i < 5 && i < len(allVideos); i++ {
        fmt.Printf("  %d. %s (%d views)\n", i+1, allVideos[i].Title, allVideos[i].ViewCount)
    }

    // Clip stats
    totalClipViews := 0
    for _, c := range allClips {
        totalClipViews += c.ViewCount
    }

    fmt.Printf("\nClips: %d\n", len(allClips))
    fmt.Printf("  Total Views: %d\n", totalClipViews)

    // Top clips
    sort.Slice(allClips, func(i, j int) bool {
        return allClips[i].ViewCount > allClips[j].ViewCount
    })
    fmt.Printf("\nTop 5 Clips:\n")
    for i := 0; i < 5 && i < len(allClips); i++ {
        fmt.Printf("  %d. %s (%d views) - by %s\n",
            i+1, allClips[i].Title, allClips[i].ViewCount, allClips[i].CreatorName)
    }
}
```

