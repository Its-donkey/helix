package helix

import (
	"context"
	"net/url"
	"time"
)

// Team represents a Twitch team.
type Team struct {
	ID                 string     `json:"id"`
	TeamName           string     `json:"team_name"`
	TeamDisplayName    string     `json:"team_display_name"`
	Info               string     `json:"info"`
	ThumbnailURL       string     `json:"thumbnail_url"`
	BackgroundImageURL string     `json:"background_image_url"`
	Banner             string     `json:"banner"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	Users              []TeamUser `json:"users,omitempty"`
}

// TeamUser represents a user in a team.
type TeamUser struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// ChannelTeam represents a team that a channel belongs to.
type ChannelTeam struct {
	BroadcasterID      string    `json:"broadcaster_id"`
	BroadcasterLogin   string    `json:"broadcaster_login"`
	BroadcasterName    string    `json:"broadcaster_name"`
	BackgroundImageURL string    `json:"background_image_url"`
	Banner             string    `json:"banner"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Info               string    `json:"info"`
	ThumbnailURL       string    `json:"thumbnail_url"`
	TeamName           string    `json:"team_name"`
	TeamDisplayName    string    `json:"team_display_name"`
	ID                 string    `json:"id"`
}

// GetChannelTeams gets the teams a broadcaster belongs to.
func (c *Client) GetChannelTeams(ctx context.Context, broadcasterID string) (*Response[ChannelTeam], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[ChannelTeam]
	if err := c.get(ctx, "/teams/channel", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTeamsParams contains parameters for GetTeams.
type GetTeamsParams struct {
	Name string // Team name
	ID   string // Team ID
}

// GetTeams gets team information.
func (c *Client) GetTeams(ctx context.Context, params *GetTeamsParams) (*Response[Team], error) {
	q := url.Values{}
	if params.Name != "" {
		q.Set("name", params.Name)
	}
	if params.ID != "" {
		q.Set("id", params.ID)
	}

	var resp Response[Team]
	if err := c.get(ctx, "/teams", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
