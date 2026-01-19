package helix

import (
	"context"
	"io"
	"net/url"
	"time"
)

// Schedule represents a channel's stream schedule.
type Schedule struct {
	Segments         []ScheduleSegment `json:"segments"`
	BroadcasterID    string            `json:"broadcaster_id"`
	BroadcasterName  string            `json:"broadcaster_name"`
	BroadcasterLogin string            `json:"broadcaster_login"`
	Vacation         *Vacation         `json:"vacation,omitempty"`
}

// ScheduleSegment represents a segment in a schedule.
type ScheduleSegment struct {
	ID            string     `json:"id"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	Title         string     `json:"title"`
	CanceledUntil *time.Time `json:"canceled_until,omitempty"`
	Category      *Category  `json:"category,omitempty"`
	IsRecurring   bool       `json:"is_recurring"`
}

// Category represents a category/game.
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Vacation represents a vacation period.
type Vacation struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// GetChannelStreamScheduleParams contains parameters for GetChannelStreamSchedule.
type GetChannelStreamScheduleParams struct {
	BroadcasterID string
	IDs           []string // Segment IDs
	StartTime     time.Time
	UTCOffset     string // e.g., "-04:00"
	*PaginationParams
}

// ScheduleResponse represents the response from GetChannelStreamSchedule.
type ScheduleResponse struct {
	Data       Schedule    `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// GetChannelStreamSchedule gets a channel's stream schedule.
func (c *Client) GetChannelStreamSchedule(ctx context.Context, params *GetChannelStreamScheduleParams) (*ScheduleResponse, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	for _, id := range params.IDs {
		q.Add("id", id)
	}
	if !params.StartTime.IsZero() {
		q.Set("start_time", params.StartTime.Format(time.RFC3339))
	}
	if params.UTCOffset != "" {
		q.Set("utc_offset", params.UTCOffset)
	}
	addPaginationParams(q, params.PaginationParams)

	var resp ScheduleResponse
	if err := c.get(ctx, "/schedule", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetChannelICalendar gets a channel's stream schedule as iCalendar.
func (c *Client) GetChannelICalendar(ctx context.Context, broadcasterID string) (string, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	reqURL := c.baseURL + "/schedule/icalendar?" + q.Encode()

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// UpdateChannelStreamScheduleParams contains parameters for UpdateChannelStreamSchedule.
type UpdateChannelStreamScheduleParams struct {
	BroadcasterID     string     `json:"-"`
	IsVacationEnabled *bool      `json:"is_vacation_enabled,omitempty"`
	VacationStartTime *time.Time `json:"vacation_start_time,omitempty"`
	VacationEndTime   *time.Time `json:"vacation_end_time,omitempty"`
	Timezone          string     `json:"timezone,omitempty"`
}

// UpdateChannelStreamSchedule updates a channel's stream schedule settings.
// Requires: channel:manage:schedule scope.
func (c *Client) UpdateChannelStreamSchedule(ctx context.Context, params *UpdateChannelStreamScheduleParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	if params.IsVacationEnabled != nil {
		if *params.IsVacationEnabled {
			q.Set("is_vacation_enabled", "true")
		} else {
			q.Set("is_vacation_enabled", "false")
		}
	}
	if params.VacationStartTime != nil {
		q.Set("vacation_start_time", params.VacationStartTime.Format(time.RFC3339))
	}
	if params.VacationEndTime != nil {
		q.Set("vacation_end_time", params.VacationEndTime.Format(time.RFC3339))
	}
	if params.Timezone != "" {
		q.Set("timezone", params.Timezone)
	}

	return c.patch(ctx, "/schedule/settings", q, nil, nil)
}

// CreateChannelStreamScheduleSegmentParams contains parameters for CreateChannelStreamScheduleSegment.
type CreateChannelStreamScheduleSegmentParams struct {
	BroadcasterID string    `json:"-"`
	StartTime     time.Time `json:"start_time"`
	Timezone      string    `json:"timezone"`
	Duration      int       `json:"duration"` // minutes (30-1380)
	IsRecurring   bool      `json:"is_recurring,omitempty"`
	CategoryID    string    `json:"category_id,omitempty"`
	Title         string    `json:"title,omitempty"`
}

// CreateChannelStreamScheduleSegment creates a new schedule segment.
// Requires: channel:manage:schedule scope.
func (c *Client) CreateChannelStreamScheduleSegment(ctx context.Context, params *CreateChannelStreamScheduleSegmentParams) (*ScheduleSegment, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)

	var resp struct {
		Data struct {
			Segments []ScheduleSegment `json:"segments"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/schedule/segment", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data.Segments) == 0 {
		return nil, nil
	}
	return &resp.Data.Segments[0], nil
}

// UpdateChannelStreamScheduleSegmentParams contains parameters for UpdateChannelStreamScheduleSegment.
type UpdateChannelStreamScheduleSegmentParams struct {
	BroadcasterID string     `json:"-"`
	ID            string     `json:"-"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	Duration      *int       `json:"duration,omitempty"`
	CategoryID    *string    `json:"category_id,omitempty"`
	Title         *string    `json:"title,omitempty"`
	IsCanceled    *bool      `json:"is_canceled,omitempty"`
	Timezone      string     `json:"timezone,omitempty"`
}

// UpdateChannelStreamScheduleSegment updates a schedule segment.
// Requires: channel:manage:schedule scope.
func (c *Client) UpdateChannelStreamScheduleSegment(ctx context.Context, params *UpdateChannelStreamScheduleSegmentParams) (*ScheduleSegment, error) {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("id", params.ID)

	var resp struct {
		Data struct {
			Segments []ScheduleSegment `json:"segments"`
		} `json:"data"`
	}
	if err := c.patch(ctx, "/schedule/segment", q, params, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data.Segments) == 0 {
		return nil, nil
	}
	return &resp.Data.Segments[0], nil
}

// DeleteChannelStreamScheduleSegment deletes a schedule segment.
// Requires: channel:manage:schedule scope.
func (c *Client) DeleteChannelStreamScheduleSegment(ctx context.Context, broadcasterID, segmentID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("id", segmentID)

	return c.delete(ctx, "/schedule/segment", q, nil)
}
