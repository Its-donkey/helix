package helix

import (
	"context"
	"net/url"
	"time"
)

// CustomReward represents a channel points custom reward.
type CustomReward struct {
	BroadcasterID                     string              `json:"broadcaster_id"`
	BroadcasterLogin                  string              `json:"broadcaster_login"`
	BroadcasterName                   string              `json:"broadcaster_name"`
	ID                                string              `json:"id"`
	Title                             string              `json:"title"`
	Prompt                            string              `json:"prompt"`
	Cost                              int                 `json:"cost"`
	Image                             *RewardImage        `json:"image,omitempty"`
	DefaultImage                      RewardImage         `json:"default_image"`
	BackgroundColor                   string              `json:"background_color"`
	IsEnabled                         bool                `json:"is_enabled"`
	IsUserInputRequired               bool                `json:"is_user_input_required"`
	MaxPerStreamSetting               MaxPerStream        `json:"max_per_stream_setting"`
	MaxPerUserPerStreamSetting        MaxPerUserPerStream `json:"max_per_user_per_stream_setting"`
	GlobalCooldownSetting             GlobalCooldown      `json:"global_cooldown_setting"`
	IsPaused                          bool                `json:"is_paused"`
	IsInStock                         bool                `json:"is_in_stock"`
	ShouldRedemptionsSkipRequestQueue bool                `json:"should_redemptions_skip_request_queue"`
	RedemptionsRedeemedCurrentStream  int                 `json:"redemptions_redeemed_current_stream,omitempty"`
	CooldownExpiresAt                 *time.Time          `json:"cooldown_expires_at,omitempty"`
}

// RewardImage represents images for a reward.
type RewardImage struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

// MaxPerStream represents max per stream settings.
type MaxPerStream struct {
	IsEnabled    bool `json:"is_enabled"`
	MaxPerStream int  `json:"max_per_stream"`
}

// MaxPerUserPerStream represents max per user per stream settings.
type MaxPerUserPerStream struct {
	IsEnabled           bool `json:"is_enabled"`
	MaxPerUserPerStream int  `json:"max_per_user_per_stream"`
}

// GlobalCooldown represents global cooldown settings.
type GlobalCooldown struct {
	IsEnabled             bool `json:"is_enabled"`
	GlobalCooldownSeconds int  `json:"global_cooldown_seconds"`
}

// GetCustomRewardsParams contains parameters for GetCustomRewards.
type GetCustomRewardsParams struct {
	BroadcasterID         string
	IDs                   []string // Reward IDs (max 50)
	OnlyManageableRewards bool
}

