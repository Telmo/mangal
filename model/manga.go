package model

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
}

// SeriesJSON represents metadata in series.json format
type SeriesJSON struct {
	Metadata struct {
		Type                 string `json:"type"`
		Name                 string `json:"name"`
		DescriptionFormatted string `json:"description_formatted"`
		DescriptionText      string `json:"description_text"`
		Publisher           string `json:"publisher"`
		Status              string `json:"status"`
		Year                int    `json:"year"`
		TotalChapters       int    `json:"total_chapters"`
		TotalIssues        int    `json:"total_issues"`
		BookType           string `json:"booktype"`
		ComicImage         string `json:"ComicImage"`
		ComicID            int    `json:"comicId"`
		PublicationRun     string `json:"publication_run"`
	} `json:"metadata"`
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
