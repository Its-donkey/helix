---
layout: default
title: Predictions API
description: Manage Twitch predictions for a broadcaster's channel.
---

## GetPredictions

Get information about predictions for a broadcaster.

**Requires:** `channel:read:predictions`

```go
// Get all active predictions
resp, err := client.GetPredictions(ctx, &helix.GetPredictionsParams{
    BroadcasterID: "12345",
})

// Get specific predictions by ID (max 100)
resp, err = client.GetPredictions(ctx, &helix.GetPredictionsParams{
    BroadcasterID: "12345",
    IDs:           []string{"prediction-id-1", "prediction-id-2"},
})

// Get predictions with pagination
resp, err = client.GetPredictions(ctx, &helix.GetPredictionsParams{
    BroadcasterID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

for _, prediction := range resp.Data {
    fmt.Printf("Prediction: %s - Status: %s\n", prediction.Title, prediction.Status)
    fmt.Printf("Broadcaster: %s (%s)\n", prediction.BroadcasterName, prediction.BroadcasterLogin)
    for _, outcome := range prediction.Outcomes {
        fmt.Printf("  - %s (Color: %s)\n", outcome.Title, outcome.Color)
        fmt.Printf("    Users: %d, Channel Points: %d\n", outcome.Users, outcome.ChannelPoints)
    }
}
```

**Response Fields:**
- `ID` - Prediction ID
- `BroadcasterID`, `BroadcasterName`, `BroadcasterLogin` - Broadcaster information
- `Title` - Prediction title
- `WinningOutcomeID` - ID of the winning outcome (if resolved)
- `Outcomes` - Array of prediction outcomes
  - `ID` - Outcome ID
  - `Title` - Outcome title
  - `Users` - Number of users who predicted this outcome
  - `ChannelPoints` - Total Channel Points wagered on this outcome
  - `TopPredictors` - Top predictors for this outcome
  - `Color` - Outcome color (BLUE or PINK)
- `PredictionWindow` - Prediction window in seconds
- `Status` - Prediction status: `ACTIVE`, `CANCELED`, `LOCKED`, `RESOLVED`
- `CreatedAt` - When the prediction was created
- `EndedAt` - When the prediction ended (if applicable)
- `LockedAt` - When the prediction was locked (if applicable)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "bc637af0-7766-4525-9308-4112f4cbf178",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "title": "Will I beat this boss?",
      "winning_outcome_id": "73085848-a94d-4040-9d21-2cb7a89374b7",
      "outcomes": [
        {
          "id": "73085848-a94d-4040-9d21-2cb7a89374b7",
          "title": "Yes",
          "users": 10,
          "channel_points": 15000,
          "top_predictors": [
            {
              "user_id": "12345678",
              "user_login": "cooluser123",
              "user_name": "CoolUser123",
              "channel_points_used": 5000,
              "channel_points_won": 9500
            }
          ],
          "color": "BLUE"
        },
        {
          "id": "906b70ba-1f12-47ea-9e95-e5b93cac8c4e",
          "title": "No",
          "users": 5,
          "channel_points": 7500,
          "top_predictors": [
            {
              "user_id": "87654321",
              "user_login": "skeptic99",
              "user_name": "Skeptic99",
              "channel_points_used": 3000,
              "channel_points_won": 0
            }
          ],
          "color": "PINK"
        }
      ],
      "prediction_window": 300,
      "status": "RESOLVED",
      "created_at": "2025-12-11T10:15:00Z",
      "ended_at": "2025-12-11T10:20:30Z",
      "locked_at": "2025-12-11T10:20:00Z"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## CreatePrediction

Create a new prediction for a broadcaster's channel.

**Requires:** `channel:manage:predictions`

