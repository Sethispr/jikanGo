package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type AnimeService struct { c *Client }

type Anime struct {
	MalID          ID       `json:"mal_id"`
	URL            string   `json:"url"`
	Images         ImageSet `json:"images"`
	Title          string   `json:"title"`
	TitleEnglish   string   `json:"title_english"`
	TitleJapanese  string   `json:"title_japanese"`
	Type           string   `json:"type"`
	Source         string   `json:"source"`
	Episodes       int      `json:"episodes"`
	Status         string   `json:"status"`
	Airing         bool     `json:"airing"`
	Aired          DateRange `json:"aired"`
	Duration       string   `json:"duration"`
	Rating         string   `json:"rating"`
	Score          float64  `json:"score"`
	Rank           int      `json:"rank"`
	Popularity     int      `json:"popularity"`
	Members        int      `json:"members"`
	Favorites      int      `json:"favorites"`
	Synopsis       string   `json:"synopsis"`
	Season         string   `json:"season"`
	Year           int      `json:"year"`
	Studios        []Resource `json:"studios"`
	Genres         []Resource `json:"genres"`
	Themes         []Resource `json:"themes"`
	Demographics   []Resource `json:"demographics"`
}

func (s *AnimeService) ByID(ctx context.Context, id ID) (*Anime, error) {
	var r struct{ Data Anime }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *AnimeService) Characters(ctx context.Context, id ID) ([]struct {
	Character   Resource `json:"character"`
	Role        string   `json:"role"`
	VoiceActors []struct {
		Person   Resource `json:"person"`
		Language string   `json:"language"`
	} `json:"voice_actors"`
}, error) {
	var r struct{ Data []struct {
		Character   Resource `json:"character"`
		Role        string   `json:"role"`
		VoiceActors []struct {
			Person   Resource `json:"person"`
			Language string   `json:"language"`
		} `json:"voice_actors"`
	} }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/characters", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) Staff(ctx context.Context, id ID) ([]struct {
	Person    Resource `json:"person"`
	Positions []string `json:"positions"`
}, error) {
	var r struct{ Data []struct {
		Person    Resource `json:"person"`
		Positions []string `json:"positions"`
	} }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/staff", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *AnimeService) Episodes(ctx context.Context, id ID, page int) ([]struct {
	MalID  ID      `json:"mal_id"`
	Title  string  `json:"title"`
	Aired  string  `json:"aired"`
	Score  float64 `json:"score"`
	Filler bool    `json:"filler"`
	Recap  bool    `json:"recap"`
}, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data       []struct {
			MalID  ID      `json:"mal_id"`
			Title  string  `json:"title"`
			Aired  string  `json:"aired"`
			Score  float64 `json:"score"`
			Filler bool    `json:"filler"`
			Recap  bool    `json:"recap"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/episodes", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}

func (s *AnimeService) News(ctx context.Context, id ID, page int) ([]struct {
	MalID    ID     `json:"mal_id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Author   string `json:"author_username"`
	Comments int    `json:"comments"`
}, *Pagination, error) {
	q := url.Values{"page": {strconv.Itoa(page)}}
	var r struct {
		Data []struct {
			MalID    ID     `json:"mal_id"`
			Title    string `json:"title"`
			Date     string `json:"date"`
			Author   string `json:"author_username"`
			Comments int    `json:"comments"`
		} `json:"data"`
		Pagination Pagination `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/anime/%d/news", id), q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
