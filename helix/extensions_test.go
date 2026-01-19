package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClient_GetExtensionConfigurationSegment(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/extensions/configurations" {
			t.Errorf("expected /extensions/configurations, got %s", r.URL.Path)
		}

		resp := Response[ExtensionConfigurationSegment]{
			Data: []ExtensionConfigurationSegment{
				{
					Segment:       "broadcaster",
					BroadcasterID: "12345",
					Content:       `{"setting": "value"}`,
					Version:       "1.0",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionConfigurationSegment(context.Background(), &GetExtensionConfigurationSegmentParams{
		ExtensionID:   "ext123",
		Segment:       []string{"broadcaster"},
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(resp.Data))
	}
}

func TestClient_SetExtensionConfigurationSegment(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var body SetExtensionConfigurationSegmentParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.ExtensionID != "ext123" {
			t.Errorf("expected extension_id 'ext123', got %s", body.ExtensionID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SetExtensionConfigurationSegment(context.Background(), &SetExtensionConfigurationSegmentParams{
		ExtensionID:   "ext123",
		Segment:       "broadcaster",
		BroadcasterID: "12345",
		Content:       `{"setting": "new_value"}`,
		Version:       "1.1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SetExtensionRequiredConfiguration(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/extensions/required_configuration" {
			t.Errorf("expected /extensions/required_configuration, got %s", r.URL.Path)
		}

		// Verify broadcaster_id in query params
		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id=12345, got %s", broadcasterID)
		}

		// Verify body contains extension_id
		var body SetExtensionRequiredConfigurationParams
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.ExtensionID != "ext123" {
			t.Errorf("expected extension_id 'ext123', got %s", body.ExtensionID)
		}
		if body.ExtensionVersion != "1.0.0" {
			t.Errorf("expected extension_version '1.0.0', got %s", body.ExtensionVersion)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SetExtensionRequiredConfiguration(context.Background(), &SetExtensionRequiredConfigurationParams{
		BroadcasterID:         "12345",
		ExtensionID:           "ext123",
		ExtensionVersion:      "1.0.0",
		RequiredConfiguration: "config_version_1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendExtensionPubSubMessage(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/extensions/pubsub" {
			t.Errorf("expected /extensions/pubsub, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendExtensionPubSubMessage(context.Background(), &SendExtensionPubSubMessageParams{
		Target:        []string{"broadcast"},
		BroadcasterID: "12345",
		Message:       `{"data": "test"}`,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetExtensionLiveChannels(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extensions/live" {
			t.Errorf("expected /extensions/live, got %s", r.URL.Path)
		}

		resp := Response[ExtensionLiveChannel]{
			Data: []ExtensionLiveChannel{
				{
					BroadcasterID:   "12345",
					BroadcasterName: "TestUser",
					GameName:        "Test Game",
					GameID:          "game123",
					Title:           "Live Stream",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionLiveChannels(context.Background(), &GetExtensionLiveChannelsParams{
		ExtensionID: "ext123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(resp.Data))
	}
}

func TestClient_GetExtensionSecrets(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extensions/jwt/secrets" {
			t.Errorf("expected /extensions/jwt/secrets, got %s", r.URL.Path)
		}

		resp := Response[ExtensionSecret]{
			Data: []ExtensionSecret{
				{
					FormatVersion: 1,
					Secrets: []ExtensionSecretData{
						{
							Content:   "secret123",
							ActiveAt:  "2024-01-01T00:00:00Z",
							ExpiresAt: "2024-12-31T23:59:59Z",
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionSecrets(context.Background(), "ext123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 secret set, got %d", len(resp.Data))
	}
}

func TestClient_CreateExtensionSecret(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := Response[ExtensionSecret]{
			Data: []ExtensionSecret{
				{
					FormatVersion: 1,
					Secrets: []ExtensionSecretData{
						{Content: "newsecret"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CreateExtensionSecret(context.Background(), "ext123", 300)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 secret set, got %d", len(resp.Data))
	}
}

func TestClient_SendExtensionChatMessage(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/extensions/chat" {
			t.Errorf("expected /extensions/chat, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.SendExtensionChatMessage(context.Background(), &SendExtensionChatMessageParams{
		BroadcasterID:    "12345",
		Text:             "Hello from extension!",
		ExtensionID:      "ext123",
		ExtensionVersion: "1.0.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetExtensions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extensions" {
			t.Errorf("expected /extensions, got %s", r.URL.Path)
		}

		resp := Response[Extension]{
			Data: []Extension{
				{
					ID:                        "ext123",
					Name:                      "Test Extension",
					Version:                   "1.0.0",
					State:                     "Released",
					AuthorName:                "TestDev",
					CanInstall:                true,
					HasChatSupport:            true,
					ConfigurationLocation:     "hosted",
					Description:               "A test extension",
					EULAToSURL:                "https://example.com/eula",
					IconURL:                   "https://example.com/icon.png",
					PrivacyPolicyURL:          "https://example.com/privacy",
					RequestIdentityLink:       false,
					SubscriptionsSupportLevel: "optional",
					Summary:                   "Test summary",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensions(context.Background(), "ext123", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Test Extension" {
		t.Errorf("expected name 'Test Extension', got %s", resp.Data[0].Name)
	}
}

func TestClient_GetReleasedExtensions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extensions/released" {
			t.Errorf("expected /extensions/released, got %s", r.URL.Path)
		}

		resp := Response[Extension]{
			Data: []Extension{
				{
					ID:      "ext123",
					Name:    "Released Extension",
					Version: "2.0.0",
					State:   "Released",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetReleasedExtensions(context.Background(), "ext123", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(resp.Data))
	}
}

func TestClient_GetExtensionBitsProducts(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bits/extensions" {
			t.Errorf("expected /bits/extensions, got %s", r.URL.Path)
		}

		resp := Response[ExtensionBitsProduct]{
			Data: []ExtensionBitsProduct{
				{
					SKU:           "product123",
					Cost:          ExtensionBitsCost{Amount: 100, Type: "bits"},
					InDevelopment: false,
					DisplayName:   "Test Product",
					Expiration:    "",
					IsBroadcast:   true,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionBitsProducts(context.Background(), &GetExtensionBitsProductsParams{
		ShouldIncludeAll: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 product, got %d", len(resp.Data))
	}
}

func TestClient_UpdateExtensionBitsProduct(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		resp := Response[ExtensionBitsProduct]{
			Data: []ExtensionBitsProduct{
				{
					SKU:         "product123",
					DisplayName: "Updated Product",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.UpdateExtensionBitsProduct(context.Background(), &UpdateExtensionBitsProductParams{
		SKU:           "product123",
		Cost:          ExtensionBitsCost{Amount: 200, Type: "bits"},
		DisplayName:   "Updated Product",
		InDevelopment: false,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 product, got %d", len(resp.Data))
	}
}

func TestClient_GetExtensionTransactions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extensions/transactions" {
			t.Errorf("expected /extensions/transactions, got %s", r.URL.Path)
		}

		extensionID := r.URL.Query().Get("extension_id")
		if extensionID != "ext123" {
			t.Errorf("expected extension_id 'ext123', got %s", extensionID)
		}

		resp := Response[ExtensionTransaction]{
			Data: []ExtensionTransaction{
				{
					ID:               "tx123",
					Timestamp:        "2024-01-15T12:00:00Z",
					BroadcasterID:    "12345",
					BroadcasterLogin: "testuser",
					BroadcasterName:  "TestUser",
					UserID:           "67890",
					UserLogin:        "buyer",
					UserName:         "Buyer",
					ProductType:      "BITS_IN_EXTENSION",
					ProductData: ExtensionTransactionProduct{
						SKU:           "product123",
						Cost:          ExtensionBitsCost{Amount: 100, Type: "bits"},
						InDevelopment: false,
						DisplayName:   "Test Product",
					},
				},
			},
			Pagination: &Pagination{Cursor: "next-cursor"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionTransactions(context.Background(), &GetExtensionTransactionsParams{
		ExtensionID: "ext123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(resp.Data))
	}
	if resp.Data[0].ProductData.SKU != "product123" {
		t.Errorf("expected SKU 'product123', got %s", resp.Data[0].ProductData.SKU)
	}
}

func TestClient_GetExtensions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetExtensions(context.Background(), "ext123", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetReleasedExtensions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	_, err := client.GetReleasedExtensions(context.Background(), "ext123", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetExtensionConfigurationSegment_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetExtensionConfigurationSegment(context.Background(), &GetExtensionConfigurationSegmentParams{
		ExtensionID: "ext123",
		Segment:     []string{"broadcaster"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SetExtensionConfigurationSegment_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	err := client.SetExtensionConfigurationSegment(context.Background(), &SetExtensionConfigurationSegmentParams{
		ExtensionID: "ext123",
		Segment:     "broadcaster",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SetExtensionRequiredConfiguration_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	err := client.SetExtensionRequiredConfiguration(context.Background(), &SetExtensionRequiredConfigurationParams{
		BroadcasterID:    "12345",
		ExtensionID:      "ext123",
		ExtensionVersion: "1.0.0",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SendExtensionPubSubMessage_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	err := client.SendExtensionPubSubMessage(context.Background(), &SendExtensionPubSubMessageParams{
		Target:  []string{"broadcast"},
		Message: "test",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetExtensionLiveChannels_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer server.Close()

	_, err := client.GetExtensionLiveChannels(context.Background(), &GetExtensionLiveChannelsParams{
		ExtensionID: "ext123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetExtensionSecrets_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetExtensionSecrets(context.Background(), "ext123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_CreateExtensionSecret_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.CreateExtensionSecret(context.Background(), "ext123", 300)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SendExtensionChatMessage_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	err := client.SendExtensionChatMessage(context.Background(), &SendExtensionChatMessageParams{
		BroadcasterID:    "12345",
		Text:             "Hello!",
		ExtensionID:      "ext123",
		ExtensionVersion: "1.0.0",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetExtensionBitsProducts_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetExtensionBitsProducts(context.Background(), &GetExtensionBitsProductsParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateExtensionBitsProduct_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateExtensionBitsProduct(context.Background(), &UpdateExtensionBitsProductParams{
		SKU:         "product123",
		Cost:        ExtensionBitsCost{Amount: 100, Type: "bits"},
		DisplayName: "Test",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetExtensionTransactions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetExtensionTransactions(context.Background(), &GetExtensionTransactionsParams{
		ExtensionID: "ext123",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetExtensions_WithVersion(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		extensionVersion := r.URL.Query().Get("extension_version")
		if extensionVersion != "1.0.0" {
			t.Errorf("expected extension_version=1.0.0, got %s", extensionVersion)
		}

		resp := Response[Extension]{
			Data: []Extension{
				{ID: "ext123", Version: "1.0.0"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensions(context.Background(), "ext123", "1.0.0")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(resp.Data))
	}
}

func TestClient_GetReleasedExtensions_WithVersion(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		extensionVersion := r.URL.Query().Get("extension_version")
		if extensionVersion != "2.0.0" {
			t.Errorf("expected extension_version=2.0.0, got %s", extensionVersion)
		}

		resp := Response[Extension]{
			Data: []Extension{
				{ID: "ext123", Version: "2.0.0"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetReleasedExtensions(context.Background(), "ext123", "2.0.0")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(resp.Data))
	}
}

func TestClient_GetExtensionTransactions_WithIDs(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[ExtensionTransaction]{
			Data: []ExtensionTransaction{
				{ID: "tx1"},
				{ID: "tx2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetExtensionTransactions(context.Background(), &GetExtensionTransactionsParams{
		ExtensionID: "ext123",
		IDs:         []string{"tx1", "tx2"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(resp.Data))
	}
}
