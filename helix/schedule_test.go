package helix

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// Official Twitch API example values from https://dev.twitch.tv/docs/api/reference/#get-channel-stream-schedule
const (
	// Broadcaster info
	twitchScheduleBroadcasterID    = "141981764"
	twitchScheduleBroadcasterLogin = "twitchdev"
	twitchScheduleBroadcasterName  = "TwitchDev"

	// Segment IDs
	twitchScheduleSegmentID1 = "eyJzZWdtZW50SUQiOiJlNGFjYzcyNC0zNzFmLTQwMmMtODFjYS0yM2FkYTc5NzU5ZDQiLCJpc29ZZWFyIjoyMDIxLCJpc29XZWVrIjoyNn0="
	twitchScheduleSegmentID2 = "eyJzZWdtZW50SUQiOiI4Y2EwN2E2NC0xYTZkLTRjYWItYWE5Ni0xNjIyYzNjYWUzZDkiLCJpc29ZZWFyIjoyMDIxLCJpc29XZWVrIjoyMX0="

	// Category info
	twitchScheduleCategoryID   = "509670"
	twitchScheduleCategoryName = "Science & Technology"

	// Stream info
	twitchScheduleTitle    = "TwitchDev Monthly Update // July 1, 2021"
	twitchScheduleTimezone = "America/New_York"
	twitchScheduleDuration = 60 // minutes

	// iCalendar example from https://dev.twitch.tv/docs/api/reference/#get-channel-icalendar
	twitchScheduleICalendar = `BEGIN:VCALENDAR
PRODID:-//twitch.tv//StreamSchedule//1.0
VERSION:2.0
CALSCALE:GREGORIAN
REFRESH-INTERVAL;VALUE=DURATION:PT1H
NAME:TwitchDev
BEGIN:VEVENT
UID:e4acc724-371f-402c-81ca-23ada79759d4
DTSTAMP:20210323T040131Z
DTSTART;TZID=/America/New_York:20210701T140000
DTEND;TZID=/America/New_York:20210701T150000
SUMMARY:TwitchDev Monthly Update // July 1, 2021
DESCRIPTION:Science & Technology.
CATEGORIES:Science & Technology
END:VEVENT
END:VCALENDAR`
)

// Date/time values from Twitch API examples
var (
	// Stream times from https://dev.twitch.tv/docs/api/reference/#get-channel-stream-schedule
	twitchScheduleStartTime = time.Date(2021, 7, 1, 18, 0, 0, 0, time.UTC)
	twitchScheduleEndTime   = time.Date(2021, 7, 1, 19, 0, 0, 0, time.UTC)

	// Vacation times from https://dev.twitch.tv/docs/api/reference/#update-channel-stream-schedule
	twitchScheduleVacationStart = time.Date(2021, 5, 16, 0, 0, 0, 0, time.UTC)
	twitchScheduleVacationEnd   = time.Date(2021, 5, 23, 0, 0, 0, 0, time.UTC)
)

// scheduleErrorTransport is a RoundTripper that always returns an error.
type scheduleErrorTransport struct{}

func (t *scheduleErrorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network error")
}

// scheduleErrorBodyTransport returns a response with a body that errors on read.
type scheduleErrorBodyTransport struct{}

func (t *scheduleErrorBodyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&scheduleErrorReader{}),
	}, nil
}

type scheduleErrorReader struct{}

func (r *scheduleErrorReader) Read([]byte) (int, error) {
	return 0, errors.New("read error")
}

