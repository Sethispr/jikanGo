package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Sethispr/jikanGo"
)

func main() {
	var (
		id    = flag.Int("id", 0, "Character ID (e.g., 1 for Spike Spiegel)")
		query = flag.String("search", "", "Search characters by name")
		page  = flag.Int("page", 1, "Page number for search results")
	)
	flag.Parse()

	if *id == 0 && *query == "" {
		fmt.Fprint(os.Stderr, "Enter Character ID or search query: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			log.Fatal("No input")
		}
		input := strings.TrimSpace(scanner.Text())
		if num, err := strconv.Atoi(input); err == nil {
			*id = num
		} else {
			*query = input
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := jikan.New(jikan.WithTimeout(10 * time.Second))

	if *query != "" {
		if err := searchCharacters(ctx, client, *query, *page); err != nil {
			log.Fatalf("Search failed: %v", err)
		}
		return
	}

	if err := showCharacterDetails(ctx, client, jikan.ID(*id)); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func searchCharacters(ctx context.Context, c *jikan.Client, query string, page int) error {
	characters, pagination, err := c.Character.Search(ctx, query, page)
	if err != nil {
		return err
	}

	if len(characters) == 0 {
		fmt.Println("No characters found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tKANJI\tFAVORITES")
	fmt.Fprintln(w, "-\t-\t-\t-")

	for _, char := range characters {
		kanji := char.NameKanji
		if kanji == "" {
			kanji = "-"
		}
		name := truncate(char.Name, 35)
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\n", char.MalID, name, kanji, char.Favorites)
	}
	w.Flush()

	if pagination.LastPage > 1 {
		fmt.Fprintf(os.Stderr, "\nPage %d of %d (Total: %d)\n",
			page, pagination.LastPage, pagination.Items.Total)
		fmt.Fprintf(os.Stderr, "Use -page flag to navigate pages\n")
	}

	return nil
}

func showCharacterDetails(ctx context.Context, c *jikan.Client, id jikan.ID) error {
	char, err := c.Character.ByID(ctx, id)
	if err != nil {
		var apiErr *jikan.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return fmt.Errorf("character %d not found", id)
		}
		return err
	}

	type results struct {
		anime    []jikan.CharacterAnime
		manga    []jikan.CharacterManga
		voices   []jikan.CharacterVoice
		pictures []jikan.CharacterPicture
	}

	resCh := make(chan results, 1)
	go func() {
		var r results
		r.anime, _ = c.Character.Anime(ctx, id)
		r.manga, _ = c.Character.Manga(ctx, id)
		r.voices, _ = c.Character.Voices(ctx, id)
		r.pictures, _ = c.Character.Pictures(ctx, id)
		resCh <- r
	}()

	res := <-resCh
	printCharacter(char, res.anime, res.manga, res.voices, res.pictures)
	return nil
}

func printCharacter(
	char *jikan.Character,
	anime []jikan.CharacterAnime,
	manga []jikan.CharacterManga,
	voices []jikan.CharacterVoice,
	pictures []jikan.CharacterPicture,
) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Name:\t%s\n", char.Name)
	if char.NameKanji != "" {
		fmt.Fprintf(w, "Kanji:\t%s\n", char.NameKanji)
	}
	if len(char.Nicknames) > 0 {
		fmt.Fprintf(w, "Nicknames:\t%s\n", strings.Join(char.Nicknames, ", "))
	}
	fmt.Fprintf(w, "ID:\t%d\n", char.MalID)
	fmt.Fprintf(w, "Favorites:\t%d\n", char.Favorites)
	fmt.Fprintf(w, "URL:\t%s\n", char.URL)
	if char.Images.JPG.Medium != "" {
		fmt.Fprintf(w, "Image:\t%s\n", char.Images.JPG.Medium)
	}
	w.Flush()

	if char.About != "" {
		about := truncate(char.About, 500)
		fmt.Printf("\nAbout:\n%s\n", about)
	}

	if len(anime) > 0 {
		fmt.Printf("\nAnime Appearances (%d):\n", len(anime))
		aw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(aw, "ROLE\tTITLE")
		for i, a := range anime {
			if i >= 10 {
				fmt.Fprintf(aw, "...\tand %d more\n", len(anime)-10)
				break
			}
			title := truncate(a.Anime.Title, 40)
			fmt.Fprintf(aw, "%s\t%s\n", a.Role, title)
		}
		aw.Flush()
	}

	if len(manga) > 0 {
		fmt.Printf("\nManga Appearances (%d):\n", len(manga))
		mw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(mw, "ROLE\tTITLE")
		for i, m := range manga {
			if i >= 10 {
				fmt.Fprintf(mw, "...\tand %d more\n", len(manga)-10)
				break
			}
			title := truncate(m.Manga.Title, 40)
			fmt.Fprintf(mw, "%s\t%s\n", m.Role, title)
		}
		mw.Flush()
	}

	if len(voices) > 0 {
		fmt.Printf("\nVoice Actors (%d):\n", len(voices))
		vw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(vw, "LANGUAGE\tNAME")
		for i, v := range voices {
			if i >= 10 {
				fmt.Fprintf(vw, "...\tand %d more\n", len(voices)-10)
				break
			}
			fmt.Fprintf(vw, "%s\t%s\n", v.Language, v.Person.Name)
		}
		vw.Flush()
	}

	if len(pictures) > 0 {
		fmt.Printf("\nGallery (%d images):\n", len(pictures))
		for i, p := range pictures {
			if i >= 5 {
				fmt.Printf("... and %d more images\n", len(pictures)-5)
				break
			}
			if p.LargeImageURL != "" {
				fmt.Printf("  %s\n", p.LargeImageURL)
			} else {
				fmt.Printf("  %s\n", p.ImageURL)
			}
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

