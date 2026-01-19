---
layout: default
title: Polls API
description: Manage Twitch polls for a broadcaster's channel.
---

## GetPolls

Get information about polls for a broadcaster.

**Requires:** `channel:read:polls`

```go
// Get all active polls
resp, err := client.GetPolls(ctx, &helix.GetPollsParams{
    BroadcasterID: "12345",
})

// Get specific polls by ID (max 100)
resp, err = client.GetPolls(ctx, &helix.GetPollsParams{
    BroadcasterID: "12345",
    IDs:           []string{"poll-id-1", "poll-id-2"},
})

// Get polls with pagination
resp, err = client.GetPolls(ctx, &helix.GetPollsParams{
    BroadcasterID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

for _, poll := range resp.Data {
    fmt.Printf("Poll: %s - Status: %s\n", poll.Title, poll.Status)
    fmt.Printf("Broadcaster: %s (%s)\n", poll.BroadcasterName, poll.BroadcasterLogin)
    for _, choice := range poll.Choices {
        fmt.Printf("  - %s\n", choice.Title)
    }
}
```

**Response Fields:**
- `ID` - Poll ID
- `BroadcasterID`, `BroadcasterName`, `BroadcasterLogin` - Broadcaster information
- `Title` - Poll title
- `Choices` - Array of poll choices
- `BitsVotingEnabled` - Whether Bits voting is enabled
- `ChannelPointsVotingEnabled` - Whether Channel Points voting is enabled
- `Status` - Poll status: `ACTIVE`, `COMPLETED`, `TERMINATED`, `ARCHIVED`, `MODERATED`, `INVALID`
- `Duration` - Poll duration in seconds
- `StartedAt` - When the poll started
- `EndedAt` - When the poll ended (if applicable)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "ed961efd-8a3f-4cf5-a9d0-e616c590cd2a",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "title": "What game should I play next?",
      "choices": [
        {
          "id": "4c123012-1351-4f33-84b7-43856e7a0f47",
          "title": "Minecraft",
          "votes": 127,
          "channel_points_votes": 85,
          "bits_votes": 42
        },
        {
          "id": "d1ab80e5-1011-4882-8f4e-ce6d9e0d6f0f",
          "title": "Fortnite",
          "votes": 98,
          "channel_points_votes": 63,
          "bits_votes": 35
        },
        {
          "id": "a2e9e4bc-4b7e-4d0c-8c5f-3e2d1f0c6a8b",
          "title": "Among Us",
          "votes": 156,
          "channel_points_votes": 120,
          "bits_votes": 36
        },
        {
          "id": "f8d5c3e2-9b1a-4f2e-8d7c-6a5b4c3d2e1f",
          "title": "Valorant",
          "votes": 83,
          "channel_points_votes": 55,
          "bits_votes": 28
        }
      ],
      "bits_voting_enabled": true,
      "bits_per_vote": 10,
      "channel_points_voting_enabled": true,
      "channel_points_per_vote": 100,
      "status": "ACTIVE",
      "duration": 300,
      "started_at": "2024-01-15T18:16:00Z"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## CreatePoll

Create a new poll for a broadcaster's channel.

**Requires:** `channel:manage:polls`