func TestClient_GetChannelStreamSchedule(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/schedule" {
			t.Errorf("expected /schedule, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != twitchScheduleBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchScheduleBroadcasterID, broadcasterID)
		}

		resp := ScheduleResponse{
			Data: Schedule{
				BroadcasterID:    twitchScheduleBroadcasterID,
				BroadcasterName:  twitchScheduleBroadcasterName,
				BroadcasterLogin: twitchScheduleBroadcasterLogin,
				Segments: []ScheduleSegment{
					{
						ID:          twitchScheduleSegmentID1,
						StartTime:   twitchScheduleStartTime,
						EndTime:     twitchScheduleEndTime,
						Title:       twitchScheduleTitle,
						IsRecurring: false,
						Category:    &Category{ID: twitchScheduleCategoryID, Name: twitchScheduleCategoryName},
					},
					{
						ID:          twitchScheduleSegmentID2,
						StartTime:   twitchScheduleStartTime.Add(24 * time.Hour),
						EndTime:     twitchScheduleEndTime.Add(24 * time.Hour),
						Title:       twitchScheduleTitle,
						IsRecurring: true,
					},
				},
			},
			Pagination: &Pagination{Cursor: "eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6Mn19"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: twitchScheduleBroadcasterID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(resp.Data.Segments))
	}
	if resp.Data.Segments[0].Title != twitchScheduleTitle {
		t.Errorf("expected %q, got %s", twitchScheduleTitle, resp.Data.Segments[0].Title)
	}
}

