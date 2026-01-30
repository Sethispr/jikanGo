package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Sethispr/jikanGo"
)

func main() {
	fmt.Print("Lookup a MAL username: ")
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		log.Fatal("Failed to read the input..")
	}
	
	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		log.Fatal("Username is empty..")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := jikan.New(jikan.WithTimeout(15 * time.Second))
	
	fmt.Printf("\nFetching profile for '%s'..\n\n", username)
	
	user, stats, err := fetchUserData(ctx, client, username)
	if err != nil {
		var apiErr *jikan.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			log.Fatalf("User '%s' not found on MyAnimeList", username)
		}
		log.Fatalf("API Error: %v", err)
	}

	printUser(user, stats)
	
	aboutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	about, err := client.User.About(aboutCtx, username)
	if err == nil && about != "" {
		fmt.Printf("About:\n%.200s...\n\n", about)
	}
}

func fetchUserData(ctx context.Context, client *jikan.Client, username string) (*jikan.User, *jikan.UserStatistics, error) {
	type result struct {
		user  *jikan.User
		stats *jikan.UserStatistics
		err   error
	}
	
	ch := make(chan result, 2)
	
	go func() {
		u, err := client.User.ByID(ctx, username)
		ch <- result{user: u, err: err}
	}()
	
	go func() {
		s, err := client.User.Statistics(ctx, username)
		ch <- result{stats: s, err: err}
	}()
	
	var user *jikan.User
	var stats *jikan.UserStatistics
	
	for i := 0; i < 2; i++ {
		r := <-ch
		if r.err != nil {
			return nil, nil, r.err
		}
		if r.user != nil {
			user = r.user
		}
		if r.stats != nil {
			stats = r.stats
		}
	}
	
	return user, stats, nil
}

func printUser(u *jikan.User, s *jikan.UserStatistics) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintf(w, "Username:\t%s\n", u.Username)
	fmt.Fprintf(w, "Profile:\t%s\n", u.URL)
	fmt.Fprintf(w, "Gender:\t%s\n", na(u.Gender))
	fmt.Fprintf(w, "Location:\t%s\n", na(u.Location))
	fmt.Fprintf(w, "Joined:\t%s\n", u.Joined)
	fmt.Fprintf(w, "Last Online:\t%s\n", u.LastOnline)
	if u.Images.JPG.ImageURL != "" {
		fmt.Fprintf(w, "Avatar:\t%s\n", u.Images.JPG.ImageURL)
	}
	w.Flush()

	if s != nil {
		fmt.Println("\nAnime Stats")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Days Watched:\t%.1f\n", s.Anime.DaysWatched)
		fmt.Fprintf(w, "Mean Score:\t%.2f\n", s.Anime.MeanScore)
		fmt.Fprintf(w, "Watching:\t%d\n", s.Anime.Watching)
		fmt.Fprintf(w, "Completed:\t%d\n", s.Anime.Completed)
		fmt.Fprintf(w, "On Hold:\t%d\n", s.Anime.OnHold)
		fmt.Fprintf(w, "Dropped:\t%d\n", s.Anime.Dropped)
		fmt.Fprintf(w, "Plan to Watch:\t%d\n", s.Anime.PlanToWatch)
		fmt.Fprintf(w, "Total Entries:\t%d\n", s.Anime.TotalEntries)
		fmt.Fprintf(w, "Episodes Watched:\t%d\n", s.Anime.EpisodesWatched)
		w.Flush()

		fmt.Println("\nManga Stats")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "Days Read:\t%.1f\n", s.Manga.DaysRead)
		fmt.Fprintf(w, "Mean Score:\t%.2f\n", s.Manga.MeanScore)
		fmt.Fprintf(w, "Reading:\t%d\n", s.Manga.Reading)
		fmt.Fprintf(w, "Completed:\t%d\n", s.Manga.Completed)
		fmt.Fprintf(w, "On Hold:\t%d\n", s.Manga.OnHold)
		fmt.Fprintf(w, "Dropped:\t%d\n", s.Manga.Dropped)
		fmt.Fprintf(w, "Chapters Read:\t%d\n", s.Manga.ChaptersRead)
		fmt.Fprintf(w, "Volumes Read:\t%d\n", s.Manga.VolumesRead)
		w.Flush()
	}
	
	fmt.Println()
}

func na(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
