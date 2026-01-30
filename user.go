// user.go
package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type UserService struct {
	c *Client
}

type User struct {
	Username string `json:"username"`
	URL      string `json:"url"`
	Images   struct {
		JPG struct{ ImageURL string `json:"image_url"` } `json:"jpg"`
	} `json:"images"`
	LastOnline string     `json:"last_online"`
	Gender     string     `json:"gender"`
	Joined     string     `json:"joined"`
	Anime      Statistics `json:"anime_statistics"`
	Manga      Statistics `json:"manga_statistics"`
}

func (s *UserService) ByID(ctx context.Context, username string) (*User, error) {
	var r struct{ Data User }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/full", username), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *UserService) History(ctx context.Context, username string, page int) ([]struct {
	Entry     Resource `json:"entry"`
	Increment int      `json:"increment"`
	Date      string   `json:"date"`
}, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			Entry     Resource `json:"entry"`
			Increment int      `json:"increment"`
			Date      string   `json:"date"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/users/%s/history", username), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
