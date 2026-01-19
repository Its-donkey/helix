---
layout: default
title: Channel Points API
description: Manage custom Channel Points rewards and redemptions.
---

## GetCustomRewards

Get custom Channel Points rewards for a broadcaster.

**Requires:** `channel:read:redemptions` or `channel:manage:redemptions`

```go
// Get all rewards
resp, err := client.GetCustomRewards(ctx, &helix.GetCustomRewardsParams{
    BroadcasterID: "12345",
})

// Get specific rewards by ID (max 50)
resp, err = client.GetCustomRewards(ctx, &helix.GetCustomRewardsParams{
    BroadcasterID: "12345",
    IDs:           []string{"reward-id-1", "reward-id-2"},
})

// Get only manageable rewards
resp, err = client.GetCustomRewards(ctx, &helix.GetCustomRewardsParams{
    BroadcasterID:         "12345",
    OnlyManageableRewards: true,
})

for _, reward := range resp.Data {
    fmt.Printf("Reward: %s - %d points\n", reward.Title, reward.Cost)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "92af127c-7326-4483-a52b-b0da0be61c01",
      "title": "Hydrate",
      "prompt": "Drink some water!",
      "cost": 100,
      "image": {
        "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/12345/92af127c-7326-4483-a52b-b0da0be61c01/1.png",
        "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/12345/92af127c-7326-4483-a52b-b0da0be61c01/2.png",
        "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/12345/92af127c-7326-4483-a52b-b0da0be61c01/4.png"
      },
      "default_image": {
        "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/default-1.png",
        "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/default-2.png",
        "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"
      },
      "background_color": "#00E5CB",
      "is_enabled": true,
      "is_user_input_required": false,
      "max_per_stream_setting": {
        "is_enabled": true,
        "max_per_stream": 10
      },
      "max_per_user_per_stream_setting": {
        "is_enabled": false,
        "max_per_user_per_stream": 0
      },
      "global_cooldown_setting": {
        "is_enabled": true,
        "global_cooldown_seconds": 300
      },
      "is_paused": false,
      "is_in_stock": true,
      "should_redemptions_skip_request_queue": false,
      "redemptions_redeemed_current_stream": 3,
      "cooldown_expires_at": "2023-11-15T20:30:00Z"
    }
  ]
}
```

## CreateCustomReward

Create a new custom Channel Points reward.

**Requires:** `channel:manage:redemptions`

```go
resp, err := client.CreateCustomReward(ctx, &helix.CreateCustomRewardParams{
    BroadcasterID:         "12345",
    Title:                 "Highlight My Message",
    Cost:                  300,
    Prompt:                "Your message will be highlighted in chat!",
    IsEnabled:             true,
    BackgroundColor:       "#9147FF",
    IsUserInputRequired:   true,
    MaxPerStream:          10,
    MaxPerStreamEnabled:   true,
    GlobalCooldownSeconds: 60,
    GlobalCooldownEnabled: true,
    ShouldRedemptionsSkipRequestQueue: false,
})

fmt.Printf("Created reward: %s (ID: %s)\n", resp.Data[0].Title, resp.Data[0].ID)
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "a8b3e5d7-9c4f-4a1b-8e2d-6f3c5a7b9d1e",
      "title": "Highlight My Message",
      "prompt": "Your message will be highlighted in chat!",
      "cost": 300,
      "image": null,
      "default_image": {
        "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/default-1.png",
        "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/default-2.png",
        "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"
      },
      "background_color": "#9147FF",
      "is_enabled": true,
      "is_user_input_required": true,
      "max_per_stream_setting": {
        "is_enabled": true,
        "max_per_stream": 10
      },
      "max_per_user_per_stream_setting": {
        "is_enabled": false,
        "max_per_user_per_stream": 0
      },
      "global_cooldown_setting": {
        "is_enabled": true,
        "global_cooldown_seconds": 60
      },
      "is_paused": false,
      "is_in_stock": true,
      "should_redemptions_skip_request_queue": false,
      "redemptions_redeemed_current_stream": 0
    }
  ]
}
```

## UpdateCustomReward

Update an existing custom Channel Points reward.

**Requires:** `channel:manage:redemptions`

```go
resp, err := client.UpdateCustomReward(ctx, &helix.UpdateCustomRewardParams{
    BroadcasterID: "12345",
    ID:            "reward-id",
    Title:         "Updated Reward Name",
    Cost:          500,
    Prompt:        "Updated description",
    IsEnabled:     true,
    IsPaused:      false,
})

fmt.Printf("Updated reward: %s\n", resp.Data[0].Title)
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "reward-id",
      "title": "Updated Reward Name",
      "prompt": "Updated description",
      "cost": 500,
      "image": null,
      "default_image": {
        "url_1x": "https://static-cdn.jtvnw.net/custom-reward-images/default-1.png",
        "url_2x": "https://static-cdn.jtvnw.net/custom-reward-images/default-2.png",
        "url_4x": "https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"
      },
      "background_color": "#9147FF",
      "is_enabled": true,
      "is_user_input_required": false,
      "max_per_stream_setting": {
        "is_enabled": true,
        "max_per_stream": 10
      },
      "max_per_user_per_stream_setting": {
        "is_enabled": false,
        "max_per_user_per_stream": 0
      },
      "global_cooldown_setting": {
        "is_enabled": true,
        "global_cooldown_seconds": 60
      },
      "is_paused": false,
      "is_in_stock": true,
      "should_redemptions_skip_request_queue": false,
      "redemptions_redeemed_current_stream": 2,
      "cooldown_expires_at": "2023-11-15T20:45:00Z"
    }
  ]
}
```