```go
resp, err := client.CreatePoll(ctx, &helix.CreatePollParams{
    BroadcasterID: "12345",
    Title:         "What game should I play next?",
    Choices: []string{
        "Minecraft",
        "Fortnite",
        "Among Us",
        "Valorant",
    },
    Duration: 300, // 5 minutes (15-1800 seconds)
    ChannelPointsVotingEnabled: true,
    ChannelPointsPerVote:       100,
})

fmt.Printf("Created poll: %s (ID: %s)\n", resp.Data[0].Title, resp.Data[0].ID)
fmt.Printf("Status: %s, Ends at: %s\n", resp.Data[0].Status, resp.Data[0].StartedAt.Add(time.Duration(resp.Data[0].Duration) * time.Second))
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster creating the poll
- `Title` - Required. The poll title
- `Choices` - Required. Array of choice titles (strings)
- `Duration` - Required. Poll duration in seconds (15-1800)
- `ChannelPointsVotingEnabled` - Optional. Enable Channel Points voting
- `ChannelPointsPerVote` - Optional. Number of Channel Points per vote

**Sample Response:**
```json
{
  "data": [
    {
      "id": "ed961efd-8a3f-4cf5-a9d0-e616c590cd2a",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "title": "What game should I play next?",
      "choices": [
        {
          "id": "4c123012-1351-4f33-84b7-43856e7a0f47",
          "title": "Minecraft",
          "votes": 0,
          "channel_points_votes": 0,
          "bits_votes": 0
        },
        {
          "id": "d1ab80e5-1011-4882-8f4e-ce6d9e0d6f0f",
          "title": "Fortnite",
          "votes": 0,
          "channel_points_votes": 0,
          "bits_votes": 0
        },
        {
          "id": "a2e9e4bc-4b7e-4d0c-8c5f-3e2d1f0c6a8b",
          "title": "Among Us",
          "votes": 0,
          "channel_points_votes": 0,
          "bits_votes": 0
        },
        {
          "id": "f8d5c3e2-9b1a-4f2e-8d7c-6a5b4c3d2e1f",
          "title": "Valorant",
          "votes": 0,
          "channel_points_votes": 0,
          "bits_votes": 0
        }
      ],
      "bits_voting_enabled": false,
      "bits_per_vote": 0,
      "channel_points_voting_enabled": true,
      "channel_points_per_vote": 100,
      "status": "ACTIVE",
      "duration": 300,
      "started_at": "2024-01-15T18:16:00Z"
    }
  ]
}
```

## EndPoll

End an active poll before its scheduled end time.

**Requires:** `channel:manage:polls`

```go
// Terminate a poll (display results)
resp, err := client.EndPoll(ctx, &helix.EndPollParams{
    BroadcasterID: "12345",
    ID:            "poll-id",
    Status:        "TERMINATED",
})

// Archive a poll (hide results)
resp, err = client.EndPoll(ctx, &helix.EndPollParams{
    BroadcasterID: "12345",
    ID:            "poll-id",
    Status:        "ARCHIVED",
})

fmt.Printf("Ended poll: %s with status %s\n", resp.Data[0].Title, resp.Data[0].Status)
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `ID` - Required. ID of the poll to end
- `Status` - Required. The status to set: `TERMINATED` (show results) or `ARCHIVED` (hide results)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "ed961efd-8a3f-4cf5-a9d0-e616c590cd2a",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "title": "What game should I play next?",
      "choices": [
        {
          "id": "4c123012-1351-4f33-84b7-43856e7a0f47",
          "title": "Minecraft",
          "votes": 127,
          "channel_points_votes": 85,
          "bits_votes": 42
        },
        {
          "id": "d1ab80e5-1011-4882-8f4e-ce6d9e0d6f0f",
          "title": "Fortnite",
          "votes": 98,
          "channel_points_votes": 63,
          "bits_votes": 35
        },
        {
          "id": "a2e9e4bc-4b7e-4d0c-8c5f-3e2d1f0c6a8b",
          "title": "Among Us",
          "votes": 156,
          "channel_points_votes": 120,
          "bits_votes": 36
        },
        {
          "id": "f8d5c3e2-9b1a-4f2e-8d7c-6a5b4c3d2e1f",
          "title": "Valorant",
          "votes": 83,
          "channel_points_votes": 55,
          "bits_votes": 28
        }
      ],
      "bits_voting_enabled": true,
      "bits_per_vote": 10,
      "channel_points_voting_enabled": true,
      "channel_points_per_vote": 100,
      "status": "TERMINATED",
      "duration": 300,
      "started_at": "2024-01-15T18:16:00Z",
      "ended_at": "2024-01-15T18:18:45Z"
    }
  ]
}
```

