package helix

import (
	"context"
	"net/url"
)

// GuestStarSettings represents channel Guest Star settings.
type GuestStarSettings struct {
	IsModeratorSendLiveEnabled  bool   `json:"is_moderator_send_live_enabled"`
	SlotCount                   int    `json:"slot_count"`
	IsBrowserSourceAudioEnabled bool   `json:"is_browser_source_audio_enabled"`
	GroupLayout                 string `json:"group_layout"` // TILED_LAYOUT, SCREENSHARE_LAYOUT, HORIZONTAL_LAYOUT, VERTICAL_LAYOUT
	BrowserSourceToken          string `json:"browser_source_token,omitempty"`
}

// GetChannelGuestStarSettings gets the Guest Star settings for a channel.
// Requires: channel:read:guest_star, channel:manage:guest_star, or moderator:read:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) GetChannelGuestStarSettings(ctx context.Context, broadcasterID, moderatorID string) (*GuestStarSettings, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)

	var resp Response[GuestStarSettings]
	if err := c.get(ctx, "/channels/guest_star_settings", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// UpdateChannelGuestStarSettingsParams contains parameters for UpdateChannelGuestStarSettings.
type UpdateChannelGuestStarSettingsParams struct {
	BroadcasterID               string `json:"-"`
	IsModeratorSendLiveEnabled  *bool  `json:"is_moderator_send_live_enabled,omitempty"`
	SlotCount                   *int   `json:"slot_count,omitempty"`
	IsBrowserSourceAudioEnabled *bool  `json:"is_browser_source_audio_enabled,omitempty"`
	GroupLayout                 string `json:"group_layout,omitempty"`
	RegenerateBrowserSources    *bool  `json:"regenerate_browser_sources,omitempty"`
}

// UpdateChannelGuestStarSettings updates the Guest Star settings for a channel.
// Requires: channel:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) UpdateChannelGuestStarSettings(ctx context.Context, params *UpdateChannelGuestStarSettingsParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)

	return c.put(ctx, "/channels/guest_star_settings", q, params, nil)
}

// GuestStarSession represents a Guest Star session.
type GuestStarSession struct {
	ID     string           `json:"id"`
	Guests []GuestStarGuest `json:"guests"`
}

// GuestStarGuest represents a guest in a Guest Star session.
type GuestStarGuest struct {
	SlotID          string                 `json:"slot_id"`
	IsLive          bool                   `json:"is_live"`
	UserID          string                 `json:"user_id"`
	UserDisplayName string                 `json:"user_display_name"`
	UserLogin       string                 `json:"user_login"`
	Volume          int                    `json:"volume"`
	AssignedAt      string                 `json:"assigned_at"`
	AudioSettings   GuestStarAudioSettings `json:"audio_settings"`
	VideoSettings   GuestStarVideoSettings `json:"video_settings"`
}

// GuestStarAudioSettings represents audio settings for a guest.
type GuestStarAudioSettings struct {
	IsHostEnabled bool `json:"is_host_enabled"`
	IsSelfMuted   bool `json:"is_self_muted"`
	IsAvailable   bool `json:"is_available"`
}

// GuestStarVideoSettings represents video settings for a guest.
type GuestStarVideoSettings struct {
	IsHostEnabled bool `json:"is_host_enabled"`
	IsSelfMuted   bool `json:"is_self_muted"`
	IsAvailable   bool `json:"is_available"`
}

