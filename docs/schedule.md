---
layout: default
title: Schedule API
description: Manage Twitch channel stream schedules.
---

## GetChannelStreamSchedule

Get the broadcaster's stream schedule.

**Requires:** No authentication required

```go
// Get all scheduled segments
resp, err := client.GetChannelStreamSchedule(ctx, &helix.GetChannelStreamScheduleParams{
    BroadcasterID: "12345",
})

// Get specific segments by ID (max 100)
resp, err = client.GetChannelStreamSchedule(ctx, &helix.GetChannelStreamScheduleParams{
    BroadcasterID: "12345",
    IDs:           []string{"segment-id-1", "segment-id-2"},
})

// Get segments starting from a specific time with UTC offset
resp, err = client.GetChannelStreamSchedule(ctx, &helix.GetChannelStreamScheduleParams{
    BroadcasterID: "12345",
    StartTime:     "2025-12-15T10:00:00Z",
    UTCOffset:     "-04:00", // e.g., Eastern Daylight Time
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

// Access schedule information
schedule := resp.Data.Schedule
fmt.Printf("Broadcaster: %s (%s)\n", schedule.BroadcasterName, schedule.BroadcasterLogin)
if schedule.Vacation != nil {
    fmt.Printf("Vacation: %s to %s\n", schedule.Vacation.StartTime, schedule.Vacation.EndTime)
}
for _, segment := range schedule.Segments {
    fmt.Printf("Segment: %s at %s (Duration: %d minutes)\n",
        segment.Title, segment.StartTime, segment.Duration)
}
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `IDs` - Optional. Array of segment IDs to retrieve (max 100)
- `StartTime` - Optional. Start time for segments (RFC3339 format)
- `UTCOffset` - Optional. UTC offset for localization (e.g., "-04:00")
- `PaginationParams` - Optional. Pagination parameters

**Response Fields:**
- `Schedule` - The schedule object containing:
  - `BroadcasterID`, `BroadcasterName`, `BroadcasterLogin` - Broadcaster information
  - `Segments` - Array of scheduled stream segments
  - `Vacation` - Vacation information (if enabled)

**Sample Response:**
```json
{
  "data": {
    "segments": [
      {
        "id": "eyJzZWdtZW50SUQiOiI1NjkyNmI0ZC0xNGJmLTRjYjEtOGU3ZS00ZjBjMDNiZjgwZWYiLCJpc29ZZWFyIjoyMDI1LCJpc29XZWVrIjo1MH0=",
        "start_time": "2025-12-15T20:00:00Z",
        "end_time": "2025-12-15T23:00:00Z",
        "title": "Coding stream - Building a Twitch bot",
        "canceled_until": null,
        "category": {
          "id": "509658",
          "name": "Science & Technology"
        },
        "is_recurring": false
      },
      {
        "id": "eyJzZWdtZW50SUQiOiI3NmEyYmU4Yy0yNWJmLTRjYjEtOGU3ZS00ZjBjMDNiZjgwZWYiLCJpc29ZZWFyIjoyMDI1LCJpc29XZWVrIjo1MX0=",
        "start_time": "2025-12-22T20:00:00Z",
        "end_time": "2025-12-22T22:00:00Z",
        "title": "Weekly coding stream",
        "canceled_until": null,
        "category": {
          "id": "509658",
          "name": "Science & Technology"
        },
        "is_recurring": true
      }
    ],
    "broadcaster_id": "141981764",
    "broadcaster_name": "TwitchDev",
    "broadcaster_login": "twitchdev",
    "vacation": {
      "start_time": "2025-12-20T00:00:00Z",
      "end_time": "2025-12-27T00:00:00Z"
    }
  },
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik1USTROekUxTmpnME9EZzVNakV3TURBd01EQXdNREF3SW4wPSJ9fQ"
  }
}
```

## GetChannelICalendar

Get the broadcaster's stream schedule as an iCalendar string.

**Requires:** No authentication required

```go
ical, err := client.GetChannelICalendar(ctx, "12345")
if err != nil {
    fmt.Printf("Failed to get iCalendar: %v\n", err)
}
fmt.Println(ical) // iCalendar format string
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster

**Returns:** iCalendar format string

**Sample Response:**
```
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Twitch//Stream Schedule 1.0//EN
BEGIN:VEVENT
DTSTART:20251215T200000Z
DTEND:20251215T230000Z
SUMMARY:Coding stream - Building a Twitch bot
DESCRIPTION:Category: Science & Technology
UID:eyJzZWdtZW50SUQiOiI1NjkyNmI0ZC0xNGJmLTRjYjEtOGU3ZS00ZjBjMDNiZjgwZWYiLCJpc29ZZWFyIjoyMDI1LCJpc29XZWVrIjo1MH0=
END:VEVENT
BEGIN:VEVENT
DTSTART:20251222T200000Z
DTEND:20251222T220000Z
SUMMARY:Weekly coding stream
DESCRIPTION:Category: Science & Technology
RRULE:FREQ=WEEKLY
UID:eyJzZWdtZW50SUQiOiI3NmEyYmU4Yy0yNWJmLTRjYjEtOGU3ZS00ZjBjMDNiZjgwZWYiLCJpc29ZZWFyIjoyMDI1LCJpc29XZWVrIjo1MX0=
END:VEVENT
END:VCALENDAR
```

## UpdateChannelStreamSchedule

Update the broadcaster's stream schedule settings (e.g., vacation mode).

**Requires:** `channel:manage:schedule`

