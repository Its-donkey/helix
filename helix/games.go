package helix

import (
	"context"
	"net/url"
)

// Game represents a Twitch game/category.
type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
	IGDBId    string `json:"igdb_id,omitempty"`
}

// GetGamesParams contains parameters for GetGames.
type GetGamesParams struct {
	IDs     []string // Game IDs (max 100)
	Names   []string // Game names (max 100)
	IGDBIDs []string // IGDB IDs (max 100)
}

// GetGames gets information about one or more games.
func (c *Client) GetGames(ctx context.Context, params *GetGamesParams) (*Response[Game], error) {
	q := url.Values{}
	if params != nil {
		for _, id := range params.IDs {
			q.Add("id", id)
		}
		for _, name := range params.Names {
			q.Add("name", name)
		}
		for _, igdbID := range params.IGDBIDs {
			q.Add("igdb_id", igdbID)
		}
	}

	var resp Response[Game]
	if err := c.get(ctx, "/games", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTopGamesParams contains parameters for GetTopGames.
type GetTopGamesParams struct {
	*PaginationParams
}

// GetTopGames gets the top games/categories sorted by number of current viewers.
func (c *Client) GetTopGames(ctx context.Context, params *GetTopGamesParams) (*Response[Game], error) {
	q := url.Values{}
	if params != nil {
		addPaginationParams(q, params.PaginationParams)
	}

	var resp Response[Game]
	if err := c.get(ctx, "/games/top", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
