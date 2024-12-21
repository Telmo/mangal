package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/network"
	"github.com/metafates/mangal/query"
	"github.com/metafates/mangal/where"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type searchByNameResponse struct {
	Data struct {
		Page struct {
			Media []*Manga `json:"media"`
		} `json:"page"`
	} `json:"data"`
}

type searchByIDResponse struct {
	Data struct {
		Media *Manga `json:"media"`
	} `json:"data"`
}

// GetByID returns a manga by its ID.
// If the manga is not found, it returns nil.
func GetByID(id int) (*Manga, error) {
	if manga := idCacher.Get(id); manga.IsPresent() {
		return manga.MustGet(), nil
	}

	// prepare body
	log.Infof("Getting manga with id %d from Anilist", id)
	body := map[string]any{
		"query": searchByIDQuery,
		"variables": map[string]any{
			"id": id,
		},
	}

	// parse body to json
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// send request
	log.Info("Sending request to Anilist")
	req, err := http.NewRequest(http.MethodPost, "https://graphql.anilist.co", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := network.Client.Do(req)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("Anilist returned status code " + strconv.Itoa(resp.StatusCode))
		return nil, fmt.Errorf("invalid response code %d", resp.StatusCode)
	}

	// decode response
	var response searchByIDResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Error(err)
		return nil, err
	}

	manga := response.Data.Media
	if manga == nil {
		return nil, fmt.Errorf("manga with id %d not found", id)
	}

	_ = idCacher.Set(manga.ID, manga)
	return manga, nil
}

// SearchByName returns a list of mangas that match the given name.
func SearchByName(name string) ([]*Manga, error) {
	name = normalizedName(name)
	_ = query.Remember(name, 1)

	if _, failed := failCacher.Get(name).Get(); failed {
		return nil, fmt.Errorf("failed to search for %s", name)
	}

	if ids, ok := searchCacher.Get(name).Get(); ok {
		mangas := lo.FilterMap(ids, func(item, _ int) (*Manga, bool) {
			return idCacher.Get(item).Get()
		})

		if len(mangas) == 0 {
			_ = searchCacher.Delete(name)
			return SearchByName(name)
		}

		return mangas, nil
	}

	// Try variations of the name
	variations := []string{name}

	// Add name without special characters
	cleanName := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, name)
	if cleanName != name {
		variations = append(variations, cleanName)
	}

	// Add name without volume/chapter numbers
	re := regexp.MustCompile(`(?i)\s*(vol(\.|ume)?|ch(\.|apter)?)\s*\d+.*$`)
	strippedName := re.ReplaceAllString(name, "")
	if strippedName != name {
		variations = append(variations, strippedName)
	}

	// Try each variation
	for _, variant := range variations {
		// prepare body
		log.Infof("Searching anilist for manga %s", variant)
		body := map[string]any{
			"query": searchByNameQuery,
			"variables": map[string]any{
				"query": variant,
			},
		}

		// parse body to json
		jsonBody, err := json.Marshal(body)
		if err != nil {
			log.Error(err)
			continue
		}

		// send request
		log.Info("Sending request to Anilist")
		req, err := http.NewRequest(http.MethodPost, "https://graphql.anilist.co", bytes.NewBuffer(jsonBody))
		if err != nil {
			log.Error(err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := network.Client.Do(req)

		if err != nil {
			log.Error(err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Error("Anilist returned status code " + strconv.Itoa(resp.StatusCode))
			continue
		}

		// decode response
		var response searchByNameResponse

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Error(err)
			continue
		}

		mangas := response.Data.Page.Media
		if len(mangas) > 0 {
			log.Infof("Got response from Anilist, found %d results", len(mangas))

			// Write debug file if debug flag is set
			if viper.GetBool(key.MetadataDebug) {
				debugDir := filepath.Join(where.Config(), "debug")
				debugFile := filepath.Join(debugDir, "anilist_search.json")
				log.Infof("Writing search results to debug directory: %s", debugFile)
				if data, err := json.MarshalIndent(mangas, "", "  "); err != nil {
					log.Errorf("Failed to marshal search results: %v", err)
				} else {
					if err := os.MkdirAll(debugDir, 0755); err != nil {
						log.Errorf("Failed to create debug directory: %v", err)
					} else {
						if err := os.WriteFile(debugFile, data, 0644); err != nil {
							log.Errorf("Failed to write search results: %v", err)
						} else {
							log.Infof("Successfully wrote search results to %s", debugFile)
						}
					}
				}
			}

			ids := make([]int, len(mangas))
			for i, manga := range mangas {
				ids[i] = manga.ID
				_ = idCacher.Set(manga.ID, manga)
			}
			_ = searchCacher.Set(name, ids)
			return mangas, nil
		}
	}

	// If all variations failed
	_ = failCacher.Set(name, true)
	return nil, fmt.Errorf("no results found for any variation of %s", name)
}
