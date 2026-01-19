---
layout: default
title: Charity API
description: Manage charity campaigns and donations for Twitch channels.
---

## GetCharityCampaign

Get information about the charity campaign that a broadcaster is running.

**Requires:** channel:read:charity scope

```go
resp, err := client.GetCharityCampaign(ctx, &helix.GetCharityCampaignParams{
    BroadcasterID: "12345",
})
if err != nil {
    log.Fatal(err)
}
for _, campaign := range resp.Data {
    fmt.Printf("Campaign: %s\n", campaign.CharityName)
    fmt.Printf("Description: %s\n", campaign.CharityDescription)
    fmt.Printf("Website: %s\n", campaign.CharityWebsite)
    fmt.Printf("Progress: %s / %s\n",
        campaign.CurrentAmount.Value, campaign.TargetAmount.Value)
}
```

**Sample Response:**
```json
{
  "data": [
    {
      "id": "123-abc-456-def",
      "broadcaster_id": "12345",
      "broadcaster_login": "twitchstreamer",
      "broadcaster_name": "TwitchStreamer",
      "charity_name": "Example Charity Foundation",
      "charity_description": "A charity dedicated to helping those in need through community support and outreach programs.",
      "charity_logo": "https://abc.cloudfront.net/ppgf/1000/100.png",
      "charity_website": "https://www.examplecharity.org",
      "current_amount": {
        "value": 86000,
        "decimal_places": 2,
        "currency": "USD"
      },
      "target_amount": {
        "value": 150000,
        "decimal_places": 2,
        "currency": "USD"
      }
    }
  ]
}
```

## GetCharityDonations

Get the list of donations that users have made to the broadcaster's charity campaign.

**Requires:** channel:read:charity scope

```go
resp, err := client.GetCharityDonations(ctx, &helix.GetCharityDonationsParams{
    BroadcasterID: "12345",
    First:         20,
})
if err != nil {
    log.Fatal(err)
}
for _, donation := range resp.Data {
    fmt.Printf("%s donated %s %s (Campaign: %s)\n",
        donation.UserName, donation.Amount.Value,
        donation.Amount.DecimalPlaces, donation.CampaignID)
}

// Paginate through more results
if resp.Pagination.Cursor != "" {
    resp, err = client.GetCharityDonations(ctx, &helix.GetCharityDonationsParams{
        BroadcasterID: "12345",
        After:         resp.Pagination.Cursor,
    })
}
```

**Parameters:**
- `BroadcasterID` (string, required): The ID of the broadcaster who is running the charity campaign
- Pagination parameters (`First`, `After`)

**Sample Response:**
```json
{
  "data": [
    {
      "id": "donation-123456",
      "campaign_id": "123-abc-456-def",
      "user_id": "98765",
      "user_login": "generousviewer",
      "user_name": "GenerousViewer",
      "amount": {
        "value": 5000,
        "decimal_places": 2,
        "currency": "USD"
      }
    },
    {
      "id": "donation-789012",
      "campaign_id": "123-abc-456-def",
      "user_id": "54321",
      "user_login": "kindheart",
      "user_name": "KindHeart",
      "amount": {
        "value": 10000,
        "decimal_places": 2,
        "currency": "USD"
      }
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjoiMTU2ODc0NTE3NjQyNDQyMjAwMCJ9"
  }
}
```

