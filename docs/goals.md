---
layout: default
title: Goals API
description: Manage creator goals for Twitch channels.
---

## GetCreatorGoals

Get information about the broadcaster's creator goals.

**Requires:** `channel:read:goals` scope

```go
resp, err := client.GetCreatorGoals(ctx, "141981764")
if err != nil {
    log.Fatal(err)
}
for _, goal := range resp.Data {
    fmt.Printf("Goal: %s\n", goal.Description)
    fmt.Printf("Type: %s\n", goal.Type)
    fmt.Printf("Progress: %d / %d\n", goal.CurrentAmount, goal.TargetAmount)
    fmt.Printf("Created: %s\n", goal.CreatedAt)
}
```

**Parameters:**
- `broadcasterID` (string): The ID of the broadcaster

**Response Fields:**
- `ID` (string): Goal ID
- `BroadcasterID` (string): Broadcaster's user ID
- `BroadcasterName` (string): Broadcaster's display name
- `BroadcasterLogin` (string): Broadcaster's login name
- `Type` (string): Goal type - `follower`, `subscription`, `subscription_count`, `new_subscription`, or `new_subscription_count`
- `Description` (string): Goal description
- `CurrentAmount` (int): Current progress toward the goal
- `TargetAmount` (int): Target amount to complete the goal
- `CreatedAt` (time.Time): When the goal was created

**Sample Response (from Twitch docs):**
```json
{
  "data": [
    {
      "id": "1woowvbkiNv8BRxEWSqmQz6Zk92",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "type": "follower",
      "description": "Follow goal for Helix testing",
      "current_amount": 27062,
      "target_amount": 30000,
      "created_at": "2021-08-16T17:22:23Z"
    }
  ]
}
```

## Goal Types

| Type | Description |
|------|-------------|
| `follower` | Track followers |
| `subscription` | Track subscription revenue |
| `subscription_count` | Track total subscriber count |
| `new_subscription` | Track new subscription revenue |
| `new_subscription_count` | Track new subscriber count |

## EventSub Integration

You can receive real-time updates when goals are created, updated, or achieved using EventSub:

- `channel.goal.begin` - A goal is created
- `channel.goal.progress` - Progress is made toward a goal
- `channel.goal.end` - A goal ends (achieved or cancelled)
See [EventSub documentation](eventsub.md) for more details.

