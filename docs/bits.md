---
layout: default
title: Bits API
description: Retrieve Bits leaderboards and cheermote information for Twitch channels.
---

## GetBitsLeaderboard

Get the Bits leaderboard for the authenticated broadcaster.

**Requires:** `bits:read` scope

```go
// Get top 10 users for the current week
resp, err := client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
    Count:  10,
    Period: "week",
})

// Get top 100 users for all time
resp, err = client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
    Count:  100,
    Period: "all",
})

// Get leaderboard for a specific time period
resp, err = client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
    Count:     20,
    Period:    "month",
    StartedAt: "2023-01-01T00:00:00Z",
})

// Get leaderboard for a specific user
resp, err = client.GetBitsLeaderboard(ctx, &helix.GetBitsLeaderboardParams{
    UserID: "12345",
    Period: "year",
})

for _, entry := range resp.Data {
    fmt.Printf("#%d: %s - %d bits\n", entry.Rank, entry.UserName, entry.Score)
}
```

**Parameters:**
- `Count` (int): Maximum number of results (1-100)
- `Period` (string): Time period for the leaderboard (`day`, `week`, `month`, `year`, or `all`)
- `StartedAt` (string, optional): Timestamp for the period start
- `UserID` (string, optional): Filter to a specific user

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "158010205",
      "user_login": "tundra",
      "user_name": "Tundra",
      "rank": 1,
      "score": 12543
    },
    {
      "user_id": "7168163",
      "user_login": "topramens",
      "user_name": "TopRamens",
      "rank": 2,
      "score": 6900
    },
    {
      "user_id": "160942421",
      "user_login": "jadio",
      "user_name": "Jadio",
      "rank": 3,
      "score": 6240
    }
  ],
  "date_range": {
    "started_at": "2025-12-04T00:00:00Z",
    "ended_at": "2025-12-11T00:00:00Z"
  },
  "total": 3
}
```

## GetCheermotes

Get the list of available Cheermotes for a channel or globally.

**No authentication required**

```go
// Get global cheermotes
resp, err := client.GetCheermotes(ctx, nil)

// Get cheermotes for a specific broadcaster
resp, err = client.GetCheermotes(ctx, &helix.GetCheermotesParams{
    BroadcasterID: "12345",
})

for _, cheermote := range resp.Data {
    fmt.Printf("Cheermote: %s (prefix: %s)\n", cheermote.Type, cheermote.Prefix)
    for _, tier := range cheermote.Tiers {
        fmt.Printf("  Tier %d: min %d bits - %s\n", tier.ID, tier.MinBits, tier.Color)
    }
}
```

**Parameters:**
- `BroadcasterID` (string, optional): Get cheermotes available for this broadcaster (includes custom cheermotes)

**Sample Response:**
```json
{
  "data": [
    {
      "prefix": "Cheer",
      "tiers": [
        {
          "min_bits": 1,
          "id": "1",
          "color": "#979797",
          "images": {
            "dark": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1/4.png"
              }
            },
            "light": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1/4.png"
              }
            }
          },
          "can_cheer": true,
          "show_in_bits_card": true
        },
        {
          "min_bits": 100,
          "id": "100",
          "color": "#9c3ee8",
          "images": {
            "dark": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/100/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/100/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/100/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/100/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/100/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/100/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/100/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/100/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/100/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/100/4.png"
              }
            },
            "light": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/100/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/100/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/100/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/100/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/100/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/100/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/100/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/100/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/100/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/100/4.png"
              }
            }
          },
          "can_cheer": true,
          "show_in_bits_card": true
        },
        {
          "min_bits": 1000,
          "id": "1000",
          "color": "#1db2a5",
          "images": {
            "dark": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1000/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1000/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1000/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1000/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/animated/1000/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1000/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1000/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1000/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1000/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/dark/static/1000/4.png"
              }
            },
            "light": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1000/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1000/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1000/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1000/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/animated/1000/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1000/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1000/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1000/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1000/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/actions/cheer/light/static/1000/4.png"
              }
            }
          },
          "can_cheer": true,
          "show_in_bits_card": true
        }
      ],
      "type": "global_first_party",
      "order": 1,
      "last_updated": "2025-12-01T10:30:00Z",
      "is_charitable": false
    },
    {
      "prefix": "StreamlabsCharity",
      "tiers": [
        {
          "min_bits": 1,
          "id": "1",
          "color": "#1db2a5",
          "images": {
            "dark": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/animated/1/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/animated/1/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/animated/1/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/animated/1/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/animated/1/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/static/1/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/static/1/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/static/1/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/static/1/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/dark/static/1/4.png"
              }
            },
            "light": {
              "animated": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/animated/1/1.gif",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/animated/1/1.5.gif",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/animated/1/2.gif",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/animated/1/3.gif",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/animated/1/4.gif"
              },
              "static": {
                "1": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/static/1/1.png",
                "1.5": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/static/1/1.5.png",
                "2": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/static/1/2.png",
                "3": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/static/1/3.png",
                "4": "https://d3aqoihi2n8ty8.cloudfront.net/partner-actions/streamlabs-charity/light/static/1/4.png"
              }
            }
          },
          "can_cheer": true,
          "show_in_bits_card": true
        }
      ],
      "type": "channel_custom",
      "order": 2,
      "last_updated": "2025-11-15T14:22:00Z",
      "is_charitable": true
    }
  ]
}
```

