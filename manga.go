package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type MangaService struct{ c *Client }

type Manga struct {
	MalID          ID         `json:"mal_id"`
	URL            string     `json:"url"`
	Images         ImageSet   `json:"images"`
	Title          string     `json:"title"`
	TitleEnglish   string     `json:"title_english"`
	Type           string     `json:"type"`
	Chapters       int        `json:"chapters"`
	Volumes        int        `json:"volumes"`
	Status         string     `json:"status"`
	Publishing     bool       `json:"publishing"`
	Published      DateRange  `json:"published"`
	Score          float64    `json:"score"`
	Rank           int        `json:"rank"`
	Authors        []Resource `json:"authors"`
	Serializations []Resource `json:"serializations"`
	Genres         []Resource `json:"genres"`
	Synopsis       string     `json:"synopsis"`
}

func (s *MangaService) ByID(ctx context.Context, id ID) (*Manga, error) {
	var r struct{ Data Manga }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/manga/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}