func TestClient_GetChannelStreamSchedule_WithVacation(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := ScheduleResponse{
			Data: Schedule{
				BroadcasterID:    twitchScheduleBroadcasterID,
				BroadcasterName:  twitchScheduleBroadcasterName,
				BroadcasterLogin: twitchScheduleBroadcasterLogin,
				Segments:         []ScheduleSegment{},
				Vacation: &Vacation{
					StartTime: twitchScheduleVacationStart,
					EndTime:   twitchScheduleVacationEnd,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: twitchScheduleBroadcasterID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Vacation == nil {
		t.Fatal("expected vacation, got nil")
	}
}

func TestClient_GetChannelStreamSchedule_WithParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 segment ids, got %d", len(ids))
		}

		utcOffset := r.URL.Query().Get("utc_offset")
		if utcOffset != "-05:00" {
			t.Errorf("expected utc_offset=-05:00, got %s", utcOffset)
		}

		startTimeParam := r.URL.Query().Get("start_time")
		if startTimeParam == "" {
			t.Error("expected start_time to be set")
		}

		resp := ScheduleResponse{
			Data: Schedule{
				BroadcasterID: twitchScheduleBroadcasterID,
				Segments:      []ScheduleSegment{},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		IDs:           []string{twitchScheduleSegmentID1, twitchScheduleSegmentID2},
		StartTime:     twitchScheduleStartTime,
		UTCOffset:     "-05:00",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateChannelStreamSchedule(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/schedule/settings" {
			t.Errorf("expected /schedule/settings, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != twitchScheduleBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchScheduleBroadcasterID, broadcasterID)
		}

		isVacationEnabled := r.URL.Query().Get("is_vacation_enabled")
		if isVacationEnabled != "true" {
			t.Errorf("expected is_vacation_enabled=true, got %s", isVacationEnabled)
		}

		timezone := r.URL.Query().Get("timezone")
		if timezone != twitchScheduleTimezone {
			t.Errorf("expected timezone=%s, got %s", twitchScheduleTimezone, timezone)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	vacationEnabled := true
	err := client.UpdateChannelStreamSchedule(context.Background(), &UpdateChannelStreamScheduleParams{
		BroadcasterID:     twitchScheduleBroadcasterID,
		IsVacationEnabled: &vacationEnabled,
		Timezone:          twitchScheduleTimezone,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateChannelStreamSchedule_DisableVacation(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		isVacationEnabled := r.URL.Query().Get("is_vacation_enabled")
		if isVacationEnabled != "false" {
			t.Errorf("expected is_vacation_enabled=false, got %s", isVacationEnabled)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	vacationEnabled := false
	err := client.UpdateChannelStreamSchedule(context.Background(), &UpdateChannelStreamScheduleParams{
		BroadcasterID:     twitchScheduleBroadcasterID,
		IsVacationEnabled: &vacationEnabled,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateChannelStreamScheduleSegment(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/schedule/segment" {
			t.Errorf("expected /schedule/segment, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != twitchScheduleBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchScheduleBroadcasterID, broadcasterID)
		}

		var params CreateChannelStreamScheduleSegmentParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Duration != twitchScheduleDuration {
			t.Errorf("expected duration %d, got %d", twitchScheduleDuration, params.Duration)
		}
		if params.Title != twitchScheduleTitle {
			t.Errorf("expected title %q, got %s", twitchScheduleTitle, params.Title)
		}

		resp := struct {
			Data struct {
				Segments []ScheduleSegment `json:"segments"`
			} `json:"data"`
		}{
			Data: struct {
				Segments []ScheduleSegment `json:"segments"`
			}{
				Segments: []ScheduleSegment{
					{
						ID:          twitchScheduleSegmentID1,
						Title:       params.Title,
						StartTime:   twitchScheduleStartTime,
						EndTime:     twitchScheduleEndTime,
						IsRecurring: params.IsRecurring,
						Category:    &Category{ID: twitchScheduleCategoryID, Name: twitchScheduleCategoryName},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateChannelStreamScheduleSegment(context.Background(), &CreateChannelStreamScheduleSegmentParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		StartTime:     twitchScheduleStartTime,
		Timezone:      twitchScheduleTimezone,
		Duration:      twitchScheduleDuration,
		Title:         twitchScheduleTitle,
		CategoryID:    twitchScheduleCategoryID,
		IsRecurring:   false,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != twitchScheduleSegmentID1 {
		t.Errorf("expected segment ID %q, got %s", twitchScheduleSegmentID1, result.ID)
	}
}

func TestClient_UpdateChannelStreamScheduleSegment(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/schedule/segment" {
			t.Errorf("expected /schedule/segment, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		segmentID := r.URL.Query().Get("id")

		if broadcasterID != twitchScheduleBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchScheduleBroadcasterID, broadcasterID)
		}
		if segmentID != twitchScheduleSegmentID1 {
			t.Errorf("expected id=%s, got %s", twitchScheduleSegmentID1, segmentID)
		}

		// Response with updated duration (120 minutes) from Twitch example
		updatedEndTime := twitchScheduleStartTime.Add(120 * time.Minute)
		resp := struct {
			Data struct {
				Segments []ScheduleSegment `json:"segments"`
			} `json:"data"`
		}{
			Data: struct {
				Segments []ScheduleSegment `json:"segments"`
			}{
				Segments: []ScheduleSegment{
					{
						ID:        twitchScheduleSegmentID1,
						Title:     twitchScheduleTitle,
						StartTime: twitchScheduleStartTime,
						EndTime:   updatedEndTime,
						Category:  &Category{ID: twitchScheduleCategoryID, Name: twitchScheduleCategoryName},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	duration := 120 // Updated from 60 to 120 as in Twitch example
	result, err := client.UpdateChannelStreamScheduleSegment(context.Background(), &UpdateChannelStreamScheduleSegmentParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		ID:            twitchScheduleSegmentID1,
		Duration:      &duration,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != twitchScheduleTitle {
		t.Errorf("expected %q, got %s", twitchScheduleTitle, result.Title)
	}
}

func TestClient_DeleteChannelStreamScheduleSegment(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/schedule/segment" {
			t.Errorf("expected /schedule/segment, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		segmentID := r.URL.Query().Get("id")

		if broadcasterID != twitchScheduleBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchScheduleBroadcasterID, broadcasterID)
		}
		// Using segment ID 2 as in delete example
		if segmentID != twitchScheduleSegmentID2 {
			t.Errorf("expected id=%s, got %s", twitchScheduleSegmentID2, segmentID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteChannelStreamScheduleSegment(context.Background(), twitchScheduleBroadcasterID, twitchScheduleSegmentID2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetChannelICalendar(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/schedule/icalendar" {
			t.Errorf("expected /schedule/icalendar, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != twitchScheduleBroadcasterID {
			t.Errorf("expected broadcaster_id=%s, got %s", twitchScheduleBroadcasterID, broadcasterID)
		}

		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(twitchScheduleICalendar))
	})
	defer server.Close()

	result, err := client.GetChannelICalendar(context.Background(), twitchScheduleBroadcasterID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != twitchScheduleICalendar {
		t.Errorf("expected iCalendar content, got %q", result)
	}
}

func TestClient_GetChannelICalendar_Error(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{
			Transport: &scheduleErrorTransport{},
		},
		baseURL: "http://invalid",
	}

	_, err := client.GetChannelICalendar(context.Background(), twitchScheduleBroadcasterID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetChannelICalendar_ReadError(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{
			Transport: &scheduleErrorBodyTransport{},
		},
		baseURL: "http://test",
	}

	_, err := client.GetChannelICalendar(context.Background(), twitchScheduleBroadcasterID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateChannelStreamScheduleSegment_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data struct {
				Segments []ScheduleSegment `json:"segments"`
			} `json:"data"`
		}{
			Data: struct {
				Segments []ScheduleSegment `json:"segments"`
			}{
				Segments: []ScheduleSegment{},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateChannelStreamScheduleSegment(context.Background(), &CreateChannelStreamScheduleSegmentParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		StartTime:     twitchScheduleStartTime,
		Timezone:      twitchScheduleTimezone,
		Duration:      twitchScheduleDuration,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty segments, got %v", result)
	}
}

func TestClient_UpdateChannelStreamScheduleSegment_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data struct {
				Segments []ScheduleSegment `json:"segments"`
			} `json:"data"`
		}{
			Data: struct {
				Segments []ScheduleSegment `json:"segments"`
			}{
				Segments: []ScheduleSegment{},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	duration := 120
	result, err := client.UpdateChannelStreamScheduleSegment(context.Background(), &UpdateChannelStreamScheduleSegmentParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		ID:            twitchScheduleSegmentID1,
		Duration:      &duration,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty segments, got %v", result)
	}
}

func TestClient_UpdateChannelStreamSchedule_WithVacationTimes(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		vacationStart := r.URL.Query().Get("vacation_start_time")
		vacationEnd := r.URL.Query().Get("vacation_end_time")

		if vacationStart == "" {
			t.Error("expected vacation_start_time to be set")
		}
		if vacationEnd == "" {
			t.Error("expected vacation_end_time to be set")
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.UpdateChannelStreamSchedule(context.Background(), &UpdateChannelStreamScheduleParams{
		BroadcasterID:     twitchScheduleBroadcasterID,
		VacationStartTime: &twitchScheduleVacationStart,
		VacationEndTime:   &twitchScheduleVacationEnd,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetChannelStreamSchedule_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: twitchScheduleBroadcasterID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateChannelStreamSchedule_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	err := client.UpdateChannelStreamSchedule(context.Background(), &UpdateChannelStreamScheduleParams{
		BroadcasterID: twitchScheduleBroadcasterID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateChannelStreamScheduleSegment_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.CreateChannelStreamScheduleSegment(context.Background(), &CreateChannelStreamScheduleSegmentParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		StartTime:     twitchScheduleStartTime,
		Timezone:      twitchScheduleTimezone,
		Duration:      twitchScheduleDuration,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateChannelStreamScheduleSegment_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	duration := 120
	_, err := client.UpdateChannelStreamScheduleSegment(context.Background(), &UpdateChannelStreamScheduleSegmentParams{
		BroadcasterID: twitchScheduleBroadcasterID,
		ID:            twitchScheduleSegmentID1,
		Duration:      &duration,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_DeleteChannelStreamScheduleSegment_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	err := client.DeleteChannelStreamScheduleSegment(context.Background(), twitchScheduleBroadcasterID, twitchScheduleSegmentID2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
