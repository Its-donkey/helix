---
layout: default
title: Ads API
description: Manage commercial breaks and ad schedules for Twitch channels.
---

## Overview

Control advertising on your Twitch channel programmatically:

**Commercial Breaks**: Run ads during your stream
- Start commercials of various lengths (30-180 seconds)
- Check cooldown before next commercial

**Ad Schedule**: Manage automatic ad scheduling
- View upcoming scheduled ads
- Check preroll-free time remaining
- Snooze scheduled ads (limited per stream)

## Prerequisites

- **Start Commercial:** `channel:edit:commercial` scope
- **Ad Schedule (read):** `channel:read:ads` scope
- **Ad Schedule (manage):** `channel:manage:ads` scope

## StartCommercial

Start a commercial break on a channel.

**Requires:** channel:edit:commercial scope

```go
resp, err := client.StartCommercial(ctx, &helix.StartCommercialParams{
    BroadcasterID: "12345",
    Length:        60, // 30, 60, 90, 120, 150, or 180 seconds
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Commercial started: %d seconds, retry after %d seconds\n",
    resp.Data.Length, resp.Data.RetryAfter)
```

**Sample Response:**
```json
{
  "data": [
    {
      "length": 60,
      "message": "",
      "retry_after": 480
    }
  ]
}
```

## GetAdSchedule

Get the broadcaster's ad schedule and details about upcoming ad breaks.

**Requires:** channel:read:ads scope

```go
resp, err := client.GetAdSchedule(ctx, &helix.GetAdScheduleParams{
    BroadcasterID: "12345",
})
if err != nil {
    log.Fatal(err)
}
for _, schedule := range resp.Data {
    fmt.Printf("Snooze count: %d/%d, Next ad: %s\n",
        schedule.SnoozeCount, schedule.SnoozeRefreshAt, schedule.NextAdAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "next_ad_at": "2025-12-11T20:00:00Z",
      "last_ad_at": "2025-12-11T19:00:00Z",
      "duration": 60,
      "preroll_free_time": 1500,
      "snooze_count": 1,
      "snooze_refresh_at": "2025-12-11T21:00:00Z"
    }
  ]
}
```

## SnoozeNextAd

Snooze the next scheduled ad break. The broadcaster can snooze a limited number of times per stream.

**Requires:** channel:manage:ads scope

```go
resp, err := client.SnoozeNextAd(ctx, &helix.SnoozeNextAdParams{
    BroadcasterID: "12345",
})
if err != nil {
    log.Fatal(err)
}
for _, schedule := range resp.Data {
    fmt.Printf("Ad snoozed! Snoozes remaining: %d, Next ad: %s\n",
        schedule.SnoozeCount, schedule.NextAdAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "snooze_count": 2,
      "snooze_refresh_at": "2025-12-11T21:00:00Z",
      "next_ad_at": "2025-12-11T20:30:00Z"
    }
  ]
}
```

