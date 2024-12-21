package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type MangaDexResponse struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Title map[string]string `json:"title"`
			AltTitles []map[string]string `json:"altTitles"`
			Description map[string]string `json:"description"`
			Status string `json:"status"`
			Year   *int   `json:"year"`
			Tags []struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Attributes struct {
					Name map[string]string `json:"name"`
					Group string `json:"group"`
				} `json:"attributes"`
			} `json:"tags"`
			ContentRating string `json:"contentRating"`
			OriginalLanguage string `json:"originalLanguage"`
		} `json:"attributes"`
		Relationships []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"relationships"`
	} `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

func main() {
	// Parse command line arguments
	flag.Parse()
	searchQuery := flag.Arg(0)
	if searchQuery == "" {
		fmt.Println("Usage: mangadex_search <manga/manhwa name>")
		os.Exit(1)
	}

	// Create the search URL
	baseURL := "https://api.mangadex.org/manga"
	params := url.Values{}
	params.Add("title", searchQuery)
	params.Add("limit", "5")
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("includes[]", "artist")
	params.Add("order[relevance]", "desc")

	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Make the request
	resp, err := http.Get(searchURL)
	if err != nil {
		log.Fatalf("Failed to search MangaDex: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	// Parse the response
	var result MangaDexResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Check if we found any results
	if len(result.Data) == 0 {
		fmt.Println("No results found")
		return
	}

	// Convert the results to a more readable format
	type FormattedResult struct {
		Title       string   `json:"title"`
		AltTitles   []string `json:"altTitles"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		Year        *int     `json:"year"`
		Tags        []string `json:"tags"`
		Type        string   `json:"type"`
		Language    string   `json:"originalLanguage"`
		Rating      string   `json:"contentRating"`
	}

	formattedResults := make([]FormattedResult, 0)
	for _, manga := range result.Data {
		// Get the English title or the first available title
		title := ""
		if t, ok := manga.Attributes.Title["en"]; ok {
			title = t
		} else {
			for _, t := range manga.Attributes.Title {
				title = t
				break
			}
		}

		// Get the English description or the first available description
		description := ""
		if d, ok := manga.Attributes.Description["en"]; ok {
			description = d
		} else {
			for _, d := range manga.Attributes.Description {
				description = d
				break
			}
		}

		// Get alternative titles
		altTitles := make([]string, 0)
		for _, titles := range manga.Attributes.AltTitles {
			for _, title := range titles {
				altTitles = append(altTitles, title)
			}
		}

		// Get tags
		tags := make([]string, 0)
		for _, tag := range manga.Attributes.Tags {
			if name, ok := tag.Attributes.Name["en"]; ok {
				tags = append(tags, name)
			}
		}

		formattedResults = append(formattedResults, FormattedResult{
			Title:       title,
			AltTitles:   altTitles,
			Description: description,
			Status:      manga.Attributes.Status,
			Year:        manga.Attributes.Year,
			Tags:        tags,
			Type:        manga.Type,
			Language:    manga.Attributes.OriginalLanguage,
			Rating:      manga.Attributes.ContentRating,
		})
	}

	// Convert the formatted results to JSON
	data, err := json.MarshalIndent(formattedResults, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results to JSON: %v", err)
	}

	// Write the JSON output to /tmp/series.json
	outputFile := "/tmp/series.json"
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Search results written to %s\n", outputFile)
}
