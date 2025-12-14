package helix

import (
	"context"
	"net/url"
	"time"
)

// Video represents a Twitch video (VOD, highlight, or upload).
type Video struct {
	ID            string         `json:"id"`
	StreamID      string         `json:"stream_id,omitempty"`
	UserID        string         `json:"user_id"`
	UserLogin     string         `json:"user_login"`
	UserName      string         `json:"user_name"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
	PublishedAt   time.Time      `json:"published_at"`
	URL           string         `json:"url"`
	ThumbnailURL  string         `json:"thumbnail_url"`
	Viewable      string         `json:"viewable"`
	ViewCount     int            `json:"view_count"`
	Language      string         `json:"language"`
	Type          string         `json:"type"` // archive, highlight, upload
	Duration      string         `json:"duration"`
	MutedSegments []MutedSegment `json:"muted_segments,omitempty"`
}

// MutedSegment represents a muted segment in a video.
type MutedSegment struct {
	Duration int `json:"duration"`
	Offset   int `json:"offset"`
}

// GetVideosParams contains parameters for GetVideos.
type GetVideosParams struct {
	IDs      []string // Video IDs (max 100)
	UserID   string
	GameID   string
	Language string
	Period   string // all, day, week, month
	Sort     string // time, trending, views
	Type     string // all, archive, highlight, upload
	*PaginationParams
}

// GetVideos gets videos.
func (c *Client) GetVideos(ctx context.Context, params *GetVideosParams) (*Response[Video], error) {
	q := url.Values{}
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	if params.UserID != "" {
		q.Set("user_id", params.UserID)
	}
	if params.GameID != "" {
		q.Set("game_id", params.GameID)
	}
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	if params.Period != "" {
		q.Set("period", params.Period)
	}
	if params.Sort != "" {
		q.Set("sort", params.Sort)
	}
	if params.Type != "" {
		q.Set("type", params.Type)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Video]
	if err := c.get(ctx, "/videos", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVideosResponse represents the response from DeleteVideos.
type DeleteVideosResponse []string

// DeleteVideos deletes one or more videos.
// Requires: channel:manage:videos scope.
func (c *Client) DeleteVideos(ctx context.Context, videoIDs []string) ([]string, error) {
	q := url.Values{}
	for _, id := range videoIDs {
		q.Add("id", id)
	}

	var resp struct {
		Data []string `json:"data"`
	}
	if err := c.delete(ctx, "/videos", q, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
