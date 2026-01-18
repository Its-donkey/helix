# Hype Train API

Manage hype train events for Twitch channels.

> **Note:** The Get Hype Train Events endpoint is deprecated and will be removed January 15, 2026. Use [EventSub](eventsub.md) for real-time hype train events instead.

## GetHypeTrainEvents

Get the broadcaster's hype train events.

**Requires:** `channel:read:hype_train` scope

**DEPRECATED:** This endpoint is deprecated. Use EventSub instead.

```go
resp, err := client.GetHypeTrainEvents(ctx, &helix.GetHypeTrainEventsParams{
    BroadcasterID: "270954519",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})
if err != nil {
    log.Fatal(err)
}
for _, event := range resp.Data {
    fmt.Printf("Event: %s (%s)\n", event.ID, event.EventType)
    fmt.Printf("Hype Train Level: %d\n", event.EventData.Level)
    fmt.Printf("Progress: %d / %d\n", event.EventData.Total, event.EventData.Goal)
    fmt.Printf("Started: %s, Expires: %s\n", event.EventData.StartedAt, event.EventData.ExpiresAt)
}

// Paginate through more results
if resp.Pagination != nil && resp.Pagination.Cursor != "" {
    resp, err = client.GetHypeTrainEvents(ctx, &helix.GetHypeTrainEventsParams{
        BroadcasterID: "270954519",
        PaginationParams: &helix.PaginationParams{
            After: resp.Pagination.Cursor,
        },
    })
}
```

**Parameters:**
- `BroadcasterID` (string): The ID of the broadcaster
- `PaginationParams.First` (int, optional): Maximum number of results (1-100)
- `PaginationParams.After` (string, optional): Cursor for forward pagination

**Response Fields:**
- `ID` (string): Event ID
- `EventType` (string): Type of event (e.g., `hypetrain.progression`)
- `EventTimestamp` (time.Time): When the event occurred
- `Version` (string): Event version
- `EventData.ID` (string): Hype train ID
- `EventData.BroadcasterID` (string): Broadcaster's user ID
- `EventData.Level` (int): Current level of the hype train
- `EventData.Total` (int): Total points contributed
- `EventData.Goal` (int): Points needed to reach the next level
- `EventData.StartedAt` (time.Time): When the hype train started
- `EventData.ExpiresAt` (time.Time): When the hype train expires
- `EventData.CooldownEndTime` (time.Time): When a new hype train can start
- `EventData.LastContribution` (HypeTrainContribution): Most recent contribution
- `EventData.TopContributions` ([]HypeTrainContribution): Top contributors

**Sample Response (from Twitch docs):**
```json
{
  "data": [
    {
      "id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
      "event_type": "hypetrain.progression",
      "event_timestamp": "2020-04-24T20:07:24Z",
      "version": "1.0",
      "event_data": {
        "broadcaster_id": "270954519",
        "cooldown_end_time": "2020-04-24T20:13:21.003802269Z",
        "expires_at": "2020-04-24T20:12:21.003802269Z",
        "goal": 1800,
        "id": "70f0c7d8-ff60-4c50-b138-f3a352833b50",
        "last_contribution": {
          "total": 200,
          "type": "BITS",
          "user": "134247454"
        },
        "level": 2,
        "started_at": "2020-04-24T20:05:47.30473127Z",
        "top_contributions": [
          {
            "total": 600,
            "type": "BITS",
            "user": "134247450"
          }
        ],
        "total": 600
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjI3MDk1NDUxOToxNTg3NzU4ODQ0OjFiMEFzYkluQ0haVzJTUUZRa0N6cU4wN0liMiJ9fQ"
  }
}
```

## GetHypeTrainStatus

Get the current hype train status for a broadcaster's channel. Returns `nil` if no hype train is active.

**Requires:** `channel:read:hype_train` scope

```go
status, err := client.GetHypeTrainStatus(ctx, "270954519")
if err != nil {
    log.Fatal(err)
}
if status == nil {
    fmt.Println("No active hype train")
    return
}

fmt.Printf("Hype Train Level: %d\n", status.Level)
fmt.Printf("Progress: %d / %d\n", status.Total, status.Goal)
fmt.Printf("Started: %s, Expires: %s\n", status.StartedAt, status.ExpiresAt)
fmt.Printf("Last contribution: %s (%d %s)\n",
    status.LastContribution.User,
    status.LastContribution.Total,
    status.LastContribution.Type)

fmt.Println("Top Contributors:")
for _, contrib := range status.TopContributions {
    fmt.Printf("  %s: %d (%s)\n", contrib.User, contrib.Total, contrib.Type)
}
```

**Parameters:**
- `broadcasterID` (string): The ID of the broadcaster

**Response Fields:**
- `ID` (string): Hype train ID
- `BroadcasterID` (string): Broadcaster's user ID
- `Level` (int): Current level of the hype train
- `Total` (int): Total points contributed to the hype train
- `Goal` (int): Points needed to reach the next level
- `TopContributions` ([]HypeTrainContribution): List of top contributors
- `LastContribution` (HypeTrainContribution): Most recent contribution
- `StartedAt` (time.Time): When the hype train started
- `ExpiresAt` (time.Time): When the hype train expires
- `CooldownEndTime` (time.Time): When a new hype train can start

**HypeTrainContribution Fields:**
- `Total` (int): Contribution amount
- `Type` (string): Contribution type (`BITS`, `SUBS`, `OTHER`)
- `User` (string): User ID of the contributor

**Sample Response (based on Twitch docs):**
```json
{
  "data": [
    {
      "id": "70f0c7d8-ff60-4c50-b138-f3a352833b50",
      "broadcaster_id": "270954519",
      "level": 2,
      "total": 600,
      "goal": 1800,
      "top_contributions": [
        {
          "total": 600,
          "type": "BITS",
          "user": "134247450"
        }
      ],
      "last_contribution": {
        "total": 200,
        "type": "BITS",
        "user": "134247454"
      },
      "started_at": "2020-04-24T20:05:47.30473127Z",
      "expires_at": "2020-04-24T20:12:21.003802269Z",
      "cooldown_end_time": "2020-04-24T20:13:21.003802269Z"
    }
  ]
}
```

## EventSub Alternative

For real-time hype train events, use EventSub subscriptions instead of polling the deprecated API:

- `channel.hype_train.begin` - Hype train starts
- `channel.hype_train.progress` - Hype train level increases
- `channel.hype_train.end` - Hype train ends

See [EventSub documentation](eventsub.md) for more details.
