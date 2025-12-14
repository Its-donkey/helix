package helix

import (
	"context"
	"net/url"
)

// CharityCampaign represents a charity campaign.
type CharityCampaign struct {
	ID                 string        `json:"id"`
	BroadcasterID      string        `json:"broadcaster_id"`
	BroadcasterLogin   string        `json:"broadcaster_login"`
	BroadcasterName    string        `json:"broadcaster_name"`
	CharityName        string        `json:"charity_name"`
	CharityDescription string        `json:"charity_description"`
	CharityLogo        string        `json:"charity_logo"`
	CharityWebsite     string        `json:"charity_website"`
	CurrentAmount      CharityAmount `json:"current_amount"`
	TargetAmount       CharityAmount `json:"target_amount"`
}

// CharityAmount represents a monetary amount for charity.
type CharityAmount struct {
	Value         int    `json:"value"`
	DecimalPlaces int    `json:"decimal_places"`
	Currency      string `json:"currency"`
}

// GetCharityCampaign gets the active charity campaign for a channel.
// Requires: channel:read:charity scope.
func (c *Client) GetCharityCampaign(ctx context.Context, broadcasterID string) (*CharityCampaign, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[CharityCampaign]
	if err := c.get(ctx, "/charity/campaigns", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// CharityDonation represents a donation to a charity campaign.
type CharityDonation struct {
	ID         string        `json:"id"`
	CampaignID string        `json:"campaign_id"`
	UserID     string        `json:"user_id"`
	UserLogin  string        `json:"user_login"`
	UserName   string        `json:"user_name"`
	Amount     CharityAmount `json:"amount"`
}

// GetCharityDonationsParams contains parameters for GetCharityDonations.
type GetCharityDonationsParams struct {
	BroadcasterID string
	*PaginationParams
}

// GetCharityDonations gets the donations for a charity campaign.
// Requires: channel:read:charity scope.
func (c *Client) GetCharityDonations(ctx context.Context, params *GetCharityDonationsParams) (*Response[CharityDonation], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[CharityDonation]
	if err := c.get(ctx, "/charity/donations", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
