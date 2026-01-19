---
layout: default
title: Channels API
description: Retrieve and manage Twitch channel information, followers, editors, and VIPs.
---

## GetChannelInformation

Get information about one or more channels.

**Requires:** No authentication required

```go
resp, err := client.GetChannelInformation(ctx, &helix.GetChannelInformationParams{
    BroadcasterIDs: []string{"12345", "67890"},
})
for _, channel := range resp.Data {
    fmt.Printf("Channel: %s, Game: %s, Title: %s\n",
        channel.BroadcasterName, channel.GameName, channel.Title)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "141981764",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "broadcaster_language": "en",
      "game_id": "509658",
      "game_name": "Just Chatting",
      "title": "TwitchDev Monthly Update // May 6, 2021",
      "delay": 0,
      "tags": ["English", "Educational"],
      "content_classification_labels": ["MatureGame"],
      "is_branded_content": false
    }
  ]
}
```

## ModifyChannelInformation

Update channel information for a broadcaster.

**Requires:** `channel:manage:broadcast`

```go
err := client.ModifyChannelInformation(ctx, &helix.ModifyChannelInformationParams{
    BroadcasterID:       "12345",
    GameID:              "509658",
    BroadcasterLanguage: "en",
    Title:               "Epic gameplay stream!",
    Delay:               0,
    Tags:                []string{"English", "Gaming"},
    ContentClassificationLabels: []string{"MatureGame"},
    IsBrandedContent:    false,
})
```

**Sample Response:**
```json
{}
```

Note: This endpoint returns no content on success (204 No Content).

## GetChannelEditors

Get a list of users who have editor permissions for a channel.

**Requires:** `channel:read:editors`

```go
resp, err := client.GetChannelEditors(ctx, "12345")
for _, editor := range resp.Data {
    fmt.Printf("Editor: %s (created at %s)\n", editor.UserName, editor.CreatedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "182891647",
      "user_name": "mauerbac",
      "created_at": "2019-02-15T21:19:50.380833Z"
    },
    {
      "user_id": "135093069",
      "user_name": "BlueLava",
      "created_at": "2018-03-07T16:28:29.872937Z"
    }
  ]
}
```

## GetFollowedChannels

Get a list of channels that a user follows.

**Requires:** `user:read:follows`

```go
// Get all channels a user follows
resp, err := client.GetFollowedChannels(ctx, &helix.GetFollowedChannelsParams{
    UserID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 20,
    },
})

// Check if a user follows a specific broadcaster
resp, err = client.GetFollowedChannels(ctx, &helix.GetFollowedChannelsParams{
    UserID:        "12345",
    BroadcasterID: "67890",
})

for _, follow := range resp.Data {
    fmt.Printf("%s follows %s since %s\n",
        follow.UserName, follow.BroadcasterName, follow.FollowedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "broadcaster_id": "141981764",
      "broadcaster_login": "twitchdev",
      "broadcaster_name": "TwitchDev",
      "followed_at": "2022-05-24T22:22:08Z"
    },
    {
      "broadcaster_id": "41245072",
      "broadcaster_login": "amazonappstv",
      "broadcaster_name": "AmazonAppsTV",
      "followed_at": "2022-05-13T14:45:17Z"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  },
  "total": 274
}
```

## GetChannelFollowers

Get a list of users that follow a broadcaster.

**Requires:** `moderator:read:followers`

```go
// Get all followers for a channel
resp, err := client.GetChannelFollowers(ctx, &helix.GetChannelFollowersParams{
    BroadcasterID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

// Check if a specific user follows the channel
resp, err = client.GetChannelFollowers(ctx, &helix.GetChannelFollowersParams{
    BroadcasterID: "12345",
    UserID:        "67890",
})

for _, follower := range resp.Data {
    fmt.Printf("%s followed on %s\n", follower.UserName, follower.FollowedAt)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "11111",
      "user_login": "userloginname",
      "user_name": "UserDisplayName",
      "followed_at": "2022-05-24T22:22:08Z"
    },
    {
      "user_id": "22222",
      "user_login": "anotheruser",
      "user_name": "AnotherUser",
      "followed_at": "2022-05-20T18:30:45Z"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  },
  "total": 12345
}
```

## GetVIPs

Get a list of VIPs for a channel.

**Requires:** `channel:read:vips`

```go
// Get all VIPs for a channel
resp, err := client.GetVIPs(ctx, &helix.GetVIPsParams{
    BroadcasterID: "12345",
    PaginationParams: &helix.PaginationParams{
        First: 100,
    },
})

// Check specific users for VIP status
resp, err = client.GetVIPs(ctx, &helix.GetVIPsParams{
    BroadcasterID: "12345",
    UserIDs:       []string{"67890", "11111"},
})

for _, vip := range resp.Data {
    fmt.Printf("VIP: %s (%s)\n", vip.UserName, vip.UserLogin)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "user_id": "11111",
      "user_login": "userloginname",
      "user_name": "UserDisplayName"
    },
    {
      "user_id": "22222",
      "user_login": "anotheruser",
      "user_name": "AnotherUser"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## AddChannelVIP

Add a VIP to a channel.

**Requires:** `channel:manage:vips`

```go
err := client.AddChannelVIP(ctx, "12345", "67890")
if err != nil {
    fmt.Printf("Failed to add VIP: %v\n", err)
}
```

**Sample Response:**
```json
{}
```

Note: This endpoint returns no content on success (204 No Content).

## RemoveChannelVIP

Remove a VIP from a channel.

**Requires:** `channel:manage:vips`

```go
err := client.RemoveChannelVIP(ctx, "12345", "67890")
if err != nil {
    fmt.Printf("Failed to remove VIP: %v\n", err)
}
```

**Sample Response:**
```json
{}
```
Note: This endpoint returns no content on success (204 No Content).

