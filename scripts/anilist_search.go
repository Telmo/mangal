package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/metafates/mangal/anilist"
	"github.com/metafates/mangal/query"
)

func main() {
	// Parse command line arguments
	flag.Parse()
	searchQuery := flag.Arg(0)
	if searchQuery == "" {
		fmt.Println("Usage: anilist_search <manga/anime name>")
		os.Exit(1)
	}

	// Try different variations of the search query
	queries := []string{
		searchQuery,
		strings.ToLower(searchQuery),
		strings.ToUpper(searchQuery),
		"TBATE", // Common abbreviation for The Beginning After The End
	}

	var allResults []*anilist.Manga
	for _, q := range queries {
		// Remember the query for better search results
		query.Remember(q, 1)

		// Search for the manga/anime
		results, err := anilist.SearchByName(q)
		if err != nil {
			continue
		}

		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		fmt.Println("No results found")
		return
	}

	// Remove duplicates based on ID
	seen := make(map[int]bool)
	uniqueResults := make([]*anilist.Manga, 0)
	for _, result := range allResults {
		if !seen[result.ID] {
			seen[result.ID] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	// Convert the results to JSON
	data, err := json.MarshalIndent(uniqueResults, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results to JSON: %v", err)
	}

	// Print the JSON output
	fmt.Println(string(data))
}
