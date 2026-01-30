// club.go
package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ClubService struct {
	c *Client
}

type Club struct {
	MalID    int    `json:"mal_id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Members  int    `json:"members"`
	Category string `json:"category"`
	Access   string `json:"access"`
	Created  string `json:"created"`
}

func (s *ClubService) ByID(ctx context.Context, id ID) (*Club, error) {
	var r struct{ Data Club }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/clubs/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *ClubService) Search(ctx context.Context, query string, page int) ([]Club, *Pagination, error) {
	q := url.Values{"q": {query}, "page": {strconv.Itoa(page)}}
	var r struct {
		Data       []Club     `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/clubs", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *ClubService) Members(ctx context.Context, id ID, page int) ([]struct {
	Username string `json:"username"`
	Images   struct {
		JPG struct{ ImageURL string `json:"image_url"` } `json:"jpg"`
	} `json:"images"`
}, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			Username string `json:"username"`
			Images   struct {
				JPG struct{ ImageURL string `json:"image_url"` } `json:"jpg"`
			} `json:"images"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/clubs/%d/members", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
