package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// ----------------------------
// Kitsu API Manga Response
// ----------------------------
type MangaResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			CanonicalTitle   string `json:"canonicalTitle"`
			AbbreviatedTitle string `json:"abbreviatedTitle"`
			ChapterCount     int    `json:"chapterCount"`
			VolumeCount      int    `json:"volumeCount"`
			AverageRating    string `json:"averageRating"`
			PopularityRank   int    `json:"popularityRank"`
			Synopsis         string `json:"synopsis"`
			PosterImage      struct {
				Small  string `json:"small"`
				Medium string `json:"medium"`
				Large  string `json:"large"`
			} `json:"posterImage"`
		} `json:"attributes"`
	} `json:"data"`
}

// ----------------------------
// Precompiled regex for cleaning HTML
// ----------------------------
var htmlTagRegex = regexp.MustCompile(`<.*?>`)

// CleanHTML removes HTML tags and <br> from synopsis
func CleanHTML(input string) string {
	input = strings.ReplaceAll(input, "<br>", "\n")
	input = strings.ReplaceAll(input, "<br />", "\n")
	return strings.TrimSpace(htmlTagRegex.ReplaceAllString(input, ""))
}

// ----------------------------
// Search Manga by title
// ----------------------------
func searchManga(title string) error {
	encodedTitle := url.QueryEscape(title)
	apiURL := fmt.Sprintf("https://kitsu.io/api/edge/manga?filter[text]=%s&page[limit]=1", encodedTitle)

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Kitsu API returned status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response failed: %v", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("empty response from Kitsu API")
	}

	var result MangaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(result.Data) == 0 {
		fmt.Println("No manga found for:", title)
		return nil
	}

	manga := result.Data[0].Attributes

	fmt.Println("=======================================")
	fmt.Printf("Title             : %s\n", manga.CanonicalTitle)
	if manga.AbbreviatedTitle != "" {
		fmt.Printf("Short Title       : %s\n", manga.AbbreviatedTitle)
	}
	if manga.ChapterCount > 0 {
		fmt.Printf("Chapters          : %d\n", manga.ChapterCount)
	}
	if manga.VolumeCount > 0 {
		fmt.Printf("Volumes           : %d\n", manga.VolumeCount)
	}
	if manga.AverageRating != "" {
		fmt.Printf("Average Rating    : %s\n", manga.AverageRating)
	}
	if manga.PopularityRank > 0 {
		fmt.Printf("Popularity Rank   : %d\n", manga.PopularityRank)
	}
	fmt.Println("---------------------------------------")
	fmt.Println("Synopsis:")
	fmt.Println(CleanHTML(manga.Synopsis))
	fmt.Println("---------------------------------------")
	fmt.Printf("Cover Image       : %s\n", manga.PosterImage.Medium)
	fmt.Println("=======================================")

	return nil
}

// ----------------------------
// CLI Entry Point
// ----------------------------
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: kitsumanga <title>")
		return
	}

	title := strings.Join(os.Args[1:], " ")
	if err := searchManga(title); err != nil {
		fmt.Println("Error:", err)
	}
}