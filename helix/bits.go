package helix

import (
	"context"
	"net/url"
	"time"
)

// BitsLeaderboard represents a bits leaderboard entry.
type BitsLeaderboard struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Rank      int    `json:"rank"`
	Score     int    `json:"score"`
}

// BitsLeaderboardResponse represents the response from GetBitsLeaderboard.
type BitsLeaderboardResponse struct {
	Data      []BitsLeaderboard `json:"data"`
	DateRange DateRange         `json:"date_range"`
	Total     int               `json:"total"`
}

// DateRange represents a date range.
type DateRange struct {
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// GetBitsLeaderboardParams contains parameters for GetBitsLeaderboard.
type GetBitsLeaderboardParams struct {
	Count     int       // Max entries (1-100)
	Period    string    // day, week, month, year, all
	StartedAt time.Time // Start of the date range
	UserID    string    // Filter to specific user
}

// GetBitsLeaderboard gets the bits leaderboard for a channel.
// Requires: bits:read scope.
func (c *Client) GetBitsLeaderboard(ctx context.Context, params *GetBitsLeaderboardParams) (*BitsLeaderboardResponse, error) {
	q := url.Values{}
	if params != nil {
		if params.Count > 0 {
			q.Set("count", url.QueryEscape(string(rune(params.Count))))
		}
		if params.Period != "" {
			q.Set("period", params.Period)
		}
		if !params.StartedAt.IsZero() {
			q.Set("started_at", params.StartedAt.Format(time.RFC3339))
		}
		if params.UserID != "" {
			q.Set("user_id", params.UserID)
		}
	}

	var resp BitsLeaderboardResponse
	if err := c.get(ctx, "/bits/leaderboard", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Cheermote represents a cheermote.
type Cheermote struct {
	Prefix       string          `json:"prefix"`
	Tiers        []CheermoteTier `json:"tiers"`
	Type         string          `json:"type"`
	Order        int             `json:"order"`
	LastUpdated  time.Time       `json:"last_updated"`
	IsCharitable bool            `json:"is_charitable"`
}

// CheermoteTier represents a tier of a cheermote.
type CheermoteTier struct {
	MinBits        int             `json:"min_bits"`
	ID             string          `json:"id"`
	Color          string          `json:"color"`
	Images         CheermoteImages `json:"images"`
	CanCheer       bool            `json:"can_cheer"`
	ShowInBitsCard bool            `json:"show_in_bits_card"`
}

// CheermoteImages contains the images for a cheermote tier.
type CheermoteImages struct {
	Dark  CheermoteTheme `json:"dark"`
	Light CheermoteTheme `json:"light"`
}

// CheermoteTheme contains themed cheermote images.
type CheermoteTheme struct {
	Animated map[string]string `json:"animated"`
	Static   map[string]string `json:"static"`
}

// GetCheermotes gets the cheermotes available to a broadcaster.
func (c *Client) GetCheermotes(ctx context.Context, broadcasterID string) (*Response[Cheermote], error) {
	q := url.Values{}
	if broadcasterID != "" {
		q.Set("broadcaster_id", broadcasterID)
	}

	var resp Response[Cheermote]
	if err := c.get(ctx, "/bits/cheermotes", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