```go
resp, err := client.CreatePrediction(ctx, &helix.CreatePredictionParams{
    BroadcasterID: "12345",
    Title:         "Will I beat this boss?",
    Outcomes: []helix.PredictionOutcome{
        {Title: "Yes"},
        {Title: "No"},
    },
    PredictionWindow: 300, // 5 minutes (30-1800 seconds)
})

fmt.Printf("Created prediction: %s (ID: %s)\n", resp.Data[0].Title, resp.Data[0].ID)
fmt.Printf("Status: %s, Created at: %s\n", resp.Data[0].Status, resp.Data[0].CreatedAt)

// Example with more outcomes (2-10 allowed)
resp, err = client.CreatePrediction(ctx, &helix.CreatePredictionParams{
    BroadcasterID: "12345",
    Title:         "What will happen next?",
    Outcomes: []helix.PredictionOutcome{
        {Title: "Win"},
        {Title: "Lose"},
        {Title: "Draw"},
        {Title: "Rage Quit"},
    },
    PredictionWindow: 600, // 10 minutes
})
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster creating the prediction
- `Title` - Required. The prediction title
- `Outcomes` - Required. Array of outcomes (2-10 outcomes)
  - `Title` - Required. The outcome title
- `PredictionWindow` - Required. Prediction window in seconds (30-1800)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "bc637af0-7766-4525-9308-4112f4cbf178",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "title": "Will I beat this boss?",
      "winning_outcome_id": null,
      "outcomes": [
        {
          "id": "73085848-a94d-4040-9d21-2cb7a89374b7",
          "title": "Yes",
          "users": 0,
          "channel_points": 0,
          "top_predictors": null,
          "color": "BLUE"
        },
        {
          "id": "906b70ba-1f12-47ea-9e95-e5b93cac8c4e",
          "title": "No",
          "users": 0,
          "channel_points": 0,
          "top_predictors": null,
          "color": "PINK"
        }
      ],
      "prediction_window": 300,
      "status": "ACTIVE",
      "created_at": "2025-12-11T10:15:00Z",
      "ended_at": null,
      "locked_at": null
    }
  ]
}
```

## EndPrediction

End an active prediction by resolving, canceling, or locking it.

**Requires:** `channel:manage:predictions`

```go
// Resolve a prediction with a winning outcome
resp, err := client.EndPrediction(ctx, &helix.EndPredictionParams{
    BroadcasterID:    "12345",
    ID:               "prediction-id",
    Status:           "RESOLVED",
    WinningOutcomeID: "outcome-id", // Required when Status is RESOLVED
})

// Cancel a prediction (refund all Channel Points)
resp, err = client.EndPrediction(ctx, &helix.EndPredictionParams{
    BroadcasterID: "12345",
    ID:            "prediction-id",
    Status:        "CANCELED",
})

// Lock a prediction (prevent new predictions)
resp, err = client.EndPrediction(ctx, &helix.EndPredictionParams{
    BroadcasterID: "12345",
    ID:            "prediction-id",
    Status:        "LOCKED",
})

fmt.Printf("Ended prediction: %s with status %s\n", resp.Data[0].Title, resp.Data[0].Status)
```

**Parameters:**
- `BroadcasterID` - Required. ID of the broadcaster
- `ID` - Required. ID of the prediction to end
- `Status` - Required. The status to set: `RESOLVED`, `CANCELED`, or `LOCKED`
- `WinningOutcomeID` - Required if Status is `RESOLVED`. ID of the winning outcome

**Sample Response:**
```json
{
  "data": [
    {
      "id": "bc637af0-7766-4525-9308-4112f4cbf178",
      "broadcaster_id": "141981764",
      "broadcaster_name": "TwitchDev",
      "broadcaster_login": "twitchdev",
      "title": "Will I beat this boss?",
      "winning_outcome_id": "73085848-a94d-4040-9d21-2cb7a89374b7",
      "outcomes": [
        {
          "id": "73085848-a94d-4040-9d21-2cb7a89374b7",
          "title": "Yes",
          "users": 10,
          "channel_points": 15000,
          "top_predictors": [
            {
              "user_id": "12345678",
              "user_login": "cooluser123",
              "user_name": "CoolUser123",
              "channel_points_used": 5000,
              "channel_points_won": 9500
            }
          ],
          "color": "BLUE"
        },
        {
          "id": "906b70ba-1f12-47ea-9e95-e5b93cac8c4e",
          "title": "No",
          "users": 5,
          "channel_points": 7500,
          "top_predictors": [
            {
              "user_id": "87654321",
              "user_login": "skeptic99",
              "user_name": "Skeptic99",
              "channel_points_used": 3000,
              "channel_points_won": 0
            }
          ],
          "color": "PINK"
        }
      ],
      "prediction_window": 300,
      "status": "RESOLVED",
      "created_at": "2025-12-11T10:15:00Z",
      "ended_at": "2025-12-11T10:20:30Z",
      "locked_at": "2025-12-11T10:20:00Z"
    }
  ]
}
```

