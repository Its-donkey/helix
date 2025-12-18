package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetConduits(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/eventsub/conduits" {
			t.Errorf("expected /eventsub/conduits, got %s", r.URL.Path)
		}

		resp := Response[Conduit]{
			Data: []Conduit{
				{
					ID:         "conduit123",
					ShardCount: 5,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetConduits(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 conduit, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "conduit123" {
		t.Errorf("expected conduit ID 'conduit123', got %s", resp.Data[0].ID)
	}
}

func TestClient_CreateConduit(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/eventsub/conduits" {
			t.Errorf("expected /eventsub/conduits, got %s", r.URL.Path)
		}

		var body CreateConduitParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.ShardCount != 10 {
			t.Errorf("expected shard_count 10, got %d", body.ShardCount)
		}

		resp := Response[Conduit]{
			Data: []Conduit{
				{
					ID:         "newconduit",
					ShardCount: 10,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CreateConduit(context.Background(), 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected conduit, got nil")
	}
	if resp.ShardCount != 10 {
		t.Errorf("expected shard_count 10, got %d", resp.ShardCount)
	}
}

func TestClient_UpdateConduit(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		var body UpdateConduitParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.ID != "conduit123" {
			t.Errorf("expected ID 'conduit123', got %s", body.ID)
		}

		resp := Response[Conduit]{
			Data: []Conduit{
				{
					ID:         "conduit123",
					ShardCount: 15,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.UpdateConduit(context.Background(), &UpdateConduitParams{
		ID:         "conduit123",
		ShardCount: 15,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected conduit, got nil")
	}
}

func TestClient_DeleteConduit(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		conduitID := r.URL.Query().Get("id")
		if conduitID != "conduit123" {
			t.Errorf("expected id 'conduit123', got %s", conduitID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteConduit(context.Background(), "conduit123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetConduitShards(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/eventsub/conduits/shards" {
			t.Errorf("expected /eventsub/conduits/shards, got %s", r.URL.Path)
		}

		conduitID := r.URL.Query().Get("conduit_id")
		if conduitID != "conduit123" {
			t.Errorf("expected conduit_id 'conduit123', got %s", conduitID)
		}

		resp := GetConduitShardsResponse{
			Data: []ConduitShard{
				{
					ID:     "0",
					Status: "enabled",
					Transport: ConduitShardTransport{
						Method:   "webhook",
						Callback: "https://example.com/webhook",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetConduitShards(context.Background(), &GetConduitShardsParams{
		ConduitID: "conduit123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 shard, got %d", len(resp.Data))
	}
}

func TestClient_UpdateConduitShards(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		var body UpdateConduitShardsParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.ConduitID != "conduit123" {
			t.Errorf("expected conduit_id 'conduit123', got %s", body.ConduitID)
		}

		resp := UpdateConduitShardsResponse{
			Data: []ConduitShard{
				{
					ID:     "0",
					Status: "enabled",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.UpdateConduitShards(context.Background(), &UpdateConduitShardsParams{
		ConduitID: "conduit123",
		Shards: []UpdateConduitShardParams{
			{
				ID: "0",
				Transport: UpdateConduitShardTransport{
					Method:   "webhook",
					Callback: "https://example.com/webhook",
					Secret:   "secret123",
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 shard, got %d", len(resp.Data))
	}
}

func TestClient_GetConduits_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetConduits(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateConduit_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.CreateConduit(context.Background(), 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateConduit_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Conduit]{
			Data: []Conduit{}, // Empty response
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateConduit(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty response, got %v", result)
	}
}

func TestClient_UpdateConduit_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	_, err := client.UpdateConduit(context.Background(), &UpdateConduitParams{
		ID:         "conduit123",
		ShardCount: 15,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateConduit_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[Conduit]{
			Data: []Conduit{}, // Empty response
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateConduit(context.Background(), &UpdateConduitParams{
		ID:         "conduit123",
		ShardCount: 15,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty response, got %v", result)
	}
}

func TestClient_GetConduitShards_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetConduitShards(context.Background(), &GetConduitShardsParams{
		ConduitID: "conduit123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetConduitShards_WithStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		if status != "enabled" {
			t.Errorf("expected status 'enabled', got %s", status)
		}

		resp := GetConduitShardsResponse{
			Data: []ConduitShard{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetConduitShards(context.Background(), &GetConduitShardsParams{
		ConduitID: "conduit123",
		Status:    "enabled",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateConduitShards_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateConduitShards(context.Background(), &UpdateConduitShardsParams{
		ConduitID: "conduit123",
		Shards:    []UpdateConduitShardParams{},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
