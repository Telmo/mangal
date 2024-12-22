package model

import "strings"

// Date represents a date with year, month and day
type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

// MangaMetadata contains metadata about a manga series
type MangaMetadata struct {
	// Genres of the manga
	Genres []string `json:"genres"`
	// Summary in plain text with newlines
	Summary string `json:"summary"`
	// Staff that worked on the manga
	Staff struct {
		// Story authors
		Story []string `json:"story"`
		// Art authors
		Art []string `json:"art"`
		// Translation group
		Translation []string `json:"translation"`
		// Lettering group
		Lettering []string `json:"lettering"`
	} `json:"staff"`
	// Cover images of the manga
	Cover struct {
		// ExtraLarge is the largest cover image
		ExtraLarge string `json:"extraLarge"`
		// Large is the second-largest cover image
		Large string `json:"large"`
		// Medium cover image
		Medium string `json:"medium"`
		// Color average color of the cover image
		Color string `json:"color"`
	} `json:"cover"`
	// BannerImage is the banner image of the manga
	BannerImage string `json:"bannerImage"`
	// Tags of the manga
	Tags []string `json:"tags"`
	// Characters of the manga
	Characters []string `json:"characters"`
	// Status of the manga
	Status string `json:"status"`
	// StartDate is the date when the manga started
	StartDate Date `json:"startDate"`
	// EndDate is the date when the manga ended
	EndDate Date `json:"endDate"`
	// Synonyms other names of the manga
	Synonyms []string `json:"synonyms"`
	// Chapters is the amount of chapters the manga will have when completed
	Chapters int `json:"chapters"`
	// URLs external URLs of the manga
	URLs []string `json:"urls"`
	// Publisher of the manga
	Publisher string `json:"publisher"`
	// Format of the manga (MANGA, NOVEL, ONE_SHOT, etc.)
	Format string `json:"format"`
	// Volumes is the amount of volumes the manga has when complete
	Volumes int `json:"volumes"`
	// AverageScore is the average score of the manga
	AverageScore int `json:"averageScore"`
	// Popularity is the number of users who have the manga in their list
	Popularity int `json:"popularity"`
	// MeanScore is the mean score of the manga
	MeanScore int `json:"meanScore"`
	// IsLicensed indicates if the manga is licensed
	IsLicensed bool `json:"isLicensed"`
	// UpdatedAt is the timestamp of when the manga was last updated
	UpdatedAt int `json:"updatedAt"`
	// PublicationRun is the publication run of the manga (e.g. "1 2023 - 4 2024")
	PublicationRun string `json:"publicationRun"`
}

// SeriesJSON represents metadata in series.json format
type SeriesJSON struct {
	Metadata struct {
		Type                 string   `json:"type"`
		Name                 string   `json:"name"`
		DescriptionFormatted string   `json:"descriptionFormatted"`
		DescriptionText      string   `json:"descriptionText"`
		Publisher           string   `json:"publisher"`
		Status              string   `json:"status"`
		Year                int      `json:"year"`
		TotalChapters       int      `json:"totalChapters"`
		TotalIssues         int      `json:"totalIssues"`
		BookType            string   `json:"bookType"`
		ComicImage          string   `json:"comicImage"`
		ComicID             int      `json:"comicID"`
		PublicationRun      string   `json:"publicationRun"`
		Genres              []string `json:"genres"`
		Tags                []string `json:"tags"`
		Characters          []string `json:"characters"`
		Staff               struct {
			Story       []string `json:"story"`
			Art         []string `json:"art"`
			Translation []string `json:"translation"`
			Lettering   []string `json:"lettering"`
		} `json:"staff"`
		Volumes       int      `json:"volumes"`
		Chapters      int      `json:"chapters"`
		AverageScore  int      `json:"averageScore"`
		Popularity    int      `json:"popularity"`
		MeanScore     int      `json:"meanScore"`
		IsLicensed    bool     `json:"isLicensed"`
		UpdatedAt     int      `json:"updatedAt"`
		URLs          []string `json:"urls"`
		BannerImage   string   `json:"bannerImage"`
		Cover         struct {
			ExtraLarge string `json:"extraLarge"`
			Large      string `json:"large"`
			Medium     string `json:"medium"`
			Color      string `json:"color"`
		} `json:"cover"`
		Synonyms      []string `json:"synonyms"`
		Format        string   `json:"format"`
	} `json:"metadata"`
}

// SeriesJSONToManga converts a SeriesJSON to a Manga
func SeriesJSONToManga(seriesJSON *SeriesJSON) *Manga {
	// Convert format to lowercase and handle empty case
	format := strings.ToLower(seriesJSON.Metadata.Format)
	if format == "" {
		format = "manga"
	}

	return &Manga{
		Title:       seriesJSON.Metadata.Name,
		Description: seriesJSON.Metadata.DescriptionFormatted,
		Metadata: MangaMetadata{
			Status: seriesJSON.Metadata.Status,
			StartDate: Date{
				Year: seriesJSON.Metadata.Year,
			},
			Summary:      seriesJSON.Metadata.DescriptionText,
			Genres:      seriesJSON.Metadata.Genres,
			Tags:        seriesJSON.Metadata.Tags,
			Characters:  seriesJSON.Metadata.Characters,
			Staff:       seriesJSON.Metadata.Staff,
			Volumes:     seriesJSON.Metadata.Volumes,
			Chapters:    seriesJSON.Metadata.Chapters,
			AverageScore: seriesJSON.Metadata.AverageScore,
			Popularity:   seriesJSON.Metadata.Popularity,
			MeanScore:    seriesJSON.Metadata.MeanScore,
			IsLicensed:   seriesJSON.Metadata.IsLicensed,
			BannerImage:  seriesJSON.Metadata.BannerImage,
			URLs:         seriesJSON.Metadata.URLs,
			Cover: struct {
				ExtraLarge string `json:"extraLarge"`
				Large      string `json:"large"`
				Medium     string `json:"medium"`
				Color      string `json:"color"`
			}{
				ExtraLarge: seriesJSON.Metadata.ComicImage,
				Large:      seriesJSON.Metadata.ComicImage,
				Medium:     seriesJSON.Metadata.ComicImage,
			},
			PublicationRun: seriesJSON.Metadata.PublicationRun,
			Format:         format,
		},
	}
}

// Manga represents a manga series
type Manga struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Authors     []string `json:"authors"`
	URL         string   `json:"url"`
	Cover       string   `json:"cover"`
	SourceID    string   `json:"source_id"`
	SourceName  string   `json:"source_name"`
	Genres      []string `json:"genres"`
	Staff       []string `json:"staff"`
	Metadata    MangaMetadata `json:"metadata"`
}

// Chapter represents a chapter of a manga
type Chapter struct {
	// Name of the chapter
	Name string `json:"name"`
	// URL of the chapter
	URL string `json:"url"`
	// Index of the chapter in the manga
	Index uint16 `json:"index"`
	// ID of the chapter in the source
	ID string `json:"id"`
	// Volume which the chapter belongs to
	Volume string `json:"volume"`
	// Pages of the chapter
	Pages []*Page `json:"pages"`
}

// Page represents a page in a chapter
type Page struct {
	// URL of the page. Used to download the page.
	URL string `json:"url"`
	// Index of the page in the chapter
	Index uint16 `json:"index"`
	// Extension of the page image
	Extension string `json:"extension"`
	// Size of the page in bytes
	Size uint64 `json:"-"`
}
