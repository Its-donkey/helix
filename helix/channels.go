package helix

import (
	"context"
	"net/url"
	"time"
)

// Channel represents a Twitch channel.
type Channel struct {
	BroadcasterID               string   `json:"broadcaster_id"`
	BroadcasterLogin            string   `json:"broadcaster_login"`
	BroadcasterName             string   `json:"broadcaster_name"`
	BroadcasterLanguage         string   `json:"broadcaster_language"`
	GameID                      string   `json:"game_id"`
	GameName                    string   `json:"game_name"`
	Title                       string   `json:"title"`
	Delay                       int      `json:"delay"`
	Tags                        []string `json:"tags"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
	IsBrandedContent            bool     `json:"is_branded_content"`
}

// GetChannelInformationParams contains parameters for GetChannelInformation.
type GetChannelInformationParams struct {
	BroadcasterIDs []string // Up to 100 broadcaster IDs
}

// GetChannelInformation gets channel information for one or more users.
func (c *Client) GetChannelInformation(ctx context.Context, params *GetChannelInformationParams) (*Response[Channel], error) {
	q := url.Values{}
	for _, id := range params.BroadcasterIDs {
		q.Add("broadcaster_id", id)
	}

	var resp Response[Channel]
	if err := c.get(ctx, "/channels", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ModifyChannelInformationParams contains parameters for ModifyChannelInformation.
type ModifyChannelInformationParams struct {
	BroadcasterID               string                              `json:"-"`
	GameID                      string                              `json:"game_id,omitempty"`
	BroadcasterLanguage         string                              `json:"broadcaster_language,omitempty"`
	Title                       string                              `json:"title,omitempty"`
	Delay                       *int                                `json:"delay,omitempty"`
	Tags                        []string                            `json:"tags,omitempty"`
	ContentClassificationLabels []ContentClassificationLabelSetting `json:"content_classification_labels,omitempty"`
	IsBrandedContent            *bool                               `json:"is_branded_content,omitempty"`
}

// ContentClassificationLabelSetting represents a content classification label setting for modifying a channel.
type ContentClassificationLabelSetting struct {
	ID        string `json:"id"`
	IsEnabled bool   `json:"is_enabled"`
}

// ModifyChannelInformation modifies channel information.
// Requires: channel:manage:broadcast scope.
func (c *Client) ModifyChannelInformation(ctx context.Context, params *ModifyChannelInformationParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)

	return c.patch(ctx, "/channels", q, params, nil)
}

// ChannelEditor represents a channel editor.
type ChannelEditor struct {
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
}

// GetChannelEditors gets the list of channel editors.
// Requires: channel:read:editors scope.
func (c *Client) GetChannelEditors(ctx context.Context, broadcasterID string) (*Response[ChannelEditor], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[ChannelEditor]
	if err := c.get(ctx, "/channels/editors", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FollowedChannel represents a channel that a user follows.
type FollowedChannel struct {
	BroadcasterID    string    `json:"broadcaster_id"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	BroadcasterName  string    `json:"broadcaster_name"`
	FollowedAt       time.Time `json:"followed_at"`
}

// GetFollowedChannelsParams contains parameters for GetFollowedChannels.
type GetFollowedChannelsParams struct {
	UserID        string // Required: The user's ID
	BroadcasterID string // Optional: Filter by broadcaster ID
	*PaginationParams
}

// GetFollowedChannels gets the list of channels that a user follows.
// Requires: user:read:follows scope.
func (c *Client) GetFollowedChannels(ctx context.Context, params *GetFollowedChannelsParams) (*Response[FollowedChannel], error) {
	q := url.Values{}
	q.Set("user_id", params.UserID)
	if params.BroadcasterID != "" {
		q.Set("broadcaster_id", params.BroadcasterID)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[FollowedChannel]
	if err := c.get(ctx, "/channels/followed", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChannelFollower represents a user who follows a channel.
type ChannelFollower struct {
	UserID     string    `json:"user_id"`
	UserLogin  string    `json:"user_login"`
	UserName   string    `json:"user_name"`
	FollowedAt time.Time `json:"followed_at"`
}

// GetChannelFollowersParams contains parameters for GetChannelFollowers.
type GetChannelFollowersParams struct {
	BroadcasterID string // Required: The broadcaster's ID
	UserID        string // Optional: Filter by user ID
	*PaginationParams
}

// GetChannelFollowers gets the list of users that follow a channel.
// Requires: moderator:read:followers scope.
func (c *Client) GetChannelFollowers(ctx context.Context, params *GetChannelFollowersParams) (*Response[ChannelFollower], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	if params.UserID != "" {
		q.Set("user_id", params.UserID)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[ChannelFollower]
	if err := c.get(ctx, "/channels/followers", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VIP represents a VIP user in a channel.
type VIP struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// GetVIPsParams contains parameters for GetVIPs.
type GetVIPsParams struct {
	BroadcasterID string
	UserIDs       []string // Filter by user IDs (max 100)
	*PaginationParams
}

// GetVIPs gets the list of VIPs for a channel.
// Requires: channel:read:vips scope.
func (c *Client) GetVIPs(ctx context.Context, params *GetVIPsParams) (*Response[VIP], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.UserIDs {
		q.Add("user_id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[VIP]
	if err := c.get(ctx, "/channels/vips", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddChannelVIP adds a VIP to the channel.
// Requires: channel:manage:vips scope.
func (c *Client) AddChannelVIP(ctx context.Context, broadcasterID, userID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("user_id", userID)

	return c.post(ctx, "/channels/vips", q, nil, nil)
}

// RemoveChannelVIP removes a VIP from the channel.
// Requires: channel:manage:vips scope.
func (c *Client) RemoveChannelVIP(ctx context.Context, broadcasterID, userID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("user_id", userID)

	return c.delete(ctx, "/channels/vips", q, nil)
}
