package jikan

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type CharacterService struct {
	c *Client
}

type Character struct {
	MalID     ID       `json:"mal_id"`
	URL       string   `json:"url"`
	Images    ImageSet `json:"images"`
	Name      string   `json:"name"`
	NameKanji string   `json:"name_kanji"`
	Nicknames []string `json:"nicknames"`
	About     string   `json:"about"`
	Favorites int      `json:"favorites"`
}

type Entry struct {
	MalID  ID       `json:"mal_id"`
	URL    string   `json:"url"`
	Images ImageSet `json:"images"`
	Title  string   `json:"title"`
}

type CharacterAppearance struct {
	Role  string `json:"role"`
	Entry Entry  `json:"-"`
}

type CharacterAnime struct {
	Role  string `json:"role"`
	Anime Entry  `json:"anime"`
}

type CharacterManga struct {
	Role  string `json:"role"`
	Manga Entry  `json:"manga"`
}

type Person struct {
	MalID  ID       `json:"mal_id"`
	URL    string   `json:"url"`
	Images ImageSet `json:"images"`
	Name   string   `json:"name"`
}

type CharacterVoice struct {
	Language string `json:"language"`
	Person   Person `json:"person"`
}

type CharacterPicture struct {
	ImageURL      string `json:"image_url"`
	LargeImageURL string `json:"large_image_url"`
}

func (s *CharacterService) ByID(ctx context.Context, id ID) (*Character, error) {
	var r struct {
		Data Character `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}

func (s *CharacterService) Anime(ctx context.Context, id ID) ([]CharacterAnime, error) {
	var r struct {
		Data []CharacterAnime `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/anime", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Manga(ctx context.Context, id ID) ([]CharacterManga, error) {
	var r struct {
		Data []CharacterManga `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/manga", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Voices(ctx context.Context, id ID) ([]CharacterVoice, error) {
	var r struct {
		Data []CharacterVoice `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/voices", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

// Pictures retrieves all gallery pictures associated with the character.
// These are typically additional images beyond the main image in the Character struct.
func (s *CharacterService) Pictures(ctx context.Context, id ID) ([]CharacterPicture, error) {
	var r struct {
		Data []CharacterPicture `json:"data"`
	}
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d/pictures", id), nil, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

func (s *CharacterService) Search(ctx context.Context, query string, page int) ([]Character, *Pagination, error) {
	q := url.Values{}
	if query != "" {
		q.Set("q", query)
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}

	var r struct {
		Data       []Character `json:"data"`
		Pagination Pagination  `json:"pagination"`
	}
	if err := s.c.Do(ctx, http.MethodGet, "/characters", q, &r); err != nil {
		return nil, nil, err
	}
	return r.Data, &r.Pagination, nil
}
