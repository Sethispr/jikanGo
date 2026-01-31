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
		id    = flag.Int("id", 0, "Anime ID")
		query = flag.String("search", "", "Search with Name")
	)
	flag.Parse()

	if *id == 0 && *query == "" {
		fmt.Print("Enter Anime ID or search with name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			log.Fatal("Give me an input..")
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
		searchAndShow(ctx, client, *query)
	} else {
		showByID(ctx, client, jikan.ID(*id))
	}
}

func showByID(ctx context.Context, c *jikan.Client, id jikan.ID) {
	anime, err := c.Anime.ByID(ctx, id)
	if err != nil {
		var apiErr *jikan.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			log.Fatalf("Anime %d not found..", id)
		}
		log.Fatal(err)
	}

	// Get extras
	stats, _ := c.Anime.Statistics(ctx, id)
	rels, _ := c.Anime.Relations(ctx, id)
	themes, _ := c.Anime.Themes(ctx, id)
	ext, _ := c.Anime.External(ctx, id)

	printAnime(anime, stats, rels, themes, ext)
}

func searchAndShow(ctx context.Context, c *jikan.Client, query string) {
	results, _, err := c.Search.Anime(ctx, query, struct {
		Type, Status, Rating string
		Genres               []int
		OrderBy, Sort        string
		Page, Limit          int
	}{Limit: 5})
	if err != nil {
		log.Fatal(err)
	}

	for i, a := range results {
		fmt.Printf("\nResult %d/%d \n", i+1, len(results))

		// Get details
		stats, _ := c.Anime.Statistics(ctx, a.MalID)
		rels, _ := c.Anime.Relations(ctx, a.MalID)
		themes, _ := c.Anime.Themes(ctx, a.MalID)
		ext, _ := c.Anime.External(ctx, a.MalID)

		printAnime(&a, stats, rels, themes, ext)

		if i < len(results)-1 {
			fmt.Println("\nPress Enter..")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}
}

func printAnime(a *jikan.Anime, stats *struct {
	Watching    int `json:"watching"`
	Completed   int `json:"completed"`
	OnHold      int `json:"on_hold"`
	Dropped     int `json:"dropped"`
	PlanToWatch int `json:"plan_to_watch"`
	Total       int `json:"total"`
	Scores      []struct {
		Score      int     `json:"score"`
		Votes      int     `json:"votes"`
		Percentage float64 `json:"percentage"`
	} `json:"scores"`
}, rels []struct {
	Relation string           `json:"relation"`
	Entry    []jikan.Resource `json:"entry"`
}, themes *struct {
	Openings []string `json:"openings"`
	Endings  []string `json:"endings"`
}, ext []struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}) {

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Title:\t%s\n", a.Title)
	if a.TitleEnglish != "" {
		fmt.Fprintf(w, "English:\t%s\n", a.TitleEnglish)
	}
	fmt.Fprintf(w, "Type:\t%s\n", a.Type)
	fmt.Fprintf(w, "Episodes:\t%d\n", a.Episodes)
	fmt.Fprintf(w, "Status:\t%s\n", a.Status)
	fmt.Fprintf(w, "Score:\t%.2f\n", a.Score)
	fmt.Fprintf(w, "Rank:\t#%d\n", a.Rank)
	fmt.Fprintf(w, "Popularity:\t#%d\n", a.Popularity)
	if len(a.Studios) > 0 {
		fmt.Fprintf(w, "Studios:\t%s\n", join(a.Studios))
	}
	if len(a.Genres) > 0 {
		fmt.Fprintf(w, "Genres:\t%s\n", join(a.Genres))
	}
	w.Flush()

	if stats != nil {
		fmt.Println("\nStats")
		fmt.Printf("Watching: %d | Completed: %d | Dropped: %d\n",
			stats.Watching, stats.Completed, stats.Dropped)
	}

	if len(rels) > 0 {
		fmt.Println("\nRelations")
		for _, r := range rels {
			names := []string{}
			for _, e := range r.Entry {
				names = append(names, e.Name)
			}
			fmt.Printf("%s: %s\n", r.Relation, strings.Join(names, ", "))
		}
	}

	if themes != nil && len(themes.Openings) > 0 {
		fmt.Printf("\nOpenings: %d \n", len(themes.Openings))
		for i, op := range themes.Openings {
			if i >= 2 {
				fmt.Printf("... and %d more\n", len(themes.Openings)-2)
				break
			}
			fmt.Printf("  %s\n", op)
		}
	}

	if len(ext) > 0 {
		fmt.Println("\nExternal")
		for _, e := range ext {
			fmt.Printf("  %s\n", e.URL)
		}
	}

	if a.Synopsis != "" {
		fmt.Printf("\nSynopsis:\n%.300s...\n", a.Synopsis)
	}

	if a.Images.JPG.Large != "" {
		fmt.Printf("\nImage: %s\n", a.Images.JPG.Large)
	}
}

func join(r []jikan.Resource) string {
	s := make([]string, len(r))
	for i, x := range r {
		s[i] = x.Name
	}
	return strings.Join(s, ", ")
}
