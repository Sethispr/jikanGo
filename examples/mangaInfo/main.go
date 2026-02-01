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
		id      = flag.Int("id", 0, "Manga ID (ex: 1 for Monster)")
		query   = flag.String("search", "", "Search manga by title")
		page    = flag.Int("page", 1, "Page number for search/results")
		news    = flag.Bool("news", false, "Show news for manga ID")
		reviews = flag.Bool("reviews", false, "Show reviews for manga ID")
		full    = flag.Bool("full", false, "Use /full endpoint (includes relations)")
	)
	flag.Parse()

	if *id == 0 && *query == "" {
		fmt.Fprint(os.Stderr, "Enter Manga ID or search by name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			log.Fatal("No input provided")
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
		if err := searchManga(ctx, client, *query, *page); err != nil {
			log.Fatalf("Search failed: %v", err)
		}
		return
	}

	if *id > 0 {
		switch {
		case *news:
			if err := showMangaNews(ctx, client, jikan.ID(*id), *page); err != nil {
				log.Fatalf("Failed to get news: %v", err)
			}
		case *reviews:
			if err := showMangaReviews(ctx, client, jikan.ID(*id), *page); err != nil {
				log.Fatalf("Failed to get reviews: %v", err)
			}
		case *full:
			if err := showMangaFull(ctx, client, jikan.ID(*id)); err != nil {
				log.Fatalf("Failed to get manga: %v", err)
			}
		default:
			if err := showMangaDetails(ctx, client, jikan.ID(*id)); err != nil {
				log.Fatalf("Failed to get manga: %v", err)
			}
		}
	}
}

func searchManga(ctx context.Context, c *jikan.Client, query string, page int) error {
	opts := jikan.MangaSearchOptions{
		Query: query,
		Page:  page,
	}
	
	results, pagination, err := c.Manga.Search(ctx, opts)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No manga found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tTYPE\tSTATUS\tCHAPTERS")
	fmt.Fprintln(w, "-\t-\t-\t-\t-")

	for _, m := range results {
		title := truncate(m.Title, 40)
		chapters := nullInt(m.Chapters)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			m.MalID, title, m.Type, m.Status, chapters)
	}
	w.Flush()

	if pagination.LastPage > 1 {
		fmt.Fprintf(os.Stderr, "\nPage %d of %d (Total: %d)\n",
			page, pagination.LastPage, pagination.Items.Total)
	}
	return nil
}

func showMangaDetails(ctx context.Context, c *jikan.Client, id jikan.ID) error {
	manga, err := c.Manga.ByID(ctx, id)
	if err != nil {
		var apiErr *jikan.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return fmt.Errorf("manga %d not found", id)
		}
		return err
	}

	type results struct {
		chars []jikan.MangaCharacter
		stats *jikan.MangaStats
		recs  []jikan.MangaRecommendation
	}

	resCh := make(chan results, 1)
	go func() {
		var r results
		r.chars, _ = c.Manga.Characters(ctx, id)
		r.stats, _ = c.Manga.Statistics(ctx, id)
		r.recs, _ = c.Manga.Recommendations(ctx, id)
		resCh <- r
	}()

	res := <-resCh
	printManga(manga, res.chars, res.stats, res.recs)
	return nil
}

func showMangaFull(ctx context.Context, c *jikan.Client, id jikan.ID) error {
	manga, err := c.Manga.Full(ctx, id)
	if err != nil {
		var apiErr *jikan.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return fmt.Errorf("manga %d not found", id)
		}
		return err
	}

	chars, _ := c.Manga.Characters(ctx, id)
	printMangaFull(manga, chars)
	return nil
}

func showMangaNews(ctx context.Context, c *jikan.Client, id jikan.ID, page int) error {
	news, pagination, err := c.Manga.News(ctx, id, page)
	if err != nil {
		return err
	}

	if len(news) == 0 {
		fmt.Println("No news found for this manga")
		return nil
	}

	fmt.Printf("News for Manga %d:\n\n", id)
	for _, n := range news {
		date := n.Date.Format("2006-01-02")
		title := truncate(n.Title, 60)
		fmt.Printf("[%s] %s\n", date, title)
		fmt.Printf("    By: %s | Comments: %d\n", n.AuthorUsername, n.Comments)
		fmt.Printf("    %s\n\n", truncate(n.Excerpt, 100))
	}

	if pagination.LastPage > 1 {
		fmt.Fprintf(os.Stderr, "Page %d of %d\n", page, pagination.LastPage)
	}
	return nil
}

func showMangaReviews(ctx context.Context, c *jikan.Client, id jikan.ID, page int) error {
	reviews, pagination, err := c.Manga.Reviews(ctx, id, page, false, false)
	if err != nil {
		return err
	}

	if len(reviews) == 0 {
		fmt.Println("No reviews found")
		return nil
	}

	fmt.Printf("Reviews for Manga %d:\n\n", id)
	for _, r := range reviews {
		fmt.Printf("%s - Score: %d/10 | Chapters: %d\n", r.User.Username, r.Score, r.ChaptersRead)
		if len(r.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(r.Tags, ", "))
		}
		fmt.Printf("%s\n\n", truncate(r.Review, 200))
	}

	if pagination.LastPage > 1 {
		fmt.Fprintf(os.Stderr, "Page %d of %d\n", page, pagination.LastPage)
	}
	return nil
}

