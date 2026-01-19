---
layout: default
title: Schedule & Goals Examples
description: Manage stream schedules and track creator goals.
---

## Overview

This guide covers two related features for channel management:

**Stream Schedule**: Allow viewers to see when you plan to stream
- Create one-time and recurring stream segments
- Set vacation mode when taking breaks
- Export schedule as iCalendar for calendar apps

**Creator Goals**: Track progress toward milestones
- Follower goals, subscriber goals, etc.
- Real-time progress updates via EventSub
- Display goal progress on overlays

## Prerequisites

- **Schedule:** `channel:manage:schedule` to modify, `channel:read:schedule` to read
- **Goals:** `channel:read:goals` scope

## Get Stream Schedule

Retrieve a channel's stream schedule including upcoming segments, vacation status, and category information. You can also export the schedule as iCalendar format for integration with calendar applications.

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

    // Get full schedule
    schedule, err := client.GetChannelStreamSchedule(ctx, &helix.GetChannelStreamScheduleParams{
        BroadcasterID: broadcasterID,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("=== Schedule for %s ===\n", schedule.Data.BroadcasterName)

    if schedule.Data.Vacation != nil {
        fmt.Printf("On vacation: %s to %s\n",
            schedule.Data.Vacation.StartTime, schedule.Data.Vacation.EndTime)
    }

    for _, segment := range schedule.Data.Segments {
        fmt.Printf("\n%s\n", segment.Title)
        fmt.Printf("  Start: %s\n", segment.StartTime.Format(time.RFC1123))
        fmt.Printf("  End: %s\n", segment.EndTime.Format(time.RFC1123))
        if segment.Category != nil {
            fmt.Printf("  Category: %s\n", segment.Category.Name)
        }
        fmt.Printf("  Recurring: %v\n", segment.IsRecurring)
        if segment.CanceledUntil != nil {
            fmt.Printf("  Canceled until: %s\n", segment.CanceledUntil)
        }
    }

    // Get schedule with time filter
    startTime := time.Now()
    filteredSchedule, err := client.GetChannelStreamSchedule(ctx, &helix.GetChannelStreamScheduleParams{
        BroadcasterID: broadcasterID,
        StartTime:     &startTime,
        First:         10,
    })

    // Get schedule as iCalendar
    ical, err := client.GetChannelICalendar(ctx, broadcasterID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\niCalendar data:\n%s\n", ical)
}
```

## Update Schedule Settings

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

    // Update schedule settings (vacation mode)
    vacationStart := time.Now().AddDate(0, 0, 7)  // Start in 1 week
    vacationEnd := time.Now().AddDate(0, 0, 14)   // End in 2 weeks

    err := client.UpdateChannelStreamSchedule(ctx, &helix.UpdateChannelStreamScheduleParams{
        BroadcasterID:     broadcasterID,
        IsVacationEnabled: boolPtr(true),
        VacationStartTime: &vacationStart,
        VacationEndTime:   &vacationEnd,
        Timezone:          "America/New_York",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Vacation mode enabled!")

    // Disable vacation mode
    err = client.UpdateChannelStreamSchedule(ctx, &helix.UpdateChannelStreamScheduleParams{
        BroadcasterID:     broadcasterID,
        IsVacationEnabled: boolPtr(false),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Vacation mode disabled!")
}

func boolPtr(b bool) *bool { return &b }
```

## Create Schedule Segments

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

    // Create a one-time stream
    startTime := time.Now().AddDate(0, 0, 1).Truncate(time.Hour) // Tomorrow at the top of the hour
    segment, err := client.CreateScheduleSegment(ctx, &helix.CreateScheduleSegmentParams{
        BroadcasterID: broadcasterID,
        StartTime:     startTime,
        Timezone:      "America/New_York",
        Duration:      180, // 3 hours in minutes
        IsRecurring:   false,
        Title:         "Special Event Stream",
        CategoryID:    "509658", // Just Chatting
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created segment: %s (ID: %s)\n", segment.Data.Segments[0].Title, segment.Data.Segments[0].ID)

    // Create a recurring weekly stream
    weeklyStart := time.Date(2024, 1, 15, 19, 0, 0, 0, time.UTC) // Monday 7 PM
    recurringSegment, err := client.CreateScheduleSegment(ctx, &helix.CreateScheduleSegmentParams{
        BroadcasterID: broadcasterID,
        StartTime:     weeklyStart,
        Timezone:      "America/New_York",
        Duration:      240, // 4 hours
        IsRecurring:   true,
        Title:         "Weekly Game Night",
        CategoryID:    "509658",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created recurring segment: %s\n", recurringSegment.Data.Segments[0].Title)
}
```

## Update Schedule Segments

```go
// Update a segment
updated, err := client.UpdateScheduleSegment(ctx, &helix.UpdateScheduleSegmentParams{
    BroadcasterID: broadcasterID,
    ID:            "segment-id",
    Title:         "Updated Stream Title",
    CategoryID:    "26936", // Music
    Duration:      120,
    IsCanceled:    false,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Updated segment: %s\n", updated.Data.Segments[0].Title)

// Cancel a segment (keeps it in schedule but marks as canceled)
_, err = client.UpdateScheduleSegment(ctx, &helix.UpdateScheduleSegmentParams{
    BroadcasterID: broadcasterID,
    ID:            "segment-id",
    IsCanceled:    true,
})

// Reschedule a segment
newStartTime := time.Now().AddDate(0, 0, 3)
_, err = client.UpdateScheduleSegment(ctx, &helix.UpdateScheduleSegmentParams{
    BroadcasterID: broadcasterID,
    ID:            "segment-id",
    StartTime:     &newStartTime,
    Timezone:      "America/New_York",
})
```

## Delete Schedule Segments

```go
// Delete a single segment
err := client.DeleteScheduleSegment(ctx, broadcasterID, "segment-id")
if err != nil {
    log.Printf("Failed to delete segment: %v", err)
}

// Delete all future instances of a recurring segment
// (Delete the segment itself to remove all future occurrences)
```

## Get Creator Goals

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

    // Get all active goals
    goals, err := client.GetCreatorGoals(ctx, broadcasterID)
    if err != nil {
        log.Fatal(err)
    }

    if len(goals.Data) == 0 {
        fmt.Println("No active goals")
        return
    }

    fmt.Println("=== Creator Goals ===")
    for _, goal := range goals.Data {
        fmt.Printf("\n%s\n", goal.Description)
        fmt.Printf("  Type: %s\n", goal.Type)
        fmt.Printf("  Progress: %d / %d (%.1f%%)\n",
            goal.CurrentAmount, goal.TargetAmount,
            float64(goal.CurrentAmount)/float64(goal.TargetAmount)*100)
        fmt.Printf("  Created: %s\n", goal.CreatedAt)
    }
}
```

## Track Goals with EventSub

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

    // Subscribe to goal start
    ws.Subscribe(ctx, helix.EventSubTypeChannelGoalBegin, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelGoalEvent](event)
            fmt.Printf("New goal started: %s\n", e.Description)
            fmt.Printf("  Type: %s, Target: %d\n", e.Type, e.TargetAmount)
        },
    )

    // Subscribe to goal progress
    ws.Subscribe(ctx, helix.EventSubTypeChannelGoalProgress, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelGoalEvent](event)
            progress := float64(e.CurrentAmount) / float64(e.TargetAmount) * 100
            fmt.Printf("Goal progress: %d/%d (%.1f%%) - %s\n",
                e.CurrentAmount, e.TargetAmount, progress, e.Description)
        },
    )

    // Subscribe to goal completion
    ws.Subscribe(ctx, helix.EventSubTypeChannelGoalEnd, "1",
        map[string]string{"broadcaster_user_id": broadcasterID},
        func(event json.RawMessage) {
            e, _ := helix.ParseWSEvent[helix.ChannelGoalEvent](event)
            if e.IsAchieved {
                fmt.Printf("GOAL ACHIEVED: %s!\n", e.Description)
            } else {
                fmt.Printf("Goal ended (not achieved): %s\n", e.Description)
            }
        },
    )

    fmt.Println("Listening for goal events...")
    select {}
}
```

## Schedule Widget Example

Build a simple schedule display:

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

    // Get upcoming streams
    startTime := time.Now()
    schedule, err := client.GetChannelStreamSchedule(ctx, &helix.GetChannelStreamScheduleParams{
        BroadcasterID: broadcasterID,
        StartTime:     &startTime,
        First:         7, // Next 7 segments
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("=== Upcoming Streams for %s ===\n\n", schedule.Data.BroadcasterName)

    if schedule.Data.Vacation != nil {
        fmt.Printf("⚠️  On vacation until %s\n\n",
            schedule.Data.Vacation.EndTime.Format("January 2, 2006"))
    }

    for _, segment := range schedule.Data.Segments {
        if segment.CanceledUntil != nil {
            continue // Skip canceled segments
        }

        // Format the date nicely
        day := segment.StartTime.Format("Monday, Jan 2")
        startTimeStr := segment.StartTime.Format("3:04 PM")
        endTimeStr := segment.EndTime.Format("3:04 PM")

        category := "No category"
        if segment.Category != nil {
            category = segment.Category.Name
        }

        recurring := ""
        if segment.IsRecurring {
            recurring = " 🔁"
        }

        fmt.Printf("📅 %s%s\n", day, recurring)
        fmt.Printf("   %s - %s\n", startTimeStr, endTimeStr)
        fmt.Printf("   📺 %s\n", segment.Title)
        fmt.Printf("   🎮 %s\n\n", category)
    }
}
```

