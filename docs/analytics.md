---
layout: default
title: Analytics API
description: Retrieve analytics reports for extensions and games, including downloadable CSV reports.
---

## GetExtensionAnalytics

Get analytics reports for one or more extensions. The response contains URLs to download the reports as CSV files.

**Requires:** `analytics:read:extensions`

```go
resp, err := client.GetExtensionAnalytics(ctx, &helix.GetExtensionAnalyticsParams{
    ExtensionID: "your-extension-id",
    Type:        "overview_v2",
    StartedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    EndedAt:     time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})
for _, report := range resp.Data {
    fmt.Printf("Report for extension %s: %s\n", report.ExtensionID, report.URL)
    fmt.Printf("Date range: %s to %s\n", report.StartedAt, report.EndedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "extension_id": "abcd1234efgh5678",
      "url": "https://twitch-piper-reports.s3-us-west-2.amazonaws.com/extensions/overview/v2/abcd1234-5678-90ab-cdef-1234567890ab.csv",
      "type": "overview_v2",
      "date_range": {
        "started_at": "2024-01-01T00:00:00Z",
        "ended_at": "2024-01-31T23:59:59Z"
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## GetGameAnalytics

Get analytics reports for one or more games. The response contains URLs to download the reports as CSV files.

**Requires:** `analytics:read:games`

```go
resp, err := client.GetGameAnalytics(ctx, &helix.GetGameAnalyticsParams{
    GameID:    "493057", // Game ID for PUBG
    Type:      "overview_v2",
    StartedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    EndedAt:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})
for _, report := range resp.Data {
    fmt.Printf("Report for game %s: %s\n", report.GameID, report.URL)
    fmt.Printf("Date range: %s to %s\n", report.StartedAt, report.EndedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "game_id": "493057",
      "url": "https://twitch-piper-reports.s3-us-west-2.amazonaws.com/games/overview/v2/1234abcd-5678-90ef-ghij-klmnopqrstuv.csv",
      "type": "overview_v2",
      "date_range": {
        "started_at": "2024-01-01T00:00:00Z",
        "ended_at": "2024-01-31T23:59:59Z"
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