func printManga(m *jikan.Manga, chars []jikan.MangaCharacter, stats *jikan.MangaStats, recs []jikan.MangaRecommendation) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Title:\t%s\n", m.Title)
	if m.TitleEnglish != "" && m.TitleEnglish != m.Title {
		fmt.Fprintf(w, "English:\t%s\n", m.TitleEnglish)
	}
	if m.TitleJapanese != "" {
		fmt.Fprintf(w, "Japanese:\t%s\n", m.TitleJapanese)
	}
	fmt.Fprintf(w, "ID:\t%d\n", m.MalID)
	fmt.Fprintf(w, "Type:\t%s\n", m.Type)
	fmt.Fprintf(w, "Status:\t%s\n", m.Status)
	fmt.Fprintf(w, "Chapters:\t%s\n", nullInt(m.Chapters))
	fmt.Fprintf(w, "Volumes:\t%s\n", nullInt(m.Volumes))
	fmt.Fprintf(w, "Score:\t%s\n", nullFloat(m.Score))
	fmt.Fprintf(w, "Rank:\t%s\n", nullInt(m.Rank))
	fmt.Fprintf(w, "Members:\t%s\n", nullInt(m.Members))
	fmt.Fprintf(w, "Favorites:\t%s\n", nullInt(m.Favorites))
	fmt.Fprintf(w, "Published:\t%s to %s\n", m.Published.From, m.Published.To)
	fmt.Fprintf(w, "URL:\t%s\n", m.URL)
	w.Flush()

	if len(m.Authors) > 0 {
		fmt.Printf("\nAuthors: %s\n", joinResources(m.Authors, 5))
	}
	if len(m.Genres) > 0 {
		fmt.Printf("Genres: %s\n", joinResources(m.Genres, 10))
	}
	if m.Synopsis != "" {
		fmt.Printf("\nSynopsis:\n%s\n", truncate(m.Synopsis, 500))
	}

	if len(chars) > 0 {
		fmt.Printf("\nCharacters (%d):\n", len(chars))
		cw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(cw, "ROLE\tNAME")
		for i, c := range chars {
			if i >= 10 {
				fmt.Fprintf(cw, "...\tand %d more\n", len(chars)-10)
				break
			}
			name := truncate(c.Character.Name, 30)
			fmt.Fprintf(cw, "%s\t%s\n", c.Role, name)
		}
		cw.Flush()
	}

	if stats != nil {
		fmt.Printf("\nStatistics:\n")
		fmt.Printf("  Reading: %d | Completed: %d | Dropped: %d\n", 
			stats.Reading, stats.Completed, stats.Dropped)
		fmt.Printf("  Plan to Read: %d | Total: %d\n", stats.PlanToRead, stats.Total)
	}

	if len(recs) > 0 {
		fmt.Printf("\nRecommendations (%d):\n", len(recs))
		for i, r := range recs {
			if i >= 5 {
				fmt.Printf("... and %d more\n", len(recs)-5)
				break
			}
			fmt.Printf("  %s (%d votes)\n", r.Entry.Title, r.Votes)
		}
	}
}

func printMangaFull(m *jikan.MangaFull, chars []jikan.MangaCharacter) {
	printManga(&m.Manga, chars, nil, nil)

	if len(m.Relations) > 0 {
		fmt.Printf("\nRelations:\n")
		for _, rel := range m.Relations {
			if len(rel.Entry) == 0 {
				continue
			}
			names := make([]string, 0, len(rel.Entry))
			for i, e := range rel.Entry {
				if i >= 3 {
					names = append(names, fmt.Sprintf("+%d more", len(rel.Entry)-3))
					break
				}
				names = append(names, e.Name)
			}
			fmt.Printf("  %s: %s\n", rel.Relation, strings.Join(names, ", "))
		}
	}

	if len(m.External) > 0 {
		fmt.Printf("\nExternal Links:\n")
		for _, ext := range m.External {
			fmt.Printf("  %s: %s\n", ext.Name, ext.URL)
		}
	}
}

func joinResources(r []jikan.Resource, limit int) string {
	if len(r) == 0 {
		return ""
	}
	names := make([]string, 0, limit)
	for i := range r {
		if i >= limit {
			names = append(names, fmt.Sprintf("+%d", len(r)-limit))
			break
		}
		names = append(names, r[i].Name)
	}
	return strings.Join(names, ", ")
}

func nullInt(n *int) string {
	if n == nil {
		return "?"
	}
	return strconv.Itoa(*n)
}

func nullFloat(n *float64) string {
	if n == nil {
		return "?"
	}
	return fmt.Sprintf("%.2f", *n)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

