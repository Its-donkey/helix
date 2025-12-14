package helix

import (
	"context"
	"net/url"
	"time"
)

// Commercial represents a started commercial.
type Commercial struct {
	Length     int    `json:"length"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after"`
}

// StartCommercialParams contains parameters for StartCommercial.
type StartCommercialParams struct {
	BroadcasterID string `json:"broadcaster_id"`
	Length        int    `json:"length"` // 30, 60, 90, 120, 150, or 180 seconds
}

// StartCommercial starts a commercial on a channel.
// Requires: channel:edit:commercial scope.
func (c *Client) StartCommercial(ctx context.Context, params *StartCommercialParams) (*Commercial, error) {
	var resp Response[Commercial]
	if err := c.post(ctx, "/channels/commercial", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// AdSchedule represents the ad schedule for a channel.
type AdSchedule struct {
	NextAdAt        time.Time `json:"next_ad_at,omitempty"`
	LastAdAt        time.Time `json:"last_ad_at,omitempty"`
	Duration        int       `json:"duration"`
	PrerollFreeTime int       `json:"preroll_free_time"`
	SnoozeCount     int       `json:"snooze_count"`
	SnoozeRefreshAt time.Time `json:"snooze_refresh_at,omitempty"`
}

// GetAdSchedule gets the ad schedule for a channel.
// Requires: channel:read:ads scope.
func (c *Client) GetAdSchedule(ctx context.Context, broadcasterID string) (*AdSchedule, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[AdSchedule]
	if err := c.get(ctx, "/channels/ads", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// SnoozeNextAdResponse represents the response from SnoozeNextAd.
type SnoozeNextAdResponse struct {
	SnoozeCount     int       `json:"snooze_count"`
	SnoozeRefreshAt time.Time `json:"snooze_refresh_at"`
	NextAdAt        time.Time `json:"next_ad_at"`
}

// SnoozeNextAd snoozes the next scheduled ad.
// Requires: channel:manage:ads scope.
func (c *Client) SnoozeNextAd(ctx context.Context, broadcasterID string) (*SnoozeNextAdResponse, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[SnoozeNextAdResponse]
	if err := c.post(ctx, "/channels/ads/schedule/snooze", q, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}
