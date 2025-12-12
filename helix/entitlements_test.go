package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetDropsEntitlements(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/entitlements/drops" {
			t.Errorf("expected /entitlements/drops, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id '12345', got %s", userID)
		}

		resp := Response[DropsEntitlement]{
			Data: []DropsEntitlement{
				{
					ID:                "entitlement123",
					BenefitID:         "benefit456",
					Timestamp:         time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
					UserID:            "12345",
					GameID:            "game789",
					FulfillmentStatus: "CLAIMED",
					LastUpdated:       time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC),
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetDropsEntitlements(context.Background(), &GetDropsEntitlementsParams{
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 entitlement, got %d", len(resp.Data))
	}
	if resp.Data[0].FulfillmentStatus != "CLAIMED" {
		t.Errorf("expected fulfillment_status 'CLAIMED', got %s", resp.Data[0].FulfillmentStatus)
	}
}

func TestClient_GetDropsEntitlements_ByID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id != "entitlement1" {
			t.Errorf("expected id 'entitlement1', got %s", id)
		}

		resp := Response[DropsEntitlement]{
			Data: []DropsEntitlement{
				{ID: "entitlement1"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetDropsEntitlements(context.Background(), &GetDropsEntitlementsParams{
		ID: "entitlement1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 entitlement, got %d", len(resp.Data))
	}
}

func TestClient_UpdateDropsEntitlements(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/entitlements/drops" {
			t.Errorf("expected /entitlements/drops, got %s", r.URL.Path)
		}

		var body UpdateDropsEntitlementsParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if len(body.EntitlementIDs) != 2 {
			t.Errorf("expected 2 entitlement_ids, got %d", len(body.EntitlementIDs))
		}
		if body.FulfillmentStatus != "FULFILLED" {
			t.Errorf("expected fulfillment_status 'FULFILLED', got %s", body.FulfillmentStatus)
		}

		resp := Response[UpdateDropsEntitlementsResponse]{
			Data: []UpdateDropsEntitlementsResponse{
				{
					Status: "SUCCESS",
					IDs:    []string{"entitlement1", "entitlement2"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.UpdateDropsEntitlements(context.Background(), &UpdateDropsEntitlementsParams{
		EntitlementIDs:    []string{"entitlement1", "entitlement2"},
		FulfillmentStatus: "FULFILLED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 status, got %d", len(resp))
	}
	if resp[0].Status != "SUCCESS" {
		t.Errorf("expected status 'SUCCESS', got %s", resp[0].Status)
	}
}
