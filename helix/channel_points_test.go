package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetCustomRewards(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channel_points/custom_rewards" {
			t.Errorf("expected /channel_points/custom_rewards, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		resp := Response[CustomReward]{
			Data: []CustomReward{
				{
					ID:                  "reward1",
					BroadcasterID:       "12345",
					BroadcasterLogin:    "streamer",
					BroadcasterName:     "Streamer",
					Title:               "VIP for a Day",
					Prompt:              "Become VIP for 24 hours!",
					Cost:                50000,
					IsEnabled:           true,
					IsUserInputRequired: false,
					BackgroundColor:     "#9147FF",
					IsPaused:            false,
					IsInStock:           true,
					MaxPerStreamSetting: MaxPerStream{
						IsEnabled:    true,
						MaxPerStream: 5,
					},
					MaxPerUserPerStreamSetting: MaxPerUserPerStream{
						IsEnabled:           true,
						MaxPerUserPerStream: 1,
					},
					GlobalCooldownSetting: GlobalCooldown{
						IsEnabled:             true,
						GlobalCooldownSeconds: 300,
					},
					ShouldRedemptionsSkipRequestQueue: false,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCustomRewards(context.Background(), &GetCustomRewardsParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 reward, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "VIP for a Day" {
		t.Errorf("expected 'VIP for a Day', got %s", resp.Data[0].Title)
	}
	if resp.Data[0].Cost != 50000 {
		t.Errorf("expected cost 50000, got %d", resp.Data[0].Cost)
	}
}

func TestClient_GetCustomRewards_ByIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[CustomReward]{
			Data: []CustomReward{
				{ID: "reward1", Title: "Reward 1"},
				{ID: "reward2", Title: "Reward 2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCustomRewards(context.Background(), &GetCustomRewardsParams{
		BroadcasterID: "12345",
		IDs:           []string{"reward1", "reward2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 rewards, got %d", len(resp.Data))
	}
}

func TestClient_GetCustomRewards_OnlyManageable(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		onlyManageable := r.URL.Query().Get("only_manageable_rewards")
		if onlyManageable != "true" {
			t.Errorf("expected only_manageable_rewards=true, got %s", onlyManageable)
		}

		resp := Response[CustomReward]{Data: []CustomReward{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetCustomRewards(context.Background(), &GetCustomRewardsParams{
		BroadcasterID:         "12345",
		OnlyManageableRewards: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_CreateCustomReward(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/channel_points/custom_rewards" {
			t.Errorf("expected /channel_points/custom_rewards, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		var params CreateCustomRewardParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if params.Title != "New Reward" {
			t.Errorf("expected title 'New Reward', got %s", params.Title)
		}
		if params.Cost != 1000 {
			t.Errorf("expected cost 1000, got %d", params.Cost)
		}

		resp := Response[CustomReward]{
			Data: []CustomReward{
				{
					ID:        "newreward",
					Title:     params.Title,
					Cost:      params.Cost,
					IsEnabled: true,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateCustomReward(context.Background(), &CreateCustomRewardParams{
		BroadcasterID: "12345",
		Title:         "New Reward",
		Cost:          1000,
		Prompt:        "Redeem this reward!",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "newreward" {
		t.Errorf("expected reward ID 'newreward', got %s", result.ID)
	}
}

func TestClient_UpdateCustomReward(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		rewardID := r.URL.Query().Get("id")

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if rewardID != "reward123" {
			t.Errorf("expected id=reward123, got %s", rewardID)
		}

		resp := Response[CustomReward]{
			Data: []CustomReward{
				{ID: "reward123", Title: "Updated Reward", Cost: 2000},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	newCost := 2000
	result, err := client.UpdateCustomReward(context.Background(), &UpdateCustomRewardParams{
		BroadcasterID: "12345",
		ID:            "reward123",
		Title:         "Updated Reward",
		Cost:          &newCost,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cost != 2000 {
		t.Errorf("expected cost 2000, got %d", result.Cost)
	}
}

func TestClient_DeleteCustomReward(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/channel_points/custom_rewards" {
			t.Errorf("expected /channel_points/custom_rewards, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		rewardID := r.URL.Query().Get("id")

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if rewardID != "reward123" {
			t.Errorf("expected id=reward123, got %s", rewardID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.DeleteCustomReward(context.Background(), "12345", "reward123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetCustomRewardRedemptions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channel_points/custom_rewards/redemptions" {
			t.Errorf("expected /channel_points/custom_rewards/redemptions, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		rewardID := r.URL.Query().Get("reward_id")
		status := r.URL.Query().Get("status")

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if rewardID != "reward123" {
			t.Errorf("expected reward_id=reward123, got %s", rewardID)
		}
		if status != "UNFULFILLED" {
			t.Errorf("expected status=UNFULFILLED, got %s", status)
		}

		resp := Response[CustomRewardRedemption]{
			Data: []CustomRewardRedemption{
				{
					ID:               "redemption1",
					BroadcasterID:    "12345",
					BroadcasterLogin: "streamer",
					BroadcasterName:  "Streamer",
					UserID:           "67890",
					UserLogin:        "viewer",
					UserName:         "Viewer",
					UserInput:        "Hello!",
					Status:           "UNFULFILLED",
					RedeemedAt:       time.Now(),
					Reward: struct {
						ID     string `json:"id"`
						Title  string `json:"title"`
						Prompt string `json:"prompt"`
						Cost   int    `json:"cost"`
					}{
						ID:    "reward123",
						Title: "Shoutout",
						Cost:  500,
					},
				},
			},
			Pagination: &Pagination{Cursor: "next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCustomRewardRedemptions(context.Background(), &GetCustomRewardRedemptionsParams{
		BroadcasterID: "12345",
		RewardID:      "reward123",
		Status:        "UNFULFILLED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 redemption, got %d", len(resp.Data))
	}
	if resp.Data[0].Status != "UNFULFILLED" {
		t.Errorf("expected status UNFULFILLED, got %s", resp.Data[0].Status)
	}
}

func TestClient_UpdateRedemptionStatus(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/channel_points/custom_rewards/redemptions" {
			t.Errorf("expected /channel_points/custom_rewards/redemptions, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		rewardID := r.URL.Query().Get("reward_id")
		ids := r.URL.Query()["id"]

		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}
		if rewardID != "reward123" {
			t.Errorf("expected reward_id=reward123, got %s", rewardID)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 redemption ids, got %d", len(ids))
		}

		var params UpdateRedemptionStatusParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Status != "FULFILLED" {
			t.Errorf("expected status FULFILLED, got %s", params.Status)
		}

		resp := Response[CustomRewardRedemption]{
			Data: []CustomRewardRedemption{
				{ID: "redemption1", Status: "FULFILLED"},
				{ID: "redemption2", Status: "FULFILLED"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.UpdateRedemptionStatus(context.Background(), &UpdateRedemptionStatusParams{
		BroadcasterID: "12345",
		RewardID:      "reward123",
		IDs:           []string{"redemption1", "redemption2"},
		Status:        "FULFILLED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 redemptions, got %d", len(resp.Data))
	}
	if resp.Data[0].Status != "FULFILLED" {
		t.Errorf("expected status FULFILLED, got %s", resp.Data[0].Status)
	}
}

func TestClient_UpdateRedemptionStatus_Cancel(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		var params UpdateRedemptionStatusParams
		_ = json.NewDecoder(r.Body).Decode(&params)
		if params.Status != "CANCELED" {
			t.Errorf("expected status CANCELED, got %s", params.Status)
		}

		resp := Response[CustomRewardRedemption]{
			Data: []CustomRewardRedemption{
				{ID: "redemption1", Status: "CANCELED"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.UpdateRedemptionStatus(context.Background(), &UpdateRedemptionStatusParams{
		BroadcasterID: "12345",
		RewardID:      "reward123",
		IDs:           []string{"redemption1"},
		Status:        "CANCELED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data[0].Status != "CANCELED" {
		t.Errorf("expected status CANCELED, got %s", resp.Data[0].Status)
	}
}

func TestClient_GetCustomRewards_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetCustomRewards(context.Background(), &GetCustomRewardsParams{BroadcasterID: "12345"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateCustomReward_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.CreateCustomReward(context.Background(), &CreateCustomRewardParams{BroadcasterID: "12345", Title: "Test", Cost: 100})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateCustomReward_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[CustomReward]{Data: []CustomReward{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.CreateCustomReward(context.Background(), &CreateCustomRewardParams{BroadcasterID: "12345", Title: "Test", Cost: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty response, got %v", result)
	}
}

func TestClient_UpdateCustomReward_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	_, err := client.UpdateCustomReward(context.Background(), &UpdateCustomRewardParams{BroadcasterID: "12345", ID: "reward123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateCustomReward_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[CustomReward]{Data: []CustomReward{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateCustomReward(context.Background(), &UpdateCustomRewardParams{BroadcasterID: "12345", ID: "reward123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty response, got %v", result)
	}
}

func TestClient_GetCustomRewardRedemptions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetCustomRewardRedemptions(context.Background(), &GetCustomRewardRedemptionsParams{BroadcasterID: "12345", RewardID: "reward123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateRedemptionStatus_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateRedemptionStatus(context.Background(), &UpdateRedemptionStatusParams{BroadcasterID: "12345", RewardID: "reward123", IDs: []string{"red1"}, Status: "FULFILLED"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetCustomRewardRedemptions_WithIDsAndSort(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		sort := r.URL.Query().Get("sort")
		if sort != "OLDEST" {
			t.Errorf("expected sort=OLDEST, got %s", sort)
		}

		resp := Response[CustomRewardRedemption]{
			Data: []CustomRewardRedemption{
				{ID: "redemption1", Status: "FULFILLED"},
				{ID: "redemption2", Status: "FULFILLED"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetCustomRewardRedemptions(context.Background(), &GetCustomRewardRedemptionsParams{
		BroadcasterID: "12345",
		RewardID:      "reward123",
		IDs:           []string{"redemption1", "redemption2"},
		Sort:          "OLDEST",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 redemptions, got %d", len(resp.Data))
	}
}
