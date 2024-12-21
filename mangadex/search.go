package mangadex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/metafates/mangal/anilist"
)

// SearchAndSave searches for a manga and saves it to the series.json file
func SearchAndSave(query, outputPath string) (*anilist.Manga, error) {
	client := NewClient()
	
	// Search for the manga
	manga, err := client.SearchManga(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search manga: %v", err)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Convert to JSON
	data, err := json.MarshalIndent(manga, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manga to JSON: %v", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write series.json: %v", err)
	}

	return manga, nil
}
