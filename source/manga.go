package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/metafates/mangal/anilist"
	"github.com/metafates/mangal/db"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/mangadex"
	"github.com/metafates/mangal/model"
	"github.com/metafates/mangal/util"
	"github.com/metafates/mangal/where"
	"github.com/samber/mo"
	"github.com/spf13/viper"
)

type date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// Manga is a manga from a source.
type Manga struct {
	// Name of the manga.
	Name string `json:"name" jsonschema:"description=Name of the manga"`
	// URL to manga's page.
	URL string `json:"url" jsonschema:"description=URL to manga's page"`
	// ID of manga in the source.
	ID string `json:"id" jsonschema:"description=ID of manga in the source"`
	// Index of the manga in search results
	Index uint16 `json:"index" jsonschema:"description=Index of the manga in search results"`
	// Chapters of the manga
	Chapters []*Chapter `json:"chapters" jsonschema:"description=Chapters of the manga"`
	// Source that the manga belongs to.
	Source Source `json:"-"`
	// Anilist is the closest anilist match
	Anilist  mo.Option[*anilist.Manga] `json:"-"`
	Metadata model.MangaMetadata `json:"metadata"`
	
	cachedTempPath  string
	populated       bool
	coverDownloaded bool
}

func (m *Manga) ToModel() *model.Manga {
	return &model.Manga{
		ID:          m.ID,
		Title:       m.Name,
		URL:         m.URL,
		Description: m.Metadata.Summary,
		SourceID:    m.Source.ID(),
		SourceName:  m.Source.Name(),
		Metadata:    m.Metadata,
	}
}

func FromModel(modelManga *model.Manga, src Source) *Manga {
	return &Manga{
		ID:       modelManga.ID,
		Name:     modelManga.Title,
		URL:      modelManga.URL,
		Source:   src,
		Metadata: modelManga.Metadata,
		Chapters: make([]*Chapter, 0),
	}
}

func (m *Manga) String() string {
	return m.Name
}

func (m *Manga) Dirname() string {
	return util.SanitizeFilename(m.Name)
}

func (m *Manga) peekPath() string {
	path := where.Downloads()

	if viper.GetBool(key.DownloaderCreateMangaDir) {
		path = filepath.Join(path, m.Dirname())
	}

	return path
}

func (m *Manga) Path(temp bool) (path string, err error) {
	if temp {
		if path = m.cachedTempPath; path != "" {
			return
		}

		path = where.Temp()
		m.cachedTempPath = path
		return
	}

	path = m.peekPath()
	_ = filesystem.Api().MkdirAll(path, os.ModePerm)
	return
}

func (m *Manga) GetCover() (string, error) {
	var covers = []string{
		m.Metadata.Cover.ExtraLarge,
		m.Metadata.Cover.Large,
		m.Metadata.Cover.Medium,
	}

	for _, cover := range covers {
		if cover != "" {
			return cover, nil
		}
	}

	return "", fmt.Errorf("no cover found")
}

func (m *Manga) DownloadCover(overwrite bool, path string, progress func(string)) error {
	if m.coverDownloaded {
		return nil
	}
	m.coverDownloaded = true

	log.Info("Downloading cover for ", m.Name)
	progress("Downloading cover")

	cover, err := m.GetCover()
	if err != nil {
		log.Warn(err)
		return nil
	}

	var extension string
	if extension = filepath.Ext(cover); extension == "" {
		extension = ".jpg"
	}

	path = filepath.Join(path, "cover"+extension)

	if !overwrite {
		exists, err := filesystem.Api().Exists(path)
		if err != nil {
			log.Error(err)
			return err
		}

		if exists {
			log.Warn("Cover already exists")
			return nil
		}
	}

	resp, err := http.Get(cover)
	if err != nil {
		log.Error(err)
		return err
	}

	defer util.Ignore(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		log.Error(err)
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}

	err = filesystem.Api().WriteFile(path, data, os.ModePerm)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("Cover downloaded")
	return nil
}

func (m *Manga) BindWithAnilist() error {
	if m.Anilist.IsPresent() {
		return nil
	}

	log.Infof("binding %s with anilist", m.Name)

	manga, err := anilist.FindClosest(m.Name)
	if err != nil {
		log.Error(err)
		return err
	}

	m.Anilist = mo.Some(manga)
	return nil
}