// GetCustomRewards gets custom rewards for a channel.
// Requires: channel:read:redemptions or channel:manage:redemptions scope.
func (c *Client) GetCustomRewards(ctx context.Context, params *GetCustomRewardsParams) (*Response[CustomReward], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	if params.OnlyManageableRewards {
		q.Set("only_manageable_rewards", "true")
	}

	var resp Response[CustomReward]
	if err := c.get(ctx, "/channel_points/custom_rewards", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateCustomRewardParams contains parameters for CreateCustomReward.
type CreateCustomRewardParams struct {
	BroadcasterID                     string `json:"-"`
	Title                             string `json:"title"`
	Cost                              int    `json:"cost"`
	Prompt                            string `json:"prompt,omitempty"`
	IsEnabled                         *bool  `json:"is_enabled,omitempty"`
	BackgroundColor                   string `json:"background_color,omitempty"`
	IsUserInputRequired               *bool  `json:"is_user_input_required,omitempty"`
	IsMaxPerStreamEnabled             *bool  `json:"is_max_per_stream_enabled,omitempty"`
	MaxPerStream                      *int   `json:"max_per_stream,omitempty"`
	IsMaxPerUserPerStreamEnabled      *bool  `json:"is_max_per_user_per_stream_enabled,omitempty"`
	MaxPerUserPerStream               *int   `json:"max_per_user_per_stream,omitempty"`
	IsGlobalCooldownEnabled           *bool  `json:"is_global_cooldown_enabled,omitempty"`
	GlobalCooldownSeconds             *int   `json:"global_cooldown_seconds,omitempty"`
	ShouldRedemptionsSkipRequestQueue *bool  `json:"should_redemptions_skip_request_queue,omitempty"`
}

// CreateCustomReward creates a custom reward.
// Requires: channel:manage:redemptions scope.
func (c *Client) CreateCustomReward(ctx context.Context, params *CreateCustomRewardParams) (*CustomReward, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)

	var resp Response[CustomReward]
	if err := c.post(ctx, "/channel_points/custom_rewards", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UpdateCustomRewardParams contains parameters for UpdateCustomReward.
type UpdateCustomRewardParams struct {
	BroadcasterID                     string `json:"-"`
	ID                                string `json:"-"`
	Title                             string `json:"title,omitempty"`
	Cost                              *int   `json:"cost,omitempty"`
	Prompt                            string `json:"prompt,omitempty"`
	IsEnabled                         *bool  `json:"is_enabled,omitempty"`
	BackgroundColor                   string `json:"background_color,omitempty"`
	IsUserInputRequired               *bool  `json:"is_user_input_required,omitempty"`
	IsMaxPerStreamEnabled             *bool  `json:"is_max_per_stream_enabled,omitempty"`
	MaxPerStream                      *int   `json:"max_per_stream,omitempty"`
	IsMaxPerUserPerStreamEnabled      *bool  `json:"is_max_per_user_per_stream_enabled,omitempty"`
	MaxPerUserPerStream               *int   `json:"max_per_user_per_stream,omitempty"`
	IsGlobalCooldownEnabled           *bool  `json:"is_global_cooldown_enabled,omitempty"`
	GlobalCooldownSeconds             *int   `json:"global_cooldown_seconds,omitempty"`
	ShouldRedemptionsSkipRequestQueue *bool  `json:"should_redemptions_skip_request_queue,omitempty"`
	IsPaused                          *bool  `json:"is_paused,omitempty"`
}

// UpdateCustomReward updates a custom reward.
// Requires: channel:manage:redemptions scope.
func (c *Client) UpdateCustomReward(ctx context.Context, params *UpdateCustomRewardParams) (*CustomReward, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("id", params.ID)

	var resp Response[CustomReward]
	if err := c.patch(ctx, "/channel_points/custom_rewards", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// DeleteCustomReward deletes a custom reward.
// Requires: channel:manage:redemptions scope.
func (c *Client) DeleteCustomReward(ctx context.Context, broadcasterID, rewardID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("id", rewardID)

	return c.delete(ctx, "/channel_points/custom_rewards", q, nil)
}

// CustomRewardRedemption represents a redemption of a custom reward.
type CustomRewardRedemption struct {
	BroadcasterID    string    `json:"broadcaster_id"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	BroadcasterName  string    `json:"broadcaster_name"`
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	UserLogin        string    `json:"user_login"`
	UserName         string    `json:"user_name"`
	UserInput        string    `json:"user_input"`
	Status           string    `json:"status"` // CANCELED, FULFILLED, UNFULFILLED
	RedeemedAt       time.Time `json:"redeemed_at"`
	Reward           struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Prompt string `json:"prompt"`
		Cost   int    `json:"cost"`
	} `json:"reward"`
}

// GetCustomRewardRedemptionsParams contains parameters for GetCustomRewardRedemptions.
type GetCustomRewardRedemptionsParams struct {
	BroadcasterID string
	RewardID      string
	Status        string // CANCELED, FULFILLED, UNFULFILLED
	IDs           []string
	Sort          string // OLDEST, NEWEST
	*PaginationParams
}

// GetCustomRewardRedemptions gets redemptions for a custom reward.
// Requires: channel:read:redemptions or channel:manage:redemptions scope.
func (c *Client) GetCustomRewardRedemptions(ctx context.Context, params *GetCustomRewardRedemptionsParams) (*Response[CustomRewardRedemption], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("reward_id", params.RewardID)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	if params.Sort != "" {
		q.Set("sort", params.Sort)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[CustomRewardRedemption]
	if err := c.get(ctx, "/channel_points/custom_rewards/redemptions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRedemptionStatusParams contains parameters for UpdateRedemptionStatus.
type UpdateRedemptionStatusParams struct {
	BroadcasterID string   `json:"-"`
	RewardID      string   `json:"-"`
	IDs           []string `json:"-"`      // Redemption IDs (max 50)
	Status        string   `json:"status"` // CANCELED or FULFILLED
}

// UpdateRedemptionStatus updates the status of custom reward redemptions.
// Requires: channel:manage:redemptions scope.
func (c *Client) UpdateRedemptionStatus(ctx context.Context, params *UpdateRedemptionStatusParams) (*Response[CustomRewardRedemption], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("reward_id", params.RewardID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}

	var resp Response[CustomRewardRedemption]
	if err := c.patch(ctx, "/channel_points/custom_rewards/redemptions", q, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