// GetGuestStarSession gets the active Guest Star session for a channel.
// Requires: channel:read:guest_star, channel:manage:guest_star, or moderator:read:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) GetGuestStarSession(ctx context.Context, broadcasterID, moderatorID string) (*GuestStarSession, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)

	var resp Response[GuestStarSession]
	if err := c.get(ctx, "/guest_star/session", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// CreateGuestStarSession creates a Guest Star session.
// Requires: channel:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) CreateGuestStarSession(ctx context.Context, broadcasterID string) (*GuestStarSession, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)

	var resp Response[GuestStarSession]
	if err := c.post(ctx, "/guest_star/session", q, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// EndGuestStarSession ends a Guest Star session.
// Requires: channel:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) EndGuestStarSession(ctx context.Context, broadcasterID, sessionID string) (*GuestStarSession, error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("session_id", sessionID)

	var resp Response[GuestStarSession]
	if err := c.delete(ctx, "/guest_star/session", q, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return &resp.Data[0], nil
}

// GuestStarInvite represents a Guest Star invite.
type GuestStarInvite struct {
	UserID           string `json:"user_id"`
	InvitedAt        string `json:"invited_at"`
	Status           string `json:"status"` // INVITED, ACCEPTED, READY
	IsVideoEnabled   bool   `json:"is_video_enabled"`
	IsAudioEnabled   bool   `json:"is_audio_enabled"`
	IsVideoAvailable bool   `json:"is_video_available"`
	IsAudioAvailable bool   `json:"is_audio_available"`
}

// GetGuestStarInvites gets the pending invites for a Guest Star session.
// Requires: channel:read:guest_star, channel:manage:guest_star, or moderator:read:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) GetGuestStarInvites(ctx context.Context, broadcasterID, moderatorID, sessionID string) (*Response[GuestStarInvite], error) {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("session_id", sessionID)

	var resp Response[GuestStarInvite]
	if err := c.get(ctx, "/guest_star/invites", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendGuestStarInvite sends an invite to a Guest Star session.
// Requires: channel:manage:guest_star or moderator:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) SendGuestStarInvite(ctx context.Context, broadcasterID, moderatorID, sessionID, guestID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("session_id", sessionID)
	q.Set("guest_id", guestID)

	return c.post(ctx, "/guest_star/invites", q, nil, nil)
}

// DeleteGuestStarInvite deletes an invite to a Guest Star session.
// Requires: channel:manage:guest_star or moderator:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) DeleteGuestStarInvite(ctx context.Context, broadcasterID, moderatorID, sessionID, guestID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("session_id", sessionID)
	q.Set("guest_id", guestID)

	return c.delete(ctx, "/guest_star/invites", q, nil)
}

// AssignGuestStarSlot assigns a guest to a slot in a Guest Star session.
// Requires: channel:manage:guest_star or moderator:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) AssignGuestStarSlot(ctx context.Context, broadcasterID, moderatorID, sessionID, guestID, slotID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("session_id", sessionID)
	q.Set("guest_id", guestID)
	q.Set("slot_id", slotID)

	return c.post(ctx, "/guest_star/slot", q, nil, nil)
}

// UpdateGuestStarSlotParams contains parameters for UpdateGuestStarSlot.
type UpdateGuestStarSlotParams struct {
	BroadcasterID     string `json:"-"`
	ModeratorID       string `json:"-"`
	SessionID         string `json:"-"`
	SourceSlotID      string `json:"-"`
	DestinationSlotID string `json:"-"`
}

// UpdateGuestStarSlot moves a guest from one slot to another.
// Requires: channel:manage:guest_star or moderator:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) UpdateGuestStarSlot(ctx context.Context, params *UpdateGuestStarSlotParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	q.Set("session_id", params.SessionID)
	q.Set("source_slot_id", params.SourceSlotID)
	if params.DestinationSlotID != "" {
		q.Set("destination_slot_id", params.DestinationSlotID)
	}

	return c.patch(ctx, "/guest_star/slot", q, nil, nil)
}

// DeleteGuestStarSlot removes a guest from a slot.
// Requires: channel:manage:guest_star or moderator:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) DeleteGuestStarSlot(ctx context.Context, broadcasterID, moderatorID, sessionID, guestID, slotID string) error {
	q := url.Values{}
	q.Set("broadcaster_id", broadcasterID)
	q.Set("moderator_id", moderatorID)
	q.Set("session_id", sessionID)
	q.Set("guest_id", guestID)
	q.Set("slot_id", slotID)

	return c.delete(ctx, "/guest_star/slot", q, nil)
}

// UpdateGuestStarSlotSettingsParams contains parameters for UpdateGuestStarSlotSettings.
type UpdateGuestStarSlotSettingsParams struct {
	BroadcasterID  string `json:"-"`
	ModeratorID    string `json:"-"`
	SessionID      string `json:"-"`
	SlotID         string `json:"-"`
	IsAudioEnabled *bool  `json:"is_audio_enabled,omitempty"`
	IsVideoEnabled *bool  `json:"is_video_enabled,omitempty"`
	IsLive         *bool  `json:"is_live,omitempty"`
	Volume         *int   `json:"volume,omitempty"`
}

// UpdateGuestStarSlotSettings updates the settings for a Guest Star slot.
// Requires: channel:manage:guest_star or moderator:manage:guest_star scope.
// Note: This is a BETA endpoint.
func (c *Client) UpdateGuestStarSlotSettings(ctx context.Context, params *UpdateGuestStarSlotSettingsParams) error {
	q := url.Values{}
	q.Set("broadcaster_id", params.BroadcasterID)
	q.Set("moderator_id", params.ModeratorID)
	q.Set("session_id", params.SessionID)
	q.Set("slot_id", params.SlotID)

	return c.patch(ctx, "/guest_star/slot_settings", q, params, nil)
}
