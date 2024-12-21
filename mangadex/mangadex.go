package mangadex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/metafates/mangal/anilist"
)

const (
	BaseURL = "https://api.mangadex.org"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// SearchManga searches for manga on MangaDex and returns them in Anilist format
func (c *Client) SearchManga(query string) (*anilist.Manga, error) {
	// Create the search URL
	params := url.Values{}
	params.Add("title", query)
	params.Add("limit", "5")
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("includes[]", "artist")
	params.Add("order[relevance]", "desc")

	searchURL := fmt.Sprintf("%s/manga?%s", BaseURL, params.Encode())

	// Make the request
	resp, err := c.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search MangaDex: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse the response
	var result MangaDexResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if we found any results
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	// Convert the first result to Anilist format
	return result.Data[0].ToAnilist(), nil
}

// GetManga gets a specific manga by ID from MangaDex
func (c *Client) GetManga(id string) (*anilist.Manga, error) {
	mangaURL := fmt.Sprintf("%s/manga/%s", BaseURL, id)
	params := url.Values{}
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("includes[]", "artist")

	resp, err := c.httpClient.Get(fmt.Sprintf("%s?%s", mangaURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to get manga: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		Data MangaDexManga `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Data.ToAnilist(), nil
}

// GetCoverArt gets the cover art URL for a manga
func (c *Client) GetCoverArt(mangaID string, filename string) (string, error) {
	return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mangaID, filename), nil
}
