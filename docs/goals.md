# Goals API

Manage creator goals for Twitch channels.

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

```go
resp, err := client.GetHypeTrainStatus(ctx, &helix.GetHypeTrainStatusParams{
    BroadcasterID: "12345",
})
if err != nil {
    log.Fatal(err)
}
for _, hypeTrain := range resp.Data {
    fmt.Printf("Hype Train Level: %d\n", hypeTrain.Level)
    fmt.Printf("Total: %d, Goal: %d\n", hypeTrain.Total, hypeTrain.Goal)
    fmt.Printf("Started: %s, Expires: %s\n", hypeTrain.StartedAt, hypeTrain.ExpiresAt)
    fmt.Printf("Last contribution: %s (%d)\n",
        hypeTrain.LastContribution.User, hypeTrain.LastContribution.Total)

    fmt.Println("Top Contributors:")
    for _, contrib := range hypeTrain.TopContributions {
        fmt.Printf("  %s: %d (%s)\n", contrib.User, contrib.Total, contrib.Type)
    }
}
```

**Parameters:**
- `BroadcasterID` (string): The ID of the broadcaster

**Response Fields:**
- `Level` (int): Current level of the hype train
- `Total` (int): Total points contributed to the hype train
- `Goal` (int): Points needed to reach the next level
- `TopContributions` (array): List of top contributors
- `LastContribution` (object): Information about the most recent contribution
- `StartedAt` (string): Timestamp when the hype train started
- `ExpiresAt` (string): Timestamp when the hype train expires

## EventSub Hype Train Events

For real-time hype train notifications, use EventSub subscriptions.

**Note:** Hype Train v1 is deprecated by Twitch. This library defaults to v2.

**V2 Fields:**
- `Type` - `regular`, `golden_kappa`, or `shared`
- `IsSharedTrain` - Whether this is a shared hype train
- `SharedTrainParticipants` - Participating broadcasters for shared trains
- `AllTimeHighLevel` / `AllTimeHighTotal` - Channel's all-time records

See [EventSub documentation](eventsub.md#hype-train-events) for subscription details, code examples, and migration guidance.

**Sample Response:**
```json
{
  "data": [
    {
      "id": "1b0AsbInCHZW2SQFQkCzqN07Ib2",
      "broadcaster_id": "12345",
      "level": 3,
      "total": 2800,
      "goal": 3500,
      "top_contributions": [
        {
          "total": 800,
          "type": "BITS",
          "user": "user456"
        },
        {
          "total": 650,
          "type": "SUBS",
          "user": "user789"
        },
        {
          "total": 400,
          "type": "BITS",
          "user": "user234"
        }
      ],
      "last_contribution": {
        "total": 150,
        "type": "SUBS",
        "user": "user101"
      },
      "started_at": "2025-03-15T18:13:45Z",
      "expires_at": "2025-03-15T18:33:45Z",
      "cooldown_end_time": "2025-03-15T19:23:45Z"
    }
  ]
}
```