```go
// Enable vacation mode
err := client.UpdateChannelStreamSchedule(ctx, &helix.UpdateChannelStreamScheduleParams{
    BroadcasterID:      "12345",
    IsVacationEnabled:  true,
    VacationStartTime:  "2025-12-20T00:00:00Z",
    VacationEndTime:    "2025-12-27T00:00:00Z",
    Timezone:           "America/New_York",
})

// Disable vacation mode
err = client.UpdateChannelStreamSchedule(ctx, &helix.UpdateChannelStreamScheduleParams{
    BroadcasterID:     "12345",
    IsVacationEnabled: false,
})
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `IsVacationEnabled` - Optional. Enable or disable vacation mode
- `VacationStartTime` - Optional. Vacation start time (RFC3339 format)
- `VacationEndTime` - Optional. Vacation end time (RFC3339 format)
- `Timezone` - Optional. Timezone (IANA format, e.g., "America/New_York")

**Sample Response:**
This endpoint returns no content on success (HTTP 204 No Content).

## CreateChannelStreamScheduleSegment

Create a new scheduled stream segment.

**Requires:** `channel:manage:schedule`

```go
resp, err := client.CreateChannelStreamScheduleSegment(ctx, &helix.CreateChannelStreamScheduleSegmentParams{
    BroadcasterID: "12345",
    StartTime:     "2025-12-15T20:00:00Z",
    Timezone:      "America/New_York",
    Duration:      180, // 3 hours (30-1380 minutes)
    IsRecurring:   false,
    CategoryID:    "509658", // Science & Technology
    Title:         "Coding stream - Building a Twitch bot",
})

// Create a recurring segment
resp, err = client.CreateChannelStreamScheduleSegment(ctx, &helix.CreateChannelStreamScheduleSegmentParams{
    BroadcasterID: "12345",
    StartTime:     "2025-12-15T20:00:00Z",
    Timezone:      "America/New_York",
    Duration:      120, // 2 hours
    IsRecurring:   true, // Repeats weekly
    CategoryID:    "509658",
    Title:         "Weekly coding stream",
})

segment := resp.Data.Segments[0]
fmt.Printf("Created segment: %s (ID: %s)\n", segment.Title, segment.ID)
fmt.Printf("Starts at: %s, Duration: %d minutes\n", segment.StartTime, segment.Duration)
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `StartTime` - Required. Start time (RFC3339 format)
- `Timezone` - Required. Timezone (IANA format)
- `Duration` - Required. Duration in minutes (30-1380)
- `IsRecurring` - Optional. Whether the segment repeats weekly
- `CategoryID` - Optional. Game/category ID
- `Title` - Optional. Segment title

**Sample Response:**
```json
{
  "data": {
    "segments": [
      {
        "id": "eyJzZWdtZW50SUQiOiI1NjkyNmI0ZC0xNGJmLTRjYjEtOGU3ZS00ZjBjMDNiZjgwZWYiLCJpc29ZZWFyIjoyMDI1LCJpc29XZWVrIjo1MH0=",
        "start_time": "2025-12-15T20:00:00Z",
        "end_time": "2025-12-15T23:00:00Z",
        "title": "Coding stream - Building a Twitch bot",
        "canceled_until": null,
        "category": {
          "id": "509658",
          "name": "Science & Technology"
        },
        "is_recurring": false
      }
    ]
  }
}
```

## UpdateChannelStreamScheduleSegment

Update an existing scheduled stream segment.

**Requires:** `channel:manage:schedule`

```go
// Update segment details
resp, err := client.UpdateChannelStreamScheduleSegment(ctx, &helix.UpdateChannelStreamScheduleSegmentParams{
    BroadcasterID: "12345",
    ID:            "segment-id",
    StartTime:     "2025-12-15T21:00:00Z",
    Duration:      240, // 4 hours
    CategoryID:    "509670", // Different category
    Title:         "Updated stream title",
    Timezone:      "America/New_York",
})

// Cancel a segment
resp, err = client.UpdateChannelStreamScheduleSegment(ctx, &helix.UpdateChannelStreamScheduleSegmentParams{
    BroadcasterID: "12345",
    ID:            "segment-id",
    IsCanceled:    true,
})

segment := resp.Data.Segments[0]
fmt.Printf("Updated segment: %s (Status: %s)\n", segment.Title,
    map[bool]string{true: "Canceled", false: "Active"}[segment.CanceledUntil != nil])
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `ID` - Required. ID of the segment to update
- `StartTime` - Optional. New start time (RFC3339 format)
- `Duration` - Optional. New duration in minutes
- `CategoryID` - Optional. New game/category ID
- `Title` - Optional. New segment title
- `IsCanceled` - Optional. Cancel or un-cancel the segment
- `Timezone` - Optional. Timezone (IANA format)

**Sample Response:**
```json
{
  "data": {
    "segments": [
      {
        "id": "eyJzZWdtZW50SUQiOiI1NjkyNmI0ZC0xNGJmLTRjYjEtOGU3ZS00ZjBjMDNiZjgwZWYiLCJpc29ZZWFyIjoyMDI1LCJpc29XZWVrIjo1MH0=",
        "start_time": "2025-12-15T21:00:00Z",
        "end_time": "2025-12-16T01:00:00Z",
        "title": "Updated stream title",
        "canceled_until": null,
        "category": {
          "id": "509670",
          "name": "Talk Shows & Podcasts"
        },
        "is_recurring": false
      }
    ]
  }
}
```

## DeleteChannelStreamScheduleSegment

Delete a scheduled stream segment.

**Requires:** `channel:manage:schedule`

```go
err := client.DeleteChannelStreamScheduleSegment(ctx, "12345", "segment-id")
if err != nil {
    fmt.Printf("Failed to delete segment: %v\n", err)
}
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `SegmentID` - Required. ID of the segment to delete

**Sample Response:**
This endpoint returns no content on success (HTTP 204 No Content).

