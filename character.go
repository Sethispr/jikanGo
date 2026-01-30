package jikan

import (
	"context"
	"fmt"
	"net/http"
)

type CharacterService struct { c *Client }

type Character struct {
	MalID     ID       `json:"mal_id"`
	URL       string   `json:"url"`
	Images    ImageSet `json:"images"`
	Name      string   `json:"name"`
	NameKanji string   `json:"name_kanji"`
	Nicknames []string `json:"nicknames"`
	About     string   `json:"about"`
	Favorites int      `json:"favorites"`
	Anime     []Resource `json:"animeography"`
	Manga     []Resource `json:"mangaography"`
	Voices    []Resource `json:"voices"`
}

func (s *CharacterService) ByID(ctx context.Context, id ID) (*Character, error) {
	var r struct{ Data Character }
	if err := s.c.Do(ctx, http.MethodGet, fmt.Sprintf("/characters/%d", id), nil, &r); err != nil {
		return nil, err
	}
	return &r.Data, nil
}
