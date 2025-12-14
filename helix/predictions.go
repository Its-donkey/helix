package helix

import (
	"context"
	"net/url"
	"time"
)

// Prediction represents a channel prediction.
type Prediction struct {
	ID               string              `json:"id"`
	BroadcasterID    string              `json:"broadcaster_id"`
	BroadcasterName  string              `json:"broadcaster_name"`
	BroadcasterLogin string              `json:"broadcaster_login"`
	Title            string              `json:"title"`
	WinningOutcomeID string              `json:"winning_outcome_id,omitempty"`
	Outcomes         []PredictionOutcome `json:"outcomes"`
	PredictionWindow int                 `json:"prediction_window"`
	Status           string              `json:"status"` // ACTIVE, CANCELED, LOCKED, RESOLVED
	CreatedAt        time.Time           `json:"created_at"`
	EndedAt          time.Time           `json:"ended_at,omitempty"`
	LockedAt         time.Time           `json:"locked_at,omitempty"`
}

// PredictionOutcome represents an outcome of a prediction.
type PredictionOutcome struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Users         int                   `json:"users"`
	ChannelPoints int                   `json:"channel_points"`
	TopPredictors []PredictionPredictor `json:"top_predictors,omitempty"`
	Color         string                `json:"color"`
}

// PredictionPredictor represents a top predictor.
type PredictionPredictor struct {
	UserID            string `json:"user_id"`
	UserLogin         string `json:"user_login"`
	UserName          string `json:"user_name"`
	ChannelPointsUsed int    `json:"channel_points_used"`
	ChannelPointsWon  int    `json:"channel_points_won"`
}

// GetPredictionsParams contains parameters for GetPredictions.
type GetPredictionsParams struct {
	BroadcasterID string
	IDs           []string // Prediction IDs (max 100)
	*PaginationParams
}

// GetPredictions gets predictions for a channel.
// Requires: channel:read:predictions scope.
func (c *Client) GetPredictions(ctx context.Context, params *GetPredictionsParams) (*Response[Prediction], error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp Response[Prediction]
	if err := c.get(ctx, "/predictions", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePredictionParams contains parameters for CreatePrediction.
type CreatePredictionParams struct {
	BroadcasterID    string                    `json:"broadcaster_id"`
	Title            string                    `json:"title"`
	Outcomes         []CreatePredictionOutcome `json:"outcomes"`          // 2-10 outcomes
	PredictionWindow int                       `json:"prediction_window"` // 30-1800 seconds
}

// CreatePredictionOutcome represents an outcome when creating a prediction.
type CreatePredictionOutcome struct {
	Title string `json:"title"`
}

// CreatePrediction creates a prediction on a channel.
// Requires: channel:manage:predictions scope.
func (c *Client) CreatePrediction(ctx context.Context, params *CreatePredictionParams) (*Prediction, error) {
	var resp Response[Prediction]
	if err := c.post(ctx, "/predictions", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// EndPredictionParams contains parameters for EndPrediction.
type EndPredictionParams struct {
	BroadcasterID    string `json:"broadcaster_id"`
	ID               string `json:"id"`
	Status           string `json:"status"`                       // RESOLVED, CANCELED, or LOCKED
	WinningOutcomeID string `json:"winning_outcome_id,omitempty"` // Required if status is RESOLVED
}

// EndPrediction ends a prediction.
// Requires: channel:manage:predictions scope.
func (c *Client) EndPrediction(ctx context.Context, params *EndPredictionParams) (*Prediction, error) {
	var resp Response[Prediction]
	if err := c.patch(ctx, "/predictions", nil, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}