func (m *Manga) PopulateMetadata(progress func(string)) error {
	if m.populated {
		return nil
	}

	if progress != nil {
		progress("Getting metadata...")
	}

	// Initialize metadata fields with defaults
	m.Metadata.Status = "Unknown"
	m.Metadata.Format = "MANGA"
	m.Metadata.Genres = make([]string, 0)
	m.Metadata.Tags = make([]string, 0)
	m.Metadata.Characters = make([]string, 0)
	m.Metadata.URLs = make([]string, 0)
	m.Metadata.Synonyms = make([]string, 0)
	m.Metadata.Staff.Story = make([]string, 0)
	m.Metadata.Staff.Art = make([]string, 0)
	m.Metadata.Staff.Translation = make([]string, 0)
	m.Metadata.Staff.Lettering = make([]string, 0)

	// Try to bind with Anilist first
	if err := m.BindWithAnilist(); err != nil {
		log.Warn("Failed to bind with Anilist:", err)
		
		// Try MangaDex as fallback
		log.Info("Trying MangaDex as fallback...")
		client := mangadex.NewClient()
		mangadexManga, err := client.SearchManga(m.Name)
		if err != nil {
			log.Warn("Failed to search on MangaDex:", err)
		} else if mangadexManga != nil {
			// Basic fields
			if mangadexManga.Description != "" {
				m.Metadata.Summary = mangadexManga.Description
			}
			if len(mangadexManga.Genres) > 0 {
				m.Metadata.Genres = mangadexManga.Genres
			}
			if mangadexManga.Status != "" {
				m.Metadata.Status = mangadexManga.Status
			}

			// Dates
			if mangadexManga.StartDate.Year > 0 {
				m.Metadata.StartDate = model.Date(mangadexManga.StartDate)
			}
			if mangadexManga.EndDate.Year > 0 {
				m.Metadata.EndDate = model.Date(mangadexManga.EndDate)
			}

			// Cover images
			if mangadexManga.CoverImage.ExtraLarge != "" {
				m.Metadata.Cover.ExtraLarge = mangadexManga.CoverImage.ExtraLarge
			}
			if mangadexManga.CoverImage.Large != "" {
				m.Metadata.Cover.Large = mangadexManga.CoverImage.Large
			}
			if mangadexManga.CoverImage.Medium != "" {
				m.Metadata.Cover.Medium = mangadexManga.CoverImage.Medium
			}
			if mangadexManga.CoverImage.Color != "" {
				m.Metadata.Cover.Color = mangadexManga.CoverImage.Color
			}

			// Banner
			if mangadexManga.BannerImage != "" {
				m.Metadata.BannerImage = mangadexManga.BannerImage
			}

			// Numeric fields
			if mangadexManga.Chapters > 0 {
				m.Metadata.Chapters = mangadexManga.Chapters
			}
			if mangadexManga.Format != "" {
				m.Metadata.Format = mangadexManga.Format
			}
			if mangadexManga.Volumes > 0 {
				m.Metadata.Volumes = mangadexManga.Volumes
			}
			if mangadexManga.AverageScore > 0 {
				m.Metadata.AverageScore = mangadexManga.AverageScore
			}
			if mangadexManga.Popularity > 0 {
				m.Metadata.Popularity = mangadexManga.Popularity
			}
			if mangadexManga.MeanScore > 0 {
				m.Metadata.MeanScore = mangadexManga.MeanScore
			}

			// Boolean fields
			m.Metadata.IsLicensed = mangadexManga.IsLicensed

			// Timestamp
			if mangadexManga.UpdatedAt > 0 {
				m.Metadata.UpdatedAt = mangadexManga.UpdatedAt
			}

			// Copy Synonyms
			if len(mangadexManga.Synonyms) > 0 {
				m.Metadata.Synonyms = make([]string, len(mangadexManga.Synonyms))
				copy(m.Metadata.Synonyms, mangadexManga.Synonyms)
			}

			// Add MangaDex URL
			if mangadexManga.SiteURL != "" {
				m.Metadata.URLs = []string{mangadexManga.SiteURL}
			}

			// Extract tags
			if len(mangadexManga.Tags) > 0 {
				for _, tag := range mangadexManga.Tags {
					if tag.Name != "" {
						m.Metadata.Tags = append(m.Metadata.Tags, tag.Name)
					}
				}
			}

			// Extract staff information from Edges
			if len(mangadexManga.Staff.Edges) > 0 {
				for _, edge := range mangadexManga.Staff.Edges {
					if edge.Node.Name.Full == "" {
						continue
					}
					switch edge.Role {
					case "Story":
						m.Metadata.Staff.Story = append(m.Metadata.Staff.Story, edge.Node.Name.Full)
					case "Art":
						m.Metadata.Staff.Art = append(m.Metadata.Staff.Art, edge.Node.Name.Full)
					case "Translation":
						m.Metadata.Staff.Translation = append(m.Metadata.Staff.Translation, edge.Node.Name.Full)
					case "Lettering":
						m.Metadata.Staff.Lettering = append(m.Metadata.Staff.Lettering, edge.Node.Name.Full)
					}
				}
			}

			// Extract character names from Nodes
			if len(mangadexManga.Characters.Nodes) > 0 {
				for _, node := range mangadexManga.Characters.Nodes {
					if node.Name.Full != "" {
						m.Metadata.Characters = append(m.Metadata.Characters, node.Name.Full)
					}
				}
			}

			m.populated = true
		}
	} else if m.Anilist.IsPresent() {
		aniManga := m.Anilist.MustGet()

		// Basic fields
		if aniManga.Description != "" {
			m.Metadata.Summary = aniManga.Description
		}
		if len(aniManga.Genres) > 0 {
			m.Metadata.Genres = aniManga.Genres
		}
		if aniManga.Status != "" {
			m.Metadata.Status = aniManga.Status
		}

		// Dates
		if aniManga.StartDate.Year > 0 {
			m.Metadata.StartDate = model.Date(aniManga.StartDate)
		}
		if aniManga.EndDate.Year > 0 {
			m.Metadata.EndDate = model.Date(aniManga.EndDate)
		}

		// Cover images
		if aniManga.CoverImage.ExtraLarge != "" {
			m.Metadata.Cover.ExtraLarge = aniManga.CoverImage.ExtraLarge
		}
		if aniManga.CoverImage.Large != "" {
			m.Metadata.Cover.Large = aniManga.CoverImage.Large
		}
		if aniManga.CoverImage.Medium != "" {
			m.Metadata.Cover.Medium = aniManga.CoverImage.Medium
		}
		if aniManga.CoverImage.Color != "" {
			m.Metadata.Cover.Color = aniManga.CoverImage.Color
		}

		// Banner
		if aniManga.BannerImage != "" {
			m.Metadata.BannerImage = aniManga.BannerImage
		}

		// Numeric fields
		if aniManga.Chapters > 0 {
			m.Metadata.Chapters = aniManga.Chapters
		}
		if aniManga.Format != "" {
			m.Metadata.Format = aniManga.Format
		}
		if aniManga.Volumes > 0 {
			m.Metadata.Volumes = aniManga.Volumes
		}
		if aniManga.AverageScore > 0 {
			m.Metadata.AverageScore = aniManga.AverageScore
		}
		if aniManga.Popularity > 0 {
			m.Metadata.Popularity = aniManga.Popularity
		}
		if aniManga.MeanScore > 0 {
			m.Metadata.MeanScore = aniManga.MeanScore
		}

		// Boolean fields
		m.Metadata.IsLicensed = aniManga.IsLicensed

		// Timestamp
		if aniManga.UpdatedAt > 0 {
			m.Metadata.UpdatedAt = aniManga.UpdatedAt
		}

		// Copy Synonyms
		if len(aniManga.Synonyms) > 0 {
			m.Metadata.Synonyms = make([]string, len(aniManga.Synonyms))
			copy(m.Metadata.Synonyms, aniManga.Synonyms)
		}

		// Extract URLs
		if aniManga.SiteURL != "" {
			m.Metadata.URLs = make([]string, 0, len(aniManga.External)+1)
			m.Metadata.URLs = append(m.Metadata.URLs, aniManga.SiteURL)
		}
		if len(aniManga.External) > 0 {
			for _, ext := range aniManga.External {
				if ext.URL != "" {
					m.Metadata.URLs = append(m.Metadata.URLs, ext.URL)
				}
			}
		}

		// Extract staff information
		if len(aniManga.Staff.Edges) > 0 {
			for _, edge := range aniManga.Staff.Edges {
				if edge.Node.Name.Full == "" {
					continue
				}
				switch edge.Role {
				case "Story":
					m.Metadata.Staff.Story = append(m.Metadata.Staff.Story, edge.Node.Name.Full)
				case "Art":
					m.Metadata.Staff.Art = append(m.Metadata.Staff.Art, edge.Node.Name.Full)
				}
			}
		}

		// Extract character names
		if len(aniManga.Characters.Nodes) > 0 {
			for _, node := range aniManga.Characters.Nodes {
				if node.Name.Full != "" {
					m.Metadata.Characters = append(m.Metadata.Characters, node.Name.Full)
				}
			}
		}

		// Extract tags
		if len(aniManga.Tags) > 0 {
			for _, tag := range aniManga.Tags {
				if tag.Name != "" {
					m.Metadata.Tags = append(m.Metadata.Tags, tag.Name)
				}
			}
		}

		m.populated = true

		// Get the transformed metadata that matches series.json format
		seriesJSON := m.SeriesJSON()

		// Write debug file if debug flag is set
		if viper.GetBool(key.MetadataDebug) {
			debugDir := filepath.Join(where.Config(), "debug")
			debugFile := filepath.Join(debugDir, "series.json")
			log.Infof("Writing series.json to debug directory: %s", debugFile)
			if data, err := json.MarshalIndent(seriesJSON, "", "  "); err != nil {
				log.Errorf("Failed to marshal debug series.json: %v", err)
			} else {
				if err := os.MkdirAll(debugDir, 0755); err != nil {
					log.Errorf("Failed to create debug directory: %v", err)
				} else {
					if err := os.WriteFile(debugFile, data, 0644); err != nil {
						log.Errorf("Failed to write debug series.json: %v", err)
					} else {
						log.Infof("Successfully wrote debug series.json to %s", debugFile)
					}
				}
			}
		}

		// Save to database with transformed metadata
		dbConn, err := db.GetDB()
		if err != nil {
			log.Error("Failed to get database:", err)
			return err
		}

		// Update metadata to match series.json format
		m.Metadata.Status = seriesJSON.Metadata.Status
		if len(m.Metadata.Staff.Story) > 0 {
			m.Metadata.Publisher = m.Metadata.Staff.Story[0]
		}
		if m.Metadata.Summary != "" {
			m.Metadata.Summary = seriesJSON.Metadata.DescriptionText
		}

		modelManga := &model.Manga{
			ID:          m.ID,
			Title:       m.Name,
			URL:         m.URL,
			Description: m.Metadata.Summary,
			SourceID:    m.Source.ID(),
			SourceName:  m.Source.Name(),
			Metadata:    m.Metadata,
		}

		if err := db.SaveMangaMetadata(dbConn, modelManga); err != nil {
			log.Error("Failed to save manga metadata to database:", err)
			return err
		}

		return nil
	}

	// Try to get metadata from database
	dbConn, err := db.GetDB()
	if err != nil {
		return err
	}

	modelManga := &model.Manga{
		ID:          m.ID,
		Title:       m.Name,
		URL:         m.URL,
		Description: m.Metadata.Summary,
		SourceID:    m.Source.ID(),
		SourceName:  m.Source.Name(),
		Metadata:    m.Metadata,
	}

	// Try to find existing manga metadata by name
	metadata, err := db.SearchMangaByName(dbConn, m.Name)
	if err == nil && metadata != nil {
		modelManga = metadata
		m.Name = metadata.Title
		m.URL = metadata.URL
		m.Metadata = metadata.Metadata
		m.populated = true
		return nil
	}

	// Fall back to JSON file if database lookup fails
	jsonPath := filepath.Join(where.Config(), "metadata", m.ID+".json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Save the default metadata
			if err := db.SaveMangaMetadata(dbConn, modelManga); err != nil {
				log.Error("Failed to save default metadata to database:", err)
				return err
			}
			return m.SaveMetadata()
		}
		return err
	}

	// Parse JSON file
	var jsonManga model.Manga
	if err := json.Unmarshal(data, &jsonManga); err != nil {
		return err
	}

	// Update metadata from JSON
	m.Name = jsonManga.Title
	m.URL = jsonManga.URL
	m.Metadata = jsonManga.Metadata
	m.populated = true

	return nil
}

