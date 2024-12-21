package mangadex

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/darylhjd/mangodex"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/model"
	"github.com/metafates/mangal/source"
	"github.com/spf13/viper"
)

func (m *Mangadex) Search(query string) ([]*source.Manga, error) {
	if cached, ok := m.cache.mangas.Get(query).Get(); ok {
		for _, manga := range cached {
			manga.Source = m
		}

		return cached, nil
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(100))

	ratings := []string{mangodex.Safe, mangodex.Suggestive}

	for _, rating := range ratings {
		params.Add("contentRating[]", rating)
	}

	if viper.GetBool(key.MangadexNSFW) {
		params.Add("contentRating[]", mangodex.Porn)
		params.Add("contentRating[]", mangodex.Erotica)
	}

	// Change Mangadex provider sorting to show more relevant results #204
	// TODO: Test this to make sure it does not affect my downloads
	// 	- params.Set("order[followedCount]", "desc")
	//  + params.Set("order[relevance]", "desc")
	params.Set("order[followedCount]", "desc")
	params.Set("title", query)

	mangaList, err := m.client.Manga.GetMangaList(params)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	var mangas []*source.Manga

	for i, manga := range mangaList.Data {
		// Convert tags to string slice
		tags := make([]string, 0, len(manga.Attributes.Tags))
		for _, tag := range manga.Attributes.Tags {
			name := tag.Attributes.Name.GetLocalString("en")
			if name != "" {
				tags = append(tags, name)
			}
		}

		// Get cover art
		var coverImage struct {
			ExtraLarge string `json:"extraLarge"`
			Large      string `json:"large"`
			Medium     string `json:"medium"`
			Color      string `json:"color"`
		}
		for _, rel := range manga.Relationships {
			if rel.Type == "cover_art" {
				if attrs, ok := rel.Attributes.(map[string]interface{}); ok {
					if fileName, ok := attrs["fileName"].(string); ok {
						coverImage.ExtraLarge = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.jpg.512.jpg", manga.ID, fileName)
						coverImage.Large = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.jpg.256.jpg", manga.ID, fileName)
						coverImage.Medium = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.jpg.128.jpg", manga.ID, fileName)
					}
					if color, ok := attrs["color"].(string); ok {
						coverImage.Color = color
					}
				}
				break
			}
		}

		// Convert status
		var status string
		if manga.Attributes.Status != nil {
			switch *manga.Attributes.Status {
			case "completed":
				status = "FINISHED"
			case "ongoing":
				status = "RELEASING"
			case "cancelled":
				status = "CANCELLED"
			case "hiatus":
				status = "HIATUS"
			default:
				status = "UNKNOWN"
			}
		} else {
			status = "UNKNOWN"
		}

		// Get description
		description := manga.GetDescription("en")

		// Convert year
		var year int
		if manga.Attributes.Year != nil {
			year = *manga.Attributes.Year
		}

		// Convert chapters and volumes
		var chapters, volumes int
		if manga.Attributes.LastChapter != nil {
			if val, err := strconv.Atoi(*manga.Attributes.LastChapter); err == nil {
				chapters = val
			}
		}
		if manga.Attributes.LastVolume != nil {
			if val, err := strconv.Atoi(*manga.Attributes.LastVolume); err == nil {
				volumes = val
			}
		}

		m := source.Manga{
			Name:   manga.GetTitle(viper.GetString(key.MangadexLanguage)),
			URL:    fmt.Sprintf("https://mangadex.org/title/%s", manga.ID),
			Index:  uint16(i),
			ID:     manga.ID,
			Source: m,
			Metadata: model.MangaMetadata{
				Genres:  tags, // Using tags as genres since MangaDex doesn't separate them
				Summary: description,
				Status:  status,
				Format:  "MANGA",
				StartDate: model.Date{
					Year:  year,
					Month: 1, // MangaDex doesn't provide month/day
					Day:   1,
				},
				EndDate: model.Date{
					Year:  0, // MangaDex doesn't provide end date
					Month: 0,
					Day:   0,
				},
				Chapters:     chapters,
				Volumes:      volumes,
				AverageScore: 0, // MangaDex doesn't provide this
				MeanScore:    0, // MangaDex doesn't provide this
				Popularity:   0, // MangaDex doesn't provide this
				IsLicensed:   false, // MangaDex doesn't provide this info
				Cover: struct {
					ExtraLarge string `json:"extraLarge"`
					Large      string `json:"large"`
					Medium     string `json:"medium"`
					Color      string `json:"color"`
				}{
					ExtraLarge: coverImage.ExtraLarge,
					Large:      coverImage.Large,
					Medium:     coverImage.Medium,
					Color:      coverImage.Color,
				},
				BannerImage:  "", // MangaDex doesn't provide banner images
				Tags:         tags,
				Characters:   []string{}, // MangaDex doesn't provide character info
				Staff: struct {
					Story       []string `json:"story"`
					Art         []string `json:"art"`
					Translation []string `json:"translation"`
					Lettering   []string `json:"lettering"`
				}{
					Story:       []string{}, // MangaDex doesn't provide this
					Art:         []string{}, // MangaDex doesn't provide this
					Translation: []string{},
					Lettering:   []string{},
				},
				URLs: []string{fmt.Sprintf("https://mangadex.org/title/%s", manga.ID)},
			},
		}

		mangas = append(mangas, &m)
	}

	_ = m.cache.mangas.Set(query, mangas)
	return mangas, nil
}
