package helix

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// IngestBaseURL is the base URL for the Twitch Ingest API.
	IngestBaseURL = "https://ingest.twitch.tv"
)

// IngestServer represents a Twitch ingest server for live video streaming.
type IngestServer struct {
	ID           int     `json:"_id"`
	Availability float64 `json:"availability"`
	Default      bool    `json:"default"`
	Name         string  `json:"name"`
	URLTemplate  string  `json:"url_template"`
	Priority     int     `json:"priority"`
}

// IngestServersResponse represents the response from the ingest servers endpoint.
type IngestServersResponse struct {
	Ingests []IngestServer `json:"ingests"`
}

// GetIngestServers returns a list of endpoints for ingesting live video into Twitch.
// This endpoint does not require authentication.
// Note: This endpoint uses a different base URL (ingest.twitch.tv) than the Helix API.
func (c *Client) GetIngestServers(ctx context.Context) (*IngestServersResponse, error) {
	url := IngestBaseURL + "/ingests"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			ErrorType:  "ingest_error",
			Message:    string(body),
		}
	}

	var result IngestServersResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

// GetIngestServerByName finds an ingest server by its name (location).
// Returns nil if no server with the given name is found.
func (r *IngestServersResponse) GetIngestServerByName(name string) *IngestServer {
	for i := range r.Ingests {
		if r.Ingests[i].Name == name {
			return &r.Ingests[i]
		}
	}
	return nil
}

// GetRTMPURL returns the RTMP URL for this ingest server with the stream key inserted.
func (s *IngestServer) GetRTMPURL(streamKey string) string {
	// URL template looks like: rtmp://iad05.contribute.live-video.net/app/{stream_key}
	// Replace {stream_key} placeholder with actual key
	return replaceStreamKey(s.URLTemplate, streamKey)
}

func replaceStreamKey(template, streamKey string) string {
	result := template
	for i := 0; i < len(result); i++ {
		if i+12 <= len(result) && result[i:i+12] == "{stream_key}" {
			result = result[:i] + streamKey + result[i+12:]
			break
		}
	}
	return result
}
