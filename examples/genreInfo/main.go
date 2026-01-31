package main

import (
	"bufio"
	"context"
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
		anime      = flag.Bool("anime", false, "List anime genres")
		manga      = flag.Bool("manga", false, "List manga genres")
		filter     = flag.String("filter", "", "Filter by: genres, explicit_genres, themes, demographics")
		page       = flag.Int("page", 1, "Page number")
		limit      = flag.Int("limit", 0, "Results per page (max 100)")
		listFilter = flag.Bool("list-filters", false, "Show available filters")
	)
	flag.Parse()

	if *listFilter {
		fmt.Println("Available filters:")
		fmt.Println("  genres")
		fmt.Println("  explicit_genres")
		fmt.Println("  themes")
		fmt.Println("  demographics")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := jikan.New(jikan.WithTimeout(10 * time.Second))

	if !*anime && !*manga {
		fmt.Fprint(os.Stderr, "Select type:\n1. Anime\n2. Manga\nChoice: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			choice := strings.TrimSpace(scanner.Text())
			switch choice {
			case "1", "anime":
				*anime = true
			case "2", "manga":
				*manga = true
			default:
				log.Fatal("Invalid choice")
			}
		}
	}

	if *anime {
		if err := listGenres(ctx, client, "anime", *filter, *page, *limit); err != nil {
			log.Fatalf("Error: %v", err)
		}
		return
	}

	if *manga {
		if err := listGenres(ctx, client, "manga", *filter, *page, *limit); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}

func listGenres(ctx context.Context, c *jikan.Client, mediaType, filterStr string, page, limit int) error {
	var f jikan.GenreFilter
	switch filterStr {
	case "genres":
		f = jikan.GenreGenres
	case "explicit_genres", "explicit":
		f = jikan.GenreExplicit
	case "themes":
		f = jikan.GenreThemes
	case "demographics", "demo":
		f = jikan.GenreDemographics
	case "":
		f = ""
	default:
		return fmt.Errorf("invalid filter: %s", filterStr)
	}

	var genres []*jikan.Genre
	var pagination *jikan.Pagination
	var err error

	if mediaType == "anime" {
		genres, pagination, err = c.Genre.Anime(ctx, f, page, limit)
	} else {
		genres, pagination, err = c.Genre.Manga(ctx, f, page, limit)
	}

	if err != nil {
		return err
	}

	if len(genres) == 0 {
		fmt.Println("No genres found..")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tCOUNT\tTYPE\n")
	fmt.Fprintf(w, "-\t-\t-\t-\n")

	for _, g := range genres {
		category := normalizeFilter(f)
		if category == "" {
			category = "all"
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", g.MalID, g.Name, g.Count, category)
	}
	w.Flush()

	if pagination != nil && pagination.LastPage > 1 {
		fmt.Fprintf(os.Stderr, "\nPage %d of %d (Total: %d)\n",
			page, pagination.LastPage, pagination.Items.Total)
	}

	return nil
}

func normalizeFilter(f jikan.GenreFilter) string {
	switch f {
	case jikan.GenreGenres:
		return "genres"
	case jikan.GenreExplicit:
		return "explicit"
	case jikan.GenreThemes:
		return "themes"
	case jikan.GenreDemographics:
		return "demographics"
	default:
		return ""
	}
}
