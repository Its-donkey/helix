package helix

import (
	"context"
	"net/url"
	"time"
)

// HypeTrainEvent represents a hype train event.
type HypeTrainEvent struct {
	ID             string             `json:"id"`
	EventType      string             `json:"event_type"`
	EventTimestamp time.Time          `json:"event_timestamp"`
	Version        string             `json:"version"`
	EventData      HypeTrainEventData `json:"event_data"`
}

// HypeTrainEventData contains the hype train event data.
type HypeTrainEventData struct {
	ID               string                  `json:"id"`
	BroadcasterID    string                  `json:"broadcaster_id"`
	CooldownEndTime  time.Time               `json:"cooldown_end_time"`
	ExpiresAt        time.Time               `json:"expires_at"`
	Goal             int                     `json:"goal"`
	LastContribution HypeTrainContribution   `json:"last_contribution"`
	Level            int                     `json:"level"`
	StartedAt        time.Time               `json:"started_at"`
	TopContributions []HypeTrainContribution `json:"top_contributions"`
	Total            int                     `json:"total"`
}

// HypeTrainContribution represents a contribution to a hype train.
type HypeTrainContribution struct {
	Total int    `json:"total"`
	Type  string `json:"type"` // BITS, SUBS, OTHER
	User  string `json:"user"`
}

// GetHypeTrainEventsParams contains parameters for GetHypeTrainEvents.
type GetHypeTrainEventsParams struct {
	BroadcasterID string
	*PaginationParams
}

// GetHypeTrainEvents gets hype train events for a channel.
// Requires: channel:read:hype_train scope.
// Note: This endpoint is deprecated; use EventSub instead.
func (c *Client) GetHypeTrainEvents(ctx context.Context, params *GetHypeTrainEventsParams) (*Response[HypeTrainEvent], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[HypeTrainEvent]
	if err := c.get(ctx, "/hypetrain/events", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// HypeTrainStatus represents the current hype train status.
type HypeTrainStatus struct {
	ID               string                  `json:"id"`
	BroadcasterID    string                  `json:"broadcaster_id"`
	Level            int                     `json:"level"`
	Total            int                     `json:"total"`
	Goal             int                     `json:"goal"`
	TopContributions []HypeTrainContribution `json:"top_contributions"`
	LastContribution HypeTrainContribution   `json:"last_contribution"`
	StartedAt        time.Time               `json:"started_at"`
	ExpiresAt        time.Time               `json:"expires_at"`
	CooldownEndTime  time.Time               `json:"cooldown_end_time"`
}

// GetHypeTrainStatus gets the current hype train status for a channel.
// Requires: channel:read:hype_train scope.
func (c *Client) GetHypeTrainStatus(ctx context.Context, broadcasterID string) (*HypeTrainStatus, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[HypeTrainStatus]
	if err := c.get(ctx, "/hypetrain/status", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}