func (m *Manga) SaveMetadata() error {
	seriesJSON := m.SeriesJSON()

	data, err := json.MarshalIndent(seriesJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Save to JSON file for backward compatibility
	metadataDir := filepath.Join(where.Config(), "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	jsonPath := filepath.Join(metadataDir, m.ID+".json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func (m *Manga) SaveMetadataToFile(path string) error {
	seriesJSON := m.SeriesJSON()

	data, err := json.MarshalIndent(seriesJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Also write to debug directory
	debugFile := "debug/series.json"
	log.Infof("Writing series.json to debug directory: %s", debugFile)
	if err := os.MkdirAll(filepath.Dir(debugFile), 0755); err != nil {
		log.Errorf("Failed to create debug directory: %v", err)
	} else {
		if err := os.WriteFile(debugFile, data, 0644); err != nil {
			log.Errorf("Failed to write debug series.json: %v", err)
		} else {
			log.Infof("Successfully wrote debug series.json to %s", debugFile)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

func (m *Manga) SeriesJSON() *model.SeriesJSON {
	var status string
	switch m.Metadata.Status {
	case "FINISHED":
		status = "Completed"
	case "RELEASING":
		status = "Ongoing"
	case "NOT_YET_RELEASED":
		status = "Unreleased"
	case "CANCELLED":
		status = "Cancelled"
	case "HIATUS":
		status = "Hiatus"
	default:
		status = "Unknown"
	}

	var publisher string
	if len(m.Metadata.Staff.Story) > 0 {
		publisher = m.Metadata.Staff.Story[0]
	}

	totalChapters := len(m.Chapters)
	if m.Metadata.Chapters > 0 {
		totalChapters = m.Metadata.Chapters
	}

	seriesJSON := &model.SeriesJSON{}
	seriesJSON.Metadata.Type = "comicSeries"
	seriesJSON.Metadata.Name = m.Name

	// Handle description
	if m.Metadata.Summary != "" {
		seriesJSON.Metadata.DescriptionFormatted = m.Metadata.Summary
		seriesJSON.Metadata.DescriptionText = m.Metadata.Summary
	}

	// Handle publisher
	if publisher != "" {
		seriesJSON.Metadata.Publisher = publisher
	}

	// Handle status and year
	seriesJSON.Metadata.Status = status
	if m.Metadata.StartDate.Year > 0 {
		seriesJSON.Metadata.Year = m.Metadata.StartDate.Year
	}

	// Handle chapters
	seriesJSON.Metadata.TotalChapters = totalChapters
	seriesJSON.Metadata.TotalIssues = totalChapters
	seriesJSON.Metadata.BookType = "manga"

	// Handle cover image
	if m.Metadata.Cover.ExtraLarge != "" {
		seriesJSON.Metadata.ComicImage = m.Metadata.Cover.ExtraLarge
	} else if m.Metadata.Cover.Large != "" {
		seriesJSON.Metadata.ComicImage = m.Metadata.Cover.Large
	} else if m.Metadata.Cover.Medium != "" {
		seriesJSON.Metadata.ComicImage = m.Metadata.Cover.Medium
	}

	// Format publication run as "MM YYYY - MM YYYY"
	if m.Metadata.StartDate.Year > 0 {
		start := fmt.Sprintf("%d %d", m.Metadata.StartDate.Month, m.Metadata.StartDate.Year)
		end := "0 0"
		if m.Metadata.EndDate.Year > 0 {
			end = fmt.Sprintf("%d %d", m.Metadata.EndDate.Month, m.Metadata.EndDate.Year)
		}
		seriesJSON.Metadata.PublicationRun = fmt.Sprintf("%s - %s", start, end)
	} else {
		seriesJSON.Metadata.PublicationRun = "0 0 - 0 0"
	}

	// Initialize metadata fields
	seriesJSON.Metadata.Genres = make([]string, 0)
	seriesJSON.Metadata.Tags = make([]string, 0)
	seriesJSON.Metadata.Characters = make([]string, 0)
	seriesJSON.Metadata.Staff.Story = make([]string, 0)
	seriesJSON.Metadata.Staff.Art = make([]string, 0)
	seriesJSON.Metadata.Staff.Translation = make([]string, 0)
	seriesJSON.Metadata.Staff.Lettering = make([]string, 0)
	seriesJSON.Metadata.URLs = make([]string, 0)
	seriesJSON.Metadata.Synonyms = make([]string, 0)

	// Copy metadata fields
	if len(m.Metadata.Genres) > 0 {
		seriesJSON.Metadata.Genres = append(seriesJSON.Metadata.Genres, m.Metadata.Genres...)
	}
	if len(m.Metadata.Tags) > 0 {
		seriesJSON.Metadata.Tags = append(seriesJSON.Metadata.Tags, m.Metadata.Tags...)
	}
	if len(m.Metadata.Characters) > 0 {
		seriesJSON.Metadata.Characters = append(seriesJSON.Metadata.Characters, m.Metadata.Characters...)
	}
	if len(m.Metadata.Staff.Story) > 0 {
		seriesJSON.Metadata.Staff.Story = append(seriesJSON.Metadata.Staff.Story, m.Metadata.Staff.Story...)
	}
	if len(m.Metadata.Staff.Art) > 0 {
		seriesJSON.Metadata.Staff.Art = append(seriesJSON.Metadata.Staff.Art, m.Metadata.Staff.Art...)
	}
	if len(m.Metadata.Staff.Translation) > 0 {
		seriesJSON.Metadata.Staff.Translation = append(seriesJSON.Metadata.Staff.Translation, m.Metadata.Staff.Translation...)
	}
	if len(m.Metadata.Staff.Lettering) > 0 {
		seriesJSON.Metadata.Staff.Lettering = append(seriesJSON.Metadata.Staff.Lettering, m.Metadata.Staff.Lettering...)
	}
	if len(m.Metadata.URLs) > 0 {
		seriesJSON.Metadata.URLs = append(seriesJSON.Metadata.URLs, m.Metadata.URLs...)
	}
	if len(m.Metadata.Synonyms) > 0 {
		seriesJSON.Metadata.Synonyms = append(seriesJSON.Metadata.Synonyms, m.Metadata.Synonyms...)
	}

	// Handle numeric fields
	if m.Metadata.Volumes > 0 {
		seriesJSON.Metadata.Volumes = m.Metadata.Volumes
	}
	if m.Metadata.Chapters > 0 {
		seriesJSON.Metadata.Chapters = m.Metadata.Chapters
	}
	if m.Metadata.AverageScore > 0 {
		seriesJSON.Metadata.AverageScore = m.Metadata.AverageScore
	}
	if m.Metadata.Popularity > 0 {
		seriesJSON.Metadata.Popularity = m.Metadata.Popularity
	}
	if m.Metadata.MeanScore > 0 {
		seriesJSON.Metadata.MeanScore = m.Metadata.MeanScore
	}

	// Handle boolean and timestamp fields
	seriesJSON.Metadata.IsLicensed = m.Metadata.IsLicensed
	if m.Metadata.UpdatedAt > 0 {
		seriesJSON.Metadata.UpdatedAt = m.Metadata.UpdatedAt
	}

	// Handle image fields
	if m.Metadata.BannerImage != "" {
		seriesJSON.Metadata.BannerImage = m.Metadata.BannerImage
	}
	seriesJSON.Metadata.Cover = m.Metadata.Cover

	// Handle format
	if m.Metadata.Format != "" {
		seriesJSON.Metadata.Format = m.Metadata.Format
	} else {
		seriesJSON.Metadata.Format = "MANGA"
	}

	// Always add the manga URL to the URLs list if not already present
	if m.URL != "" {
		found := false
		for _, url := range seriesJSON.Metadata.URLs {
			if url == m.URL {
				found = true
				break
			}
		}
		if !found {
			seriesJSON.Metadata.URLs = append(seriesJSON.Metadata.URLs, m.URL)
		}
	}

	return seriesJSON
}
