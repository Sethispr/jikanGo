// genre.go
package jikan

import (
	"context"
	"net/http"
	"net/url"
)

type GenreService struct {
	c *Client
}

type Genre struct {
	MalID ID     `json:"mal_id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}

func (s *GenreService) Anime(ctx context.Context, filter string) ([]Genre, error) {
	q := url.Values{}
	if filter != "" {
		q.Set("filter", filter)
	}
	var r struct{ Data []Genre }
	if err := s.c.Do(ctx, http.MethodGet, "/genres/anime", q, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *GenreService) Manga(ctx context.Context, filter string) ([]Genre, error) {
	q := url.Values{}
	if filter != "" {
		q.Set("filter", filter)
	}
	var r struct{ Data []Genre }
	if err := s.c.Do(ctx, http.MethodGet, "/genres/manga", q, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
