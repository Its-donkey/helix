package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetUsers_ByID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("expected /users, got %s", r.URL.Path)
		}

		ids := r.URL.Query()["id"]
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}

		resp := Response[User]{
			Data: []User{
				{ID: "12345", Login: "user1", DisplayName: "User1"},
				{ID: "67890", Login: "user2", DisplayName: "User2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUsers(context.Background(), &GetUsersParams{
		IDs: []string{"12345", "67890"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 users, got %d", len(resp.Data))
	}
}

func TestClient_GetUsers_ByLogin(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		logins := r.URL.Query()["login"]
		if len(logins) != 1 || logins[0] != "testuser" {
			t.Errorf("expected login=testuser, got %v", logins)
		}

		resp := Response[User]{
			Data: []User{
				{
					ID:              "12345",
					Login:           "testuser",
					DisplayName:     "TestUser",
					BroadcasterType: "partner",
					Description:     "Test description",
					CreatedAt:       time.Now(),
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUsers(context.Background(), &GetUsersParams{
		Logins: []string{"testuser"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Data))
	}
	if resp.Data[0].BroadcasterType != "partner" {
		t.Errorf("expected broadcaster type partner, got %s", resp.Data[0].BroadcasterType)
	}
}

func TestClient_GetUsers_NoParams(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		// Should return authenticated user when no params
		resp := Response[User]{
			Data: []User{
				{ID: "99999", Login: "authuser", DisplayName: "AuthUser"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUsers(context.Background(), nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Data))
	}
}

func TestClient_GetCurrentUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{
			Data: []User{
				{
					ID:          "12345",
					Login:       "currentuser",
					DisplayName: "CurrentUser",
					Email:       "user@example.com",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	user, err := client.GetCurrentUser(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user to not be nil")
	}
	if user.Login != "currentuser" {
		t.Errorf("expected login currentuser, got %s", user.Login)
	}
}

func TestClient_GetCurrentUser_NoUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	user, err := client.GetCurrentUser(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected user to be nil")
	}
}

func TestClient_UpdateUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		desc := r.URL.Query().Get("description")
		if desc != "New description" {
			t.Errorf("expected description 'New description', got %s", desc)
		}

		resp := Response[User]{
			Data: []User{
				{ID: "12345", Description: "New description"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	user, err := client.UpdateUser(context.Background(), &UpdateUserParams{
		Description: "New description",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Description != "New description" {
		t.Errorf("expected description 'New description', got %s", user.Description)
	}
}

func TestClient_GetUserBlockList(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/blocks" {
			t.Errorf("expected /users/blocks, got %s", r.URL.Path)
		}

		broadcasterID := r.URL.Query().Get("broadcaster_id")
		if broadcasterID != "12345" {
			t.Errorf("expected broadcaster_id 12345, got %s", broadcasterID)
		}

		resp := Response[BlockedUser]{
			Data: []BlockedUser{
				{UserID: "11111", UserLogin: "blocked1", DisplayName: "Blocked1"},
				{UserID: "22222", UserLogin: "blocked2", DisplayName: "Blocked2"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUserBlockList(context.Background(), &GetUserBlockListParams{
		BroadcasterID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 blocked users, got %d", len(resp.Data))
	}
}

func TestClient_BlockUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/users/blocks" {
			t.Errorf("expected /users/blocks, got %s", r.URL.Path)
		}

		targetUserID := r.URL.Query().Get("target_user_id")
		if targetUserID != "67890" {
			t.Errorf("expected target_user_id 67890, got %s", targetUserID)
		}

		sourceContext := r.URL.Query().Get("source_context")
		if sourceContext != "chat" {
			t.Errorf("expected source_context chat, got %s", sourceContext)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.BlockUser(context.Background(), &BlockUserParams{
		TargetUserID:  "67890",
		SourceContext: "chat",
		Reason:        "spam",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UnblockUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/users/blocks" {
			t.Errorf("expected /users/blocks, got %s", r.URL.Path)
		}

		targetUserID := r.URL.Query().Get("target_user_id")
		if targetUserID != "67890" {
			t.Errorf("expected target_user_id 67890, got %s", targetUserID)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.UnblockUser(context.Background(), "67890")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_GetUserExtensions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/extensions/list" {
			t.Errorf("expected /users/extensions/list, got %s", r.URL.Path)
		}

		resp := Response[UserExtension]{
			Data: []UserExtension{
				{
					ID:          "ext123",
					Version:     "1.0.0",
					Name:        "Test Extension",
					CanActivate: true,
					Type:        []string{"panel", "overlay"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetUserExtensions(context.Background())

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

func TestClient_GetUserActiveExtensions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/extensions" {
			t.Errorf("expected /users/extensions, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := struct {
			Data UserActiveExtensions `json:"data"`
		}{
			Data: UserActiveExtensions{
				Panel: map[string]ActiveExtension{
					"1": {Active: true, ID: "ext1", Version: "1.0.0", Name: "Panel Ext"},
					"2": {Active: false},
				},
				Overlay: map[string]ActiveExtension{
					"1": {Active: true, ID: "ext2", Version: "2.0.0", Name: "Overlay Ext"},
				},
				Component: map[string]ActiveExtension{
					"1": {Active: true, ID: "ext3", Version: "1.5.0", Name: "Component Ext", X: 100, Y: 200},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.GetUserActiveExtensions(context.Background(), "12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected active extensions, got nil")
	}
	if len(result.Panel) != 2 {
		t.Errorf("expected 2 panel slots, got %d", len(result.Panel))
	}
	if !result.Panel["1"].Active {
		t.Error("expected panel slot 1 to be active")
	}
	if result.Component["1"].X != 100 {
		t.Errorf("expected component X=100, got %d", result.Component["1"].X)
	}
}

func TestClient_GetUserActiveExtensions_NoUserID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID != "" {
			t.Errorf("expected no user_id param, got %s", userID)
		}

		resp := struct {
			Data UserActiveExtensions `json:"data"`
		}{
			Data: UserActiveExtensions{
				Panel:     map[string]ActiveExtension{},
				Overlay:   map[string]ActiveExtension{},
				Component: map[string]ActiveExtension{},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	_, err := client.GetUserActiveExtensions(context.Background(), "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_UpdateUserExtensions(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/users/extensions" {
			t.Errorf("expected /users/extensions, got %s", r.URL.Path)
		}

		var body UpdateUserExtensionsParams
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body.Data.Panel["1"].ID != "ext123" {
			t.Errorf("expected panel slot 1 id=ext123, got %s", body.Data.Panel["1"].ID)
		}

		resp := struct {
			Data UserActiveExtensions `json:"data"`
		}{
			Data: UserActiveExtensions{
				Panel: map[string]ActiveExtension{
					"1": {Active: true, ID: "ext123", Version: "1.0.0", Name: "Updated Ext"},
				},
				Overlay:   map[string]ActiveExtension{},
				Component: map[string]ActiveExtension{},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateUserExtensions(context.Background(), &UpdateUserExtensionsParams{
		Data: UserActiveExtensions{
			Panel: map[string]ActiveExtension{
				"1": {Active: true, ID: "ext123", Version: "1.0.0"},
			},
			Overlay:   map[string]ActiveExtension{},
			Component: map[string]ActiveExtension{},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Panel["1"].Name != "Updated Ext" {
		t.Errorf("expected panel slot 1 name='Updated Ext', got %s", result.Panel["1"].Name)
	}
}

func TestClient_GetAuthorizationByUser(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/users/authorization" {
			t.Errorf("expected /users/authorization, got %s", r.URL.Path)
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "12345" {
			t.Errorf("expected user_id=12345, got %s", userID)
		}

		resp := Response[UserAuthorization]{
			Data: []UserAuthorization{
				{
					ClientID: "client123",
					UserID:   "12345",
					Login:    "testuser",
					Scopes:   []string{"user:read:email", "bits:read", "channel:read:subscriptions"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.GetAuthorizationByUser(context.Background(), &GetAuthorizationByUserParams{
		UserID: "12345",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 authorization, got %d", len(resp.Data))
	}
	if resp.Data[0].ClientID != "client123" {
		t.Errorf("expected client_id=client123, got %s", resp.Data[0].ClientID)
	}
	if resp.Data[0].Login != "testuser" {
		t.Errorf("expected login=testuser, got %s", resp.Data[0].Login)
	}
	if len(resp.Data[0].Scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(resp.Data[0].Scopes))
	}
}

func TestClient_GetAuthorizationByUser_MissingUserID(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make request when user_id is missing")
	})
	defer server.Close()

	_, err := client.GetAuthorizationByUser(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil params")
	}

	_, err = client.GetAuthorizationByUser(context.Background(), &GetAuthorizationByUserParams{})
	if err == nil {
		t.Error("expected error for empty user_id")
	}
}

func TestClient_GetCurrentUser_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer server.Close()

	_, err := client.GetCurrentUser(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateUser_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateUser(context.Background(), &UpdateUserParams{Description: "test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateUser_EmptyResponse(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[User]{Data: []User{}}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.UpdateUser(context.Background(), &UpdateUserParams{Description: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty response, got %v", result)
	}
}

func TestClient_GetUserBlockList_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetUserBlockList(context.Background(), &GetUserBlockListParams{BroadcasterID: "12345"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetUserExtensions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetUserExtensions(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetUserActiveExtensions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	})
	defer server.Close()

	_, err := client.GetUserActiveExtensions(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_UpdateUserExtensions_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	})
	defer server.Close()

	_, err := client.UpdateUserExtensions(context.Background(), &UpdateUserExtensionsParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetAuthorizationByUser_Error(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	})
	defer server.Close()

	_, err := client.GetAuthorizationByUser(context.Background(), &GetAuthorizationByUserParams{UserID: "12345"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
