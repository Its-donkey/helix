package helix

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// roundTripperFunc is a helper type for testing custom HTTP transports.
type roundTripperFunc struct {
	fn func(*http.Request) (*http.Response, error)
}

func (rt *roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.fn(req)
}

func TestClient_GetIngestServers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/ingests" {
			t.Errorf("expected /ingests, got %s", r.URL.Path)
		}

		resp := IngestServersResponse{
			Ingests: []IngestServer{
				{
					ID:           1,
					Availability: 1.0,
					Default:      false,
					Name:         "US East: Atlanta, GA",
					URLTemplate:  "rtmp://atl.contribute.live-video.net/app/{stream_key}",
					Priority:     13,
				},
				{
					ID:           2,
					Availability: 1.0,
					Default:      true,
					Name:         "US East: Ashburn, VA",
					URLTemplate:  "rtmp://iad05.contribute.live-video.net/app/{stream_key}",
					Priority:     10,
				},
				{
					ID:           3,
					Availability: 1.0,
					Default:      false,
					Name:         "EU West: Amsterdam, NL",
					URLTemplate:  "rtmp://ams03.contribute.live-video.net/app/{stream_key}",
					Priority:     25,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("test", authClient, WithIngestBaseURL(server.URL), WithHTTPClient(server.Client()))

	resp, err := client.GetIngestServers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Ingests) != 3 {
		t.Errorf("expected 3 ingest servers, got %d", len(resp.Ingests))
	}
	if resp.Ingests[0].Name != "US East: Atlanta, GA" {
		t.Errorf("expected first server name 'US East: Atlanta, GA', got %s", resp.Ingests[0].Name)
	}
}

func TestClient_GetIngestServers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("test", authClient, WithIngestBaseURL(server.URL), WithHTTPClient(server.Client()))

	_, err := client.GetIngestServers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestClient_GetIngestServers_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("test", authClient, WithIngestBaseURL(server.URL), WithHTTPClient(server.Client()))

	_, err := client.GetIngestServers(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestClient_GetIngestServers_DefaultURL(t *testing.T) {
	// Test that empty ingestBaseURL defaults to IngestBaseURL constant
	// Use a custom transport to intercept the request and verify the URL
	var requestedURL string
	transport := &roundTripperFunc{fn: func(req *http.Request) (*http.Response, error) {
		requestedURL = req.URL.String()
		// Return a valid response
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ingests":[]}`)),
			Header:     make(http.Header),
		}
		return resp, nil
	}}

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("test", authClient, WithHTTPClient(&http.Client{Transport: transport}))
	client.ingestBaseURL = "" // Clear it to test default

	_, err := client.GetIngestServers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the default URL was used
	expectedURL := IngestBaseURL + "/ingests"
	if requestedURL != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, requestedURL)
	}
}

func TestClient_GetIngestServers_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never respond
		<-r.Context().Done()
	}))
	defer server.Close()

	authClient := NewAuthClient(AuthConfig{ClientID: "test"})
	client := NewClient("test", authClient, WithIngestBaseURL(server.URL), WithHTTPClient(server.Client()))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetIngestServers(ctx)
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
}

func TestGetIngestServerByName(t *testing.T) {
	resp := &IngestServersResponse{
		Ingests: []IngestServer{
			{Name: "US East: Atlanta, GA", URLTemplate: "rtmp://atl.contribute.live-video.net/app/{stream_key}"},
			{Name: "US East: Ashburn, VA", URLTemplate: "rtmp://iad05.contribute.live-video.net/app/{stream_key}"},
		},
	}

	server := resp.GetIngestServerByName("US East: Atlanta, GA")
	if server == nil {
		t.Fatal("expected to find server, got nil")
	}
	if server.Name != "US East: Atlanta, GA" {
		t.Errorf("expected name 'US East: Atlanta, GA', got %s", server.Name)
	}

	notFound := resp.GetIngestServerByName("Nonexistent Server")
	if notFound != nil {
		t.Errorf("expected nil for nonexistent server, got %v", notFound)
	}
}

func TestGetRTMPURL(t *testing.T) {
	server := &IngestServer{
		URLTemplate: "rtmp://atl.contribute.live-video.net/app/{stream_key}",
	}

	url := server.GetRTMPURL("live_123456_abcdef")
	expected := "rtmp://atl.contribute.live-video.net/app/live_123456_abcdef"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestReplaceStreamKey(t *testing.T) {
	tests := []struct {
		template  string
		streamKey string
		expected  string
	}{
		{
			template:  "rtmp://atl.contribute.live-video.net/app/{stream_key}",
			streamKey: "live_123456",
			expected:  "rtmp://atl.contribute.live-video.net/app/live_123456",
		},
		{
			template:  "rtmps://iad05.contribute.live-video.net/app/{stream_key}",
			streamKey: "test_key",
			expected:  "rtmps://iad05.contribute.live-video.net/app/test_key",
		},
		{
			template:  "rtmp://server/{stream_key}/live",
			streamKey: "mykey",
			expected:  "rtmp://server/mykey/live",
		},
		{
			template:  "no placeholder here",
			streamKey: "key",
			expected:  "no placeholder here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			result := replaceStreamKey(tt.template, tt.streamKey)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIngestServer_GetRTMPURL(t *testing.T) {
	server := &IngestServer{
		ID:          1,
		Name:        "US East: Atlanta, GA",
		URLTemplate: "rtmp://atl.contribute.live-video.net/app/{stream_key}",
	}

	url := server.GetRTMPURL("live_user_abc123")
	expected := "rtmp://atl.contribute.live-video.net/app/live_user_abc123"

	if url != expected {
		t.Errorf("expected URL %s, got %s", expected, url)
	}
}

func TestIngestServersResponse_GetIngestServerByName(t *testing.T) {
	resp := &IngestServersResponse{
		Ingests: []IngestServer{
			{ID: 1, Name: "US East: Atlanta, GA"},
			{ID: 2, Name: "US East: Ashburn, VA"},
			{ID: 3, Name: "EU West: Amsterdam, NL"},
		},
	}

	t.Run("found", func(t *testing.T) {
		server := resp.GetIngestServerByName("EU West: Amsterdam, NL")
		if server == nil {
			t.Fatal("expected server, got nil")
		}
		if server.ID != 3 {
			t.Errorf("expected ID 3, got %d", server.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := resp.GetIngestServerByName("Unknown Location")
		if server != nil {
			t.Errorf("expected nil, got %v", server)
		}
	})
}
