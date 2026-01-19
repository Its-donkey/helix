---
layout: default
title: Drops Entitlements API
description: Manage Twitch Drops entitlements for games and viewers.
---

## GetDropsEntitlements

Get a list of entitlements for a given organization that have been granted to a game, user, or both.

**Requires:** App access token or user token with `user:read:entitlements` scope

```go
// Get all entitlements for a specific user
resp, err := client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
    UserID: "12345",
})

// Get entitlements for a specific game
resp, err = client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
    GameID: "67890",
})

// Get specific entitlement by ID
resp, err = client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
    ID: "entitlement-id-123",
})

// Filter by fulfillment status
resp, err = client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
    UserID:            "12345",
    FulfillmentStatus: "CLAIMED", // or "FULFILLED"
})

// With pagination
resp, err = client.GetDropsEntitlements(ctx, &helix.GetDropsEntitlementsParams{
    UserID: "12345",
    First:  20,
    After:  cursor,
})

for _, entitlement := range resp.Data {
    fmt.Printf("Entitlement %s: %s - %s\n",
        entitlement.ID, entitlement.BenefitID, entitlement.FulfillmentStatus)
}
```

**Parameters:**
- `ID` (string, optional): Entitlement ID to filter by
- `UserID` (string, optional): User ID to filter by
- `GameID` (string, optional): Game ID to filter by
- `FulfillmentStatus` (string, optional): Filter by status (`CLAIMED` or `FULFILLED`)
- `First` (int, optional): Maximum number of results per page (1-100)
- `After` (string, optional): Cursor for forward pagination

**Sample Response:**
```json
{
  "data": [
    {
      "id": "fb78259e-fb81-4d1b-8333-34a06ffc24c0",
      "benefit_id": "74c52265-e214-48a6-91b9-23b6014e8041",
      "timestamp": "2019-01-28T04:17:53.325Z",
      "user_id": "25612345",
      "game_id": "512333",
      "fulfillment_status": "CLAIMED",
      "last_updated": "2020-01-28T04:25:53.325Z"
    },
    {
      "id": "862c4cba-047a-4ff1-9189-9c94a8100857",
      "benefit_id": "74c52265-e214-48a6-91b9-23b6014e8041",
      "timestamp": "2019-01-28T04:19:53.325Z",
      "user_id": "25612345",
      "game_id": "512333",
      "fulfillment_status": "FULFILLED",
      "last_updated": "2020-01-30T14:35:23.812Z"
    }
  ],
  "pagination": {
    "cursor": "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"
  }
}
```

## UpdateDropsEntitlements

Update the fulfillment status of a list of entitlements. Used to mark entitlements as fulfilled after the user has received the in-game item.

**Requires:** App access token or user token with `user:manage:entitlements` scope

```go
// Mark entitlements as fulfilled
resp, err := client.UpdateDropsEntitlements(ctx, &helix.UpdateDropsEntitlementsParams{
    EntitlementIDs: []string{
        "entitlement-id-1",
        "entitlement-id-2",
        "entitlement-id-3",
    },
    FulfillmentStatus: "FULFILLED",
})
if err != nil {
    log.Fatal(err)
}

// Process results
for _, status := range resp.Data {
    switch status.Status {
    case "SUCCESS":
        fmt.Printf("Successfully updated entitlement %s\n", status.ID)
    case "INVALID_ID":
        fmt.Printf("Invalid entitlement ID: %s\n", status.ID)
    case "NOT_FOUND":
        fmt.Printf("Entitlement not found: %s\n", status.ID)
    case "UNAUTHORIZED":
        fmt.Printf("Not authorized to update entitlement: %s\n", status.ID)
    case "UPDATE_FAILED":
        fmt.Printf("Failed to update entitlement: %s\n", status.ID)
    }
}
```

**Parameters:**
- `EntitlementIDs` ([]string, required): List of entitlement IDs to update (maximum 100)
- `FulfillmentStatus` (string, required): New fulfillment status (`CLAIMED` or `FULFILLED`)

**Sample Response:**
```json
{
  "data": [
    {
      "status": "SUCCESS",
      "ids": [
        "fb78259e-fb81-4d1b-8333-34a06ffc24c0",
        "862c4cba-047a-4ff1-9189-9c94a8100857"
      ]
    },
    {
      "status": "INVALID_ID",
      "ids": [
        "invalid-entitlement-id"
      ]
    },
    {
      "status": "NOT_FOUND",
      "ids": [
        "00000000-0000-0000-0000-000000000000"
      ]
    }
  ]
}
```

**Response Status Values:**
- `SUCCESS`: Entitlement was successfully updated
- `INVALID_ID`: The entitlement ID is not valid
- `NOT_FOUND`: The entitlement was not found
- `UNAUTHORIZED`: The app or user is not authorized to update this entitlement
- `UPDATE_FAILED`: The update failed for an unknown reason