## DeleteCustomReward

Delete a custom Channel Points reward.

**Requires:** `channel:manage:redemptions`

```go
err := client.DeleteCustomReward(ctx, &helix.DeleteCustomRewardParams{
    BroadcasterID: "12345",
    ID:            "reward-id",
})
```

**Note:** This endpoint returns no data on success (HTTP 204 No Content).

## GetCustomRewardRedemptions

Get redemptions for a custom Channel Points reward.

**Requires:** `channel:read:redemptions` or `channel:manage:redemptions`

```go
// Get all unfulfilled redemptions for a reward
resp, err := client.GetCustomRewardRedemptions(ctx, &helix.GetCustomRewardRedemptionsParams{
    BroadcasterID: "12345",
    RewardID:      "reward-id",
    Status:        "UNFULFILLED",
})

// Get specific redemptions by ID
resp, err = client.GetCustomRewardRedemptions(ctx, &helix.GetCustomRewardRedemptionsParams{
    BroadcasterID: "12345",
    RewardID:      "reward-id",
    IDs:           []string{"redemption-id-1", "redemption-id-2"},
})

// Get fulfilled redemptions sorted by newest first
resp, err = client.GetCustomRewardRedemptions(ctx, &helix.GetCustomRewardRedemptionsParams{
    BroadcasterID: "12345",
    RewardID:      "reward-id",
    Status:        "FULFILLED",
    Sort:          "NEWEST",
    PaginationParams: &helix.PaginationParams{
        First: 50,
    },
})

// Get canceled redemptions sorted by oldest first
resp, err = client.GetCustomRewardRedemptions(ctx, &helix.GetCustomRewardRedemptionsParams{
    BroadcasterID: "12345",
    RewardID:      "reward-id",
    Status:        "CANCELED",
    Sort:          "OLDEST",
})

for _, redemption := range resp.Data {
    fmt.Printf("User %s redeemed: %s (Status: %s)\n",
        redemption.UserName, redemption.Reward.Title, redemption.Status)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "17fa2df1-ad76-4804-bfa5-a40ef63efe63",
      "user_id": "274637212",
      "user_login": "vinh",
      "user_name": "Vinh",
      "user_input": "Play my favorite song please!",
      "status": "UNFULFILLED",
      "redeemed_at": "2023-11-15T19:30:00Z",
      "reward": {
        "id": "92af127c-7326-4483-a52b-b0da0be61c01",
        "title": "Song Request",
        "prompt": "Request your favorite song",
        "cost": 500
      }
    },
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "user_id": "123456789",
      "user_login": "coolviewer",
      "user_name": "CoolViewer",
      "user_input": "",
      "status": "UNFULFILLED",
      "redeemed_at": "2023-11-15T19:25:00Z",
      "reward": {
        "id": "92af127c-7326-4483-a52b-b0da0be61c01",
        "title": "Hydrate",
        "prompt": "Drink some water!",
        "cost": 100
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6IjE5Nzk1MTQ3MTYwMDAwMDAwMDAifX0"
  }
}
```

## UpdateRedemptionStatus

Update the status of custom Channel Points reward redemptions.

**Requires:** `channel:manage:redemptions`

```go
// Fulfill redemptions (max 50 IDs)
resp, err := client.UpdateRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
    BroadcasterID: "12345",
    RewardID:      "reward-id",
    IDs:           []string{"redemption-id-1", "redemption-id-2"},
    Status:        "FULFILLED",
})

// Cancel redemptions
resp, err = client.UpdateRedemptionStatus(ctx, &helix.UpdateRedemptionStatusParams{
    BroadcasterID: "12345",
    RewardID:      "reward-id",
    IDs:           []string{"redemption-id-3"},
    Status:        "CANCELED",
})

for _, redemption := range resp.Data {
    fmt.Printf("Updated redemption %s to %s\n", redemption.ID, redemption.Status)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "17fa2df1-ad76-4804-bfa5-a40ef63efe63",
      "user_id": "274637212",
      "user_login": "vinh",
      "user_name": "Vinh",
      "user_input": "Play my favorite song please!",
      "status": "FULFILLED",
      "redeemed_at": "2023-11-15T19:30:00Z",
      "reward": {
        "id": "92af127c-7326-4483-a52b-b0da0be61c01",
        "title": "Song Request",
        "prompt": "Request your favorite song",
        "cost": 500
      }
    },
    {
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "user_id": "123456789",
      "user_login": "coolviewer",
      "user_name": "CoolViewer",
      "user_input": "",
      "status": "FULFILLED",
      "redeemed_at": "2023-11-15T19:25:00Z",
      "reward": {
        "id": "92af127c-7326-4483-a52b-b0da0be61c01",
        "title": "Hydrate",
        "prompt": "Drink some water!",
        "cost": 100
      }
    }
  ]
}
```

