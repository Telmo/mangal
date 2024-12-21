package mangadex

import (
	"fmt"
	"strings"
	"time"

	"github.com/metafates/mangal/anilist"
)

type MangaDexResponse struct {
	Result   string        `json:"result"`
	Response string        `json:"response"`
	Data     []MangaDexManga `json:"data"`
	Limit    int          `json:"limit"`
	Offset   int          `json:"offset"`
	Total    int          `json:"total"`
}

type MangaDexManga struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Title       map[string]string   `json:"title"`
		AltTitles   []map[string]string `json:"altTitles"`
		Description map[string]string   `json:"description"`
		Status      string             `json:"status"`
		Year        *int               `json:"year"`
		Tags        []struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				Name  map[string]string `json:"name"`
				Group string           `json:"group"`
			} `json:"attributes"`
		} `json:"tags"`
		ContentRating    string `json:"contentRating"`
		OriginalLanguage string `json:"originalLanguage"`
		LastVolume       string `json:"lastVolume"`
		LastChapter      string `json:"lastChapter"`
		PublicationDemographic string `json:"publicationDemographic"`
		CreatedAt       time.Time `json:"createdAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
	} `json:"attributes"`
	Relationships []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Attributes struct {
			FileName string `json:"fileName,omitempty"`
			Name     string `json:"name,omitempty"`
		} `json:"attributes,omitempty"`
	} `json:"relationships"`
}

// ToAnilist converts a MangaDex manga to Anilist format
func (m *MangaDexManga) ToAnilist() *anilist.Manga {
	manga := &anilist.Manga{}

	// Set titles
	manga.Title.English = m.Attributes.Title["en"]
	manga.Title.Romaji = m.Attributes.Title["en"] // MangaDex doesn't have romaji titles
	manga.Title.Native = m.Attributes.Title["ja"]

	// Convert ID to int
	// Note: MangaDex uses UUID strings, we'll use a hash of it for compatibility
	manga.ID = int(hash(m.ID))

	// Set description (prefer English)
	if desc, ok := m.Attributes.Description["en"]; ok {
		manga.Description = desc
	} else {
		// Take first available description
		for _, desc := range m.Attributes.Description {
			manga.Description = desc
			break
		}
	}

	// Set cover image
	var coverFileName string
	for _, rel := range m.Relationships {
		if rel.Type == "cover_art" && rel.Attributes.FileName != "" {
			coverFileName = rel.Attributes.FileName
			break
		}
	}
	if coverFileName != "" {
		coverURL := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", m.ID, coverFileName)
		manga.CoverImage.ExtraLarge = coverURL
		manga.CoverImage.Large = coverURL
		manga.CoverImage.Medium = coverURL
	}

	// Set tags
	for _, tag := range m.Attributes.Tags {
		if name, ok := tag.Attributes.Name["en"]; ok {
			manga.Tags = append(manga.Tags, struct {
				Name        string `json:"name" jsonschema:"description=Name of the tag."`
				Description string `json:"description" jsonschema:"description=Description of the tag."`
				Rank        int    `json:"rank" jsonschema:"description=Rank of the tag. How relevant it is to the manga from 1 to 100."`
			}{
				Name:        name,
				Description: "",
				Rank:        50, // Default rank since MangaDex doesn't provide this
			})
		}
	}

	// Set status
	switch m.Attributes.Status {
	case "completed":
		manga.Status = "FINISHED"
	case "ongoing":
		manga.Status = "RELEASING"
	case "hiatus":
		manga.Status = "HIATUS"
	case "cancelled":
		manga.Status = "CANCELLED"
	default:
		manga.Status = "NOT_YET_RELEASED"
	}

	// Set start date
	if m.Attributes.Year != nil {
		manga.StartDate.Year = *m.Attributes.Year
	}

	// Set synonyms from alt titles
	for _, altTitles := range m.Attributes.AltTitles {
		for _, title := range altTitles {
			if title != "" && title != manga.Title.English && title != manga.Title.Native {
				manga.Synonyms = append(manga.Synonyms, title)
			}
		}
	}

	// Set country of origin
	manga.Country = strings.ToUpper(m.Attributes.OriginalLanguage)

	// Set site URL
	manga.SiteURL = fmt.Sprintf("https://mangadex.org/title/%s", m.ID)

	// Set format
	manga.Format = "MANGA" // MangaDex primarily hosts manga

	return manga
}

// hash generates a simple hash from a string
func hash(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return h
}
