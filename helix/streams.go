package helix

import (
	"context"
	"net/url"
	"time"
)

// Stream represents a live stream.
type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"` // "live" or ""
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Tags         []string  `json:"tags"`
	IsMature     bool      `json:"is_mature"`
}

// GetStreamsParams contains parameters for GetStreams.
type GetStreamsParams struct {
	UserIDs    []string // Filter by user IDs (max 100)
	UserLogins []string // Filter by user logins (max 100)
	GameIDs    []string // Filter by game IDs (max 100)
	Type       string   // "all" or "live"
	Language   []string // Filter by language
	*PaginationParams
}

// GetStreams gets active streams.
func (c *Client) GetStreams(ctx context.Context, params *GetStreamsParams) (*Response[Stream], error) {
	q := url.Values{}
	if params != nil {
		for _, id := range params.UserIDs {
			q.Add("user_id", id)
		}
		for _, login := range params.UserLogins {
			q.Add("user_login", login)
		}
		for _, gameID := range params.GameIDs {
			q.Add("game_id", gameID)
		}
		for _, lang := range params.Language {
			q.Add("language", lang)
		}
		if params.Type != "" {
			q.Set("type", params.Type)
		}
		addPaginationParams(q, params.PaginationParams)
	}

	var resp Response[Stream]
	if err := c.get(ctx, "/streams", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFollowedStreamsParams contains parameters for GetFollowedStreams.
type GetFollowedStreamsParams struct {
	UserID string // Required: The user's ID
	*PaginationParams
}

// GetFollowedStreams gets streams from channels that the user follows.
// Requires: user:read:follows scope.
func (c *Client) GetFollowedStreams(ctx context.Context, params *GetFollowedStreamsParams) (*Response[Stream], error) {
	q := url.Values{}
	q.Set("user_id", params.UserID)
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Stream]
	if err := c.get(ctx, "/streams/followed", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StreamKey represents a stream key.
type StreamKey struct {
	StreamKey string `json:"stream_key"`
}

// GetStreamKey gets the stream key for a broadcaster.
// Requires: channel:read:stream_key scope.
func (c *Client) GetStreamKey(ctx context.Context, broadcasterID string) (*StreamKey, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[StreamKey]
	if err := c.get(ctx, "/streams/key", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// StreamMarker represents a stream marker.
type StreamMarker struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Description     string    `json:"description"`
	PositionSeconds int       `json:"position_seconds"`
}

// CreateStreamMarkerParams contains parameters for CreateStreamMarker.
type CreateStreamMarkerParams struct {
	UserID      string `json:"user_id"`
	Description string `json:"description,omitempty"`
}

// CreateStreamMarker creates a marker in the stream.
// Requires: channel:manage:broadcast scope.
func (c *Client) CreateStreamMarker(ctx context.Context, params *CreateStreamMarkerParams) (*StreamMarker, error) {
	var resp Response[StreamMarker]
	if err := c.post(ctx, "/streams/markers", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// VideoStreamMarkers represents markers for a video.
type VideoStreamMarkers struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Videos    []struct {
		VideoID string         `json:"video_id"`
		Markers []StreamMarker `json:"markers"`
	} `json:"videos"`
}

// GetStreamMarkersParams contains parameters for GetStreamMarkers.
type GetStreamMarkersParams struct {
	UserID  string // Either UserID or VideoID is required
	VideoID string
	*PaginationParams
}

// GetStreamMarkers gets stream markers.
// Requires: user:read:broadcast scope.
func (c *Client) GetStreamMarkers(ctx context.Context, params *GetStreamMarkersParams) (*Response[VideoStreamMarkers], error) {
	q := url.Values{}
	if params.UserID != "" {
		q.Set("user_id", params.UserID)
	}
	if params.VideoID != "" {
		q.Set("video_id", params.VideoID)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[VideoStreamMarkers]
	if err := c.get(ctx, "/streams/markers", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
