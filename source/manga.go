package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/metafates/mangal/anilist"
	"github.com/metafates/mangal/database"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/log"
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

	// Try to bind with Anilist first
	if err := m.BindWithAnilist(); err != nil {
		log.Warn("Failed to bind with Anilist:", err)
	} else if m.Anilist.IsPresent() {
		aniManga := m.Anilist.MustGet()
		m.Metadata.Summary = aniManga.Description
		m.Metadata.Genres = aniManga.Genres
		m.Metadata.Status = aniManga.Status
		m.Metadata.StartDate = model.Date(aniManga.StartDate)
		m.Metadata.EndDate = model.Date(aniManga.EndDate)
		m.Metadata.Cover.ExtraLarge = aniManga.CoverImage.ExtraLarge
		m.Metadata.Cover.Large = aniManga.CoverImage.Large
		m.Metadata.Cover.Medium = aniManga.CoverImage.Medium
		m.Metadata.Cover.Color = aniManga.CoverImage.Color
		m.Metadata.BannerImage = aniManga.BannerImage
		m.Metadata.Chapters = aniManga.Chapters
		m.Metadata.Synonyms = aniManga.Synonyms

		// Extract staff information
		for _, edge := range aniManga.Staff.Edges {
			switch edge.Role {
			case "Story":
				m.Metadata.Staff.Story = append(m.Metadata.Staff.Story, edge.Node.Name.Full)
			case "Art":
				m.Metadata.Staff.Art = append(m.Metadata.Staff.Art, edge.Node.Name.Full)
			}
		}

		// Extract character names
		for _, node := range aniManga.Characters.Nodes {
			m.Metadata.Characters = append(m.Metadata.Characters, node.Name.Full)
		}

		// Extract tags
		for _, tag := range aniManga.Tags {
			m.Metadata.Tags = append(m.Metadata.Tags, tag.Name)
		}

		m.populated = true

		// Save to database
		db, err := database.GetDB()
		if err != nil {
			log.Error("Failed to get database:", err)
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

		if err := database.SaveMangaMetadata(db, modelManga); err != nil {
			log.Error("Failed to save metadata to database:", err)
			return err
		}

		// Save to JSON file as backup
		return m.SaveMetadata()
	}

	// Try to get metadata from database
	db, err := database.GetDB()
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

	metadata, err := database.GetMangaMetadata(db, m.ID)
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
			return nil
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	if err := json.Unmarshal(data, modelManga); err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	m.Name = modelManga.Title
	m.URL = modelManga.URL
	m.Metadata = modelManga.Metadata
	m.populated = true

	return nil
}

func (m *Manga) SaveMetadata() error {
	modelManga := &model.Manga{
		ID:          m.ID,
		Title:       m.Name,
		URL:         m.URL,
		Description: m.Metadata.Summary,
		SourceID:    m.Source.ID(),
		SourceName:  m.Source.Name(),
		Metadata:    m.Metadata,
	}

	// Save to JSON file for backward compatibility
	metadataDir := filepath.Join(where.Config(), "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	jsonPath := filepath.Join(metadataDir, m.ID+".json")
	data, err := json.MarshalIndent(modelManga, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
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
		status = "Continuing"
	default:
		status = "Unknown"
	}

	var publisher string
	if len(m.Metadata.Staff.Story) > 0 {
		publisher = m.Metadata.Staff.Story[0]
	}

	seriesJSON := &model.SeriesJSON{}
	seriesJSON.Metadata.Type = "comicSeries"
	seriesJSON.Metadata.Name = m.Name
	seriesJSON.Metadata.DescriptionFormatted = m.Metadata.Summary
	seriesJSON.Metadata.DescriptionText = m.Metadata.Summary
	seriesJSON.Metadata.Publisher = publisher
	seriesJSON.Metadata.Status = status
	seriesJSON.Metadata.Year = m.Metadata.StartDate.Year
	seriesJSON.Metadata.TotalChapters = len(m.Chapters)
	seriesJSON.Metadata.TotalIssues = len(m.Chapters)
	seriesJSON.Metadata.BookType = "manga"
	seriesJSON.Metadata.ComicImage = m.Metadata.Cover.ExtraLarge
	seriesJSON.Metadata.ComicID = 0

	// Format publication run as "MM YYYY - MM YYYY"
	start := fmt.Sprintf("%d %d", m.Metadata.StartDate.Month, m.Metadata.StartDate.Year)
	end := "0 0"
	if m.Metadata.EndDate.Year > 0 {
		end = fmt.Sprintf("%d %d", m.Metadata.EndDate.Month, m.Metadata.EndDate.Year)
	}
	seriesJSON.Metadata.PublicationRun = fmt.Sprintf("%s - %s", start, end)

	return seriesJSON
}
