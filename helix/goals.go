package helix

import (
	"context"
	"net/url"
	"time"
)

// CreatorGoal represents a creator goal.
type CreatorGoal struct {
	ID               string    `json:"id"`
	BroadcasterID    string    `json:"broadcaster_id"`
	BroadcasterName  string    `json:"broadcaster_name"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	Type             string    `json:"type"` // follower, subscription, subscription_count, new_subscription, new_subscription_count
	Description      string    `json:"description"`
	CurrentAmount    int       `json:"current_amount"`
	TargetAmount     int       `json:"target_amount"`
	CreatedAt        time.Time `json:"created_at"`
}

// GetCreatorGoals gets the creator goals for a channel.
// Requires: channel:read:goals scope.
func (c *Client) GetCreatorGoals(ctx context.Context, broadcasterID string) (*Response[CreatorGoal], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[CreatorGoal]
	if err := c.get(ctx, "/goals", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
