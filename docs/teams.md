---
layout: default
title: Teams API
description: Retrieve information about Twitch teams and their members.
---

## GetChannelTeams

Get information about teams that a broadcaster is a member of.

**Requires:** No authentication required

```go
resp, err := client.GetChannelTeams(ctx, &helix.GetChannelTeamsParams{
    BroadcasterID: "12345",
})
for _, team := range resp.Data {
    fmt.Printf("Team: %s (%s)\n", team.TeamDisplayName, team.TeamName)
    fmt.Printf("  Broadcaster: %s (%s)\n", team.BroadcasterName, team.BroadcasterLogin)
    fmt.Printf("  Created: %s, Updated: %s\n", team.CreatedAt, team.UpdatedAt)
    fmt.Printf("  Info: %s\n", team.Info)
    fmt.Printf("  Thumbnail: %s\n", team.ThumbnailURL)
    fmt.Printf("  Background: %s\n", team.BackgroundImageURL)
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
      "background_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/team-devs-background.png",
      "banner": "https://static-cdn.jtvnw.net/jtv_user_pictures/team-devs-banner.png",
      "created_at": "2019-02-11T12:09:22Z",
      "updated_at": "2023-08-15T18:22:10Z",
      "info": "The official Twitch Developer Team - building tools and integrations for the Twitch platform.",
      "thumbnail_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/team-devs-thumbnail.png",
      "team_name": "twitchdev",
      "team_display_name": "Twitch Developers",
      "id": "9876"
    }
  ]
}
```

## GetTeams

Get information about a specific Twitch team by name or ID.

**Requires:** No authentication required

```go
// Get team by name
resp, err := client.GetTeams(ctx, &helix.GetTeamsParams{
    Name: "teamname",
})

// Or get team by ID
resp, err = client.GetTeams(ctx, &helix.GetTeamsParams{
    ID: "12345",
})

for _, team := range resp.Data {
    fmt.Printf("Team: %s (%s)\n", team.TeamDisplayName, team.TeamName)
    fmt.Printf("  ID: %s\n", team.ID)
    fmt.Printf("  Created: %s, Updated: %s\n", team.CreatedAt, team.UpdatedAt)
    fmt.Printf("  Info: %s\n", team.Info)
    fmt.Printf("  Thumbnail: %s\n", team.ThumbnailURL)
    fmt.Printf("  Background: %s\n", team.BackgroundImageURL)
    fmt.Printf("  Banner: %s\n", team.Banner)
    fmt.Printf("  Users (%d):\n", len(team.Users))
    for _, user := range team.Users {
        fmt.Printf("    - %s (%s) [ID: %s]\n", user.Name, user.Login, user.UserID)
    }
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "9876",
      "team_name": "twitchdev",
      "team_display_name": "Twitch Developers",
      "info": "The official Twitch Developer Team - building tools and integrations for the Twitch platform.",
      "thumbnail_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/team-devs-thumbnail.png",
      "background_image_url": "https://static-cdn.jtvnw.net/jtv_user_pictures/team-devs-background.png",
      "banner": "https://static-cdn.jtvnw.net/jtv_user_pictures/team-devs-banner.png",
      "created_at": "2019-02-11T12:09:22Z",
      "updated_at": "2023-08-15T18:22:10Z",
      "users": [
        {
          "user_id": "141981764",
          "user_login": "twitchdev",
          "user_name": "TwitchDev"
        },
        {
          "user_id": "287495632",
          "user_login": "twitchapi",
          "user_name": "TwitchAPI"
        },
        {
          "user_id": "183942137",
          "user_login": "twitchsupport",
          "user_name": "TwitchSupport"
        }
      ]
    }
  ]
}
```

