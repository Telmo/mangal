package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/metafates/mangal/anilist"
	"github.com/metafates/mangal/log"
	mdapi "github.com/metafates/mangal/mangadex"
	"github.com/metafates/mangal/model"
	"github.com/metafates/mangal/provider/mangadex"
	"github.com/metafates/mangal/source"
)

// SearchManga searches for manga across multiple sources (Anilist and MangaDex)
// It first tries Anilist, and if no results are found or description is empty, falls back to MangaDex
func SearchManga(query string) (*source.Manga, error) {
	log.Infof("Searching for manga: %s", query)

	// Try Anilist first
	anilistResults, err := anilist.SearchByName(query)
	if err != nil {
		log.Errorf("Anilist search error: %v", err)
	}

	log.Infof("Got %d results from Anilist", len(anilistResults))

	if err == nil && len(anilistResults) > 0 {
		// Debug: Write Anilist response to file
		log.Info("Attempting to marshal Anilist result to JSON")
		debugJson, err := json.MarshalIndent(anilistResults[0], "", "  ")
		if err != nil {
			log.Errorf("Failed to marshal debug json: %v", err)
		} else {
			debugFile := "/tmp/debug.json"
			log.Infof("Writing debug json to %s", debugFile)

			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(debugFile), 0755); err != nil {
				log.Errorf("Failed to create debug file directory: %v", err)
			} else {
				if err := os.WriteFile(debugFile, debugJson, 0644); err != nil {
					log.Errorf("Failed to write debug json: %v", err)
				} else {
					log.Infof("Successfully wrote debug json to %s", debugFile)
				}
			}
		}

		result := anilistResults[0]

		// Convert tags to string slice
		tags := make([]string, 0, len(result.Tags))
		for _, tag := range result.Tags {
			if tag.Name != "" {
				tags = append(tags, tag.Name)
			}
		}

		// Convert characters to string slice
		characters := make([]string, 0, len(result.Characters.Nodes))
		for _, char := range result.Characters.Nodes {
			if char.Name.Full != "" {
				characters = append(characters, char.Name.Full)
			}
		}

		// Convert staff to appropriate fields
		var story, art, translation, lettering []string
		for _, edge := range result.Staff.Edges {
			if edge.Node.Name.Full == "" {
				continue
			}
			switch edge.Role {
			case "Story":
				story = append(story, edge.Node.Name.Full)
			case "Art":
				art = append(art, edge.Node.Name.Full)
			case "Translation":
				translation = append(translation, edge.Node.Name.Full)
			case "Lettering":
				lettering = append(lettering, edge.Node.Name.Full)
			}
		}

		// Convert URLs to string slice
		urls := make([]string, 0, len(result.External))
		for _, ext := range result.External {
			if ext.URL != "" {
				urls = append(urls, ext.URL)
			}
		}

		// Convert Anilist manga to source.Manga
		manga := &source.Manga{
			Name:   result.Name(),
			URL:    result.SiteURL,
			ID:     fmt.Sprintf("%d", result.ID),
			Source: mangadex.New(),
			Metadata: model.MangaMetadata{
				Summary:   strings.TrimSpace(result.Description),
				Publisher: "", // Anilist doesn't provide publisher info
				Status:    result.Status,
				StartDate: model.Date{
					Year:  result.StartDate.Year,
					Month: result.StartDate.Month,
				},
				EndDate: model.Date{
					Year:  result.EndDate.Year,
					Month: result.EndDate.Month,
				},
				Chapters: result.Chapters,
				Cover: struct {
					ExtraLarge string `json:"extraLarge"`
					Large      string `json:"large"`
					Medium     string `json:"medium"`
					Color      string `json:"color"`
				}{
					ExtraLarge: result.CoverImage.ExtraLarge,
					Large:      result.CoverImage.Large,
					Medium:     result.CoverImage.Medium,
					Color:      result.CoverImage.Color,
				},
				BannerImage: result.BannerImage,
				Genres:      result.Genres,
				Tags:        tags,
				Characters:  characters,
				Staff: struct {
					Story       []string `json:"story"`
					Art         []string `json:"art"`
					Translation []string `json:"translation"`
					Lettering   []string `json:"lettering"`
				}{
					Story:       story,
					Art:         art,
					Translation: translation,
					Lettering:   lettering,
				},
				Format:       "MANGA",
				Volumes:      0,
				AverageScore: result.AverageScore,
				Popularity:   result.Popularity,
				MeanScore:    result.MeanScore,
				IsLicensed:   result.IsLicensed,
				UpdatedAt:    0,
			},
		}
		return manga, nil
	} else {
		log.Info("No results found on Anilist, trying MangaDex...")
	}

	// If Anilist fails, returns no results, or has empty description, try MangaDex
	// Create MangaDex client
	client := mdapi.NewClient()
	mangadexResult, err := client.SearchManga(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search on both Anilist and MangaDex: %v", err)
	}

	// If no manga found on MangaDex, return error
	if mangadexResult == nil {
		return nil, fmt.Errorf("no manga found with query: %s", query)
	}

	// Convert tags to string slice
	tags := make([]string, 0, len(mangadexResult.Tags))
	for _, tag := range mangadexResult.Tags {
		if tag.Name != "" {
			tags = append(tags, tag.Name)
		}
	}

	// Convert characters to string slice
	characters := make([]string, 0, len(mangadexResult.Characters.Nodes))
	for _, char := range mangadexResult.Characters.Nodes {
		if char.Name.Full != "" {
			characters = append(characters, char.Name.Full)
		}
	}

	// Convert staff to appropriate fields
	var story, art, translation, lettering []string
	for _, edge := range mangadexResult.Staff.Edges {
		if edge.Node.Name.Full == "" {
			continue
		}
		switch edge.Role {
		case "Story":
			story = append(story, edge.Node.Name.Full)
		case "Art":
			art = append(art, edge.Node.Name.Full)
		case "Translation":
			translation = append(translation, edge.Node.Name.Full)
		case "Lettering":
			lettering = append(lettering, edge.Node.Name.Full)
		}
	}

	// Convert URLs to string slice
	urls := make([]string, 0, len(mangadexResult.External))
	for _, ext := range mangadexResult.External {
		if ext.URL != "" {
			urls = append(urls, ext.URL)
		}
	}

	// Convert MangaDex manga to source.Manga
	manga := &source.Manga{
		Name:   mangadexResult.Name(),
		URL:    mangadexResult.SiteURL,
		ID:     fmt.Sprintf("%d", mangadexResult.ID),
		Source: mangadex.New(),
		Metadata: model.MangaMetadata{
			Summary:   strings.TrimSpace(mangadexResult.Description),
			Publisher: "", // MangaDex doesn't provide publisher info
			Status:    mangadexResult.Status,
			StartDate: model.Date{
				Year:  mangadexResult.StartDate.Year,
				Month: mangadexResult.StartDate.Month,
			},
			EndDate: model.Date{
				Year:  mangadexResult.EndDate.Year,
				Month: mangadexResult.EndDate.Month,
			},
			Chapters: mangadexResult.Chapters,
			Cover: struct {
				ExtraLarge string `json:"extraLarge"`
				Large      string `json:"large"`
				Medium     string `json:"medium"`
				Color      string `json:"color"`
			}{
				ExtraLarge: mangadexResult.CoverImage.ExtraLarge,
				Large:      mangadexResult.CoverImage.Large,
				Medium:     mangadexResult.CoverImage.Medium,
				Color:      mangadexResult.CoverImage.Color,
			},
			BannerImage: mangadexResult.BannerImage,
			Genres:      mangadexResult.Genres,
			Tags:        tags,
			Characters:  characters,
			Staff: struct {
				Story       []string `json:"story"`
				Art         []string `json:"art"`
				Translation []string `json:"translation"`
				Lettering   []string `json:"lettering"`
			}{
				Story:       story,
				Art:         art,
				Translation: translation,
				Lettering:   lettering,
			},
			Format:       "MANGA",
			Volumes:      0,
			AverageScore: mangadexResult.AverageScore,
			Popularity:   mangadexResult.Popularity,
			MeanScore:    mangadexResult.MeanScore,
			IsLicensed:   mangadexResult.IsLicensed,
			UpdatedAt:    0,
		},
	}

	// If manga is still missing critical metadata, return error
	if manga.Metadata.Summary == "" {
		return nil, fmt.Errorf("no metadata found for manga: %s", query)
	}

	return manga, nil
}

// SearchAndSave searches for manga and saves the metadata to the specified output path
func SearchAndSave(query, outputPath string) (*source.Manga, error) {
	manga, err := SearchManga(query)
	if err != nil {
		return nil, err
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Save metadata to the specified file
	if err := manga.SaveMetadataToFile(outputPath); err != nil {
		return nil, fmt.Errorf("failed to save metadata to file: %v", err)
	}

	return manga, nil
}
