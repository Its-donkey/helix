package helix

import (
	"context"
	"net/url"
	"time"
)

// DropsEntitlement represents a drops entitlement.
type DropsEntitlement struct {
	ID                string    `json:"id"`
	BenefitID         string    `json:"benefit_id"`
	Timestamp         time.Time `json:"timestamp"`
	UserID            string    `json:"user_id"`
	GameID            string    `json:"game_id"`
	FulfillmentStatus string    `json:"fulfillment_status"` // CLAIMED, FULFILLED
	LastUpdated       time.Time `json:"last_updated"`
}

// GetDropsEntitlementsParams contains parameters for GetDropsEntitlements.
type GetDropsEntitlementsParams struct {
	ID                string // Entitlement ID
	UserID            string // Filter by user
	GameID            string // Filter by game
	FulfillmentStatus string // CLAIMED, FULFILLED
	*PaginationParams
}

// GetDropsEntitlements gets drops entitlements.
// Requires: App access token or user token with viewing:activity:read scope.
func (c *Client) GetDropsEntitlements(ctx context.Context, params *GetDropsEntitlementsParams) (*Response[DropsEntitlement], error) {
	q := url.Values{}
	if params != nil {
		if params.ID != "" {
			q.Set("id", params.ID)
		}
		if params.UserID != "" {
			q.Set("user_id", params.UserID)
		}
		if params.GameID != "" {
			q.Set("game_id", params.GameID)
		}
		if params.FulfillmentStatus != "" {
			q.Set("fulfillment_status", params.FulfillmentStatus)
		}
		addPaginationParams(q, params.PaginationParams)
	}

	var resp Response[DropsEntitlement]
	if err := c.get(ctx, "/entitlements/drops", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateDropsEntitlementsParams contains parameters for UpdateDropsEntitlements.
type UpdateDropsEntitlementsParams struct {
	EntitlementIDs    []string `json:"entitlement_ids"`    // Max 100
	FulfillmentStatus string   `json:"fulfillment_status"` // CLAIMED, FULFILLED
}

// UpdateDropsEntitlementsResponse represents the response from UpdateDropsEntitlements.
type UpdateDropsEntitlementsResponse struct {
	Status string   `json:"status"` // SUCCESS, INVALID_ID, NOT_FOUND, UNAUTHORIZED, UPDATE_FAILED
	IDs    []string `json:"ids"`
}

// UpdateDropsEntitlements updates the fulfillment status of drops entitlements.
// Requires: App access token or user token.
func (c *Client) UpdateDropsEntitlements(ctx context.Context, params *UpdateDropsEntitlementsParams) ([]UpdateDropsEntitlementsResponse, error) {
	var resp Response[UpdateDropsEntitlementsResponse]
	if err := c.patch(ctx, "/entitlements/drops", nil, params, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
