package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetChannelStreamSchedule(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/schedule" {
			t.Errorf("expected /schedule, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := ScheduleResponse{
			Data: Schedule{
				BroadcasterID:    "12345",
				BroadcasterName:  "Streamer",
				BroadcasterLogin: "streamer",
				Segments: []ScheduleSegment{
					{
						ID:          "seg1",
						StartTime:   time.Now().Add(24 * time.Hour),
						EndTime:     time.Now().Add(27 * time.Hour),
						Title:       "Gaming Stream",
						IsRecurring: true,
						Category:    &Category{ID: "123", Name: "Gaming"},
					},
					{
						ID:          "seg2",
						StartTime:   time.Now().Add(48 * time.Hour),
						EndTime:     time.Now().Add(51 * time.Hour),
						Title:       "Just Chatting",
						IsRecurring: false,
					},
				},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(resp.Data.Segments))
	}
	if resp.Data.Segments[0].Title != "Gaming Stream" {
		t.Errorf("expected 'Gaming Stream', got %s", resp.Data.Segments[0].Title)
	}
}

func TestClient_GetChannelStreamSchedule_WithVacation(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := ScheduleResponse{
			Data: Schedule{
				BroadcasterID: "12345",
				Segments:      []ScheduleSegment{},
				Vacation: &Vacation{
					StartTime: time.Now(),
					EndTime:   time.Now().Add(7 * 24 * time.Hour),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Vacation == nil {
		t.Fatal("expected vacation, got nil")
	}
}

func TestClient_GetChannelStreamSchedule_WithParams(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

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
			Data: Schedule{Segments: []ScheduleSegment{}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetChannelStreamSchedule(context.Background(), &GetChannelStreamScheduleParams{
		BroadcasterID: "12345",
		IDs:           []string{"seg1", "seg2"},
		StartTime:     startTime,
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
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		isVacationEnabled := r.URL.Query().Get("is_vacation_enabled")
		if isVacationEnabled != "true" {
			t.Errorf("expected is_vacation_enabled=true, got %s", isVacationEnabled)
		}

		timezone := r.URL.Query().Get("timezone")
		if timezone != "America/New_York" {
			t.Errorf("expected timezone=America/New_York, got %s", timezone)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	vacationEnabled := true
	err := client.UpdateChannelStreamSchedule(context.Background(), &UpdateChannelStreamScheduleParams{
		BroadcasterID:     "12345",
		IsVacationEnabled: &vacationEnabled,
		Timezone:          "America/New_York",
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
		BroadcasterID:     "12345",
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
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		var params CreateChannelStreamScheduleSegmentParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Duration != 180 {
			t.Errorf("expected duration 180, got %d", params.Duration)
		}
		if params.Title != "New Stream" {
			t.Errorf("expected title 'New Stream', got %s", params.Title)
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
						ID:          "newseg",
						Title:       params.Title,
						StartTime:   params.StartTime,
						IsRecurring: params.IsRecurring,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	startTime := time.Now().Add(24 * time.Hour)
	result, err := client.CreateChannelStreamScheduleSegment(context.Background(), &CreateChannelStreamScheduleSegmentParams{
		BroadcasterID: "12345",
		StartTime:     startTime,
		Timezone:      "America/New_York",
		Duration:      180,
		Title:         "New Stream",
		IsRecurring:   true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != "newseg" {
		t.Errorf("expected segment ID 'newseg', got %s", result.ID)
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

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if segmentID != "seg123" {
			t.Errorf("expected id=seg123, got %s", segmentID)
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
					{ID: "seg123", Title: "Updated Title"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	newTitle := "Updated Title"
	result, err := client.UpdateChannelStreamScheduleSegment(context.Background(), &UpdateChannelStreamScheduleSegmentParams{
		BroadcasterID: "12345",
		ID:            "seg123",
		Title:         &newTitle,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Updated Title" {
		t.Errorf("expected 'Updated Title', got %s", result.Title)
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

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if segmentID != "seg123" {
			t.Errorf("expected id=seg123, got %s", segmentID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteChannelStreamScheduleSegment(context.Background(), "12345", "seg123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
