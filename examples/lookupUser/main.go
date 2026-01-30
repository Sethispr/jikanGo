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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := jikan.New(jikan.WithTimeout(10 * time.Second))
	
	fmt.Printf("\nGetting profile for '%s'..\n\n", username)
	
	user, err := client.User.ByID(ctx, username)
	if err != nil {
		var apiErr *jikan.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			log.Fatalf("User '%s' not found on MyAnimeList", username)
		}
		log.Fatalf("API Error: %v", err)
	}

	printUser(user)
}

func printUser(u *jikan.User) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintf(w, "Username:\t%s\n", u.Username)
	fmt.Fprintf(w, "Profile:\t%s\n", u.URL)
	fmt.Fprintf(w, "Gender:\t%s\n", na(u.Gender))
	fmt.Fprintf(w, "Joined:\t%s\n", u.Joined)
	fmt.Fprintf(w, "Last Online:\t%s\n", u.LastOnline)
	if u.Images.JPG.ImageURL != "" {
		fmt.Fprintf(w, "Avatar:\t%s\n", u.Images.JPG.ImageURL)
	}
	w.Flush()

	fmt.Println("\nAnime Stats")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Days Watched:\t%.1f\n", u.Anime.Days)
	fmt.Fprintf(w, "Mean Score:\t%.2f\n", u.Anime.MeanScore)
	fmt.Fprintf(w, "Total Entries:\t%d\n", u.Anime.Total)
	fmt.Fprintf(w, "Completed:\t%d\n", u.Anime.Completed)
	fmt.Fprintf(w, "Watching:\t%d\n", u.Anime.Completed) // not implemented yet
	w.Flush()

	fmt.Println("\nManga Stats")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Days Read:\t%.1f\n", u.Manga.Days)
	fmt.Fprintf(w, "Mean Score:\t%.2f\n", u.Manga.MeanScore)
	fmt.Fprintf(w, "Total Entries:\t%d\n", u.Manga.Total)
	fmt.Fprintf(w, "Completed:\t%d\n", u.Manga.Completed)
	w.Flush()
	
	fmt.Println()
}

func na(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
