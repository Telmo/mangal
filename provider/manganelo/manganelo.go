package manganelo

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/metafates/mangal/provider/generic"
	"net/url"
	"strings"
	"time"
)

var Config = &generic.Configuration{
	Name:            "Manganelo",
	Delay:           50 * time.Millisecond,
	Parallelism:     50,
	ReverseChapters: true,
	BaseURL:         "https://chapmanganato.com/",
	GenerateSearchURL: func(query string) string {
		query = strings.TrimSpace(query)
		query = strings.ToLower(query)
		query = url.QueryEscape(query)
		template := "https://chapmanganato.com/search/%s"
		return fmt.Sprintf(template, query)
	},
	MangaExtractor: &generic.Extractor{
		Selector: ".panel-search-story .search-story-item",
		Name: func(selection *goquery.Selection) string {
			return selection.Find(".item-title").Text()
		},
		URL: func(selection *goquery.Selection) string {
			return selection.Find(".item-title").AttrOr("href", "")
		},
		Cover: func(selection *goquery.Selection) string {
			return selection.Find(".item-img").AttrOr("src", "")
		},
	},
	ChapterExtractor: &generic.Extractor{
		Selector: ".chapter-list .chapter-item",
		Name: func(selection *goquery.Selection) string {
			name := selection.Find(".chapter-name").Text()
			if strings.HasPrefix(name, "Vol.") {
				splitted := strings.Split(name, " ")
				name = strings.Join(splitted[1:], " ")
			}
			return name
		},
		URL: func(selection *goquery.Selection) string {
			return selection.Find(".chapter-name").AttrOr("href", "")
		},
		Volume: func(selection *goquery.Selection) string {
			name := selection.Find(".chapter-name").Text()
			if strings.HasPrefix(name, "Vol.") {
				splitted := strings.Split(name, " ")
				return splitted[0]
			}
			return ""
		},
	},
	PageExtractor: &generic.Extractor{
		Selector: ".container-chapter-reader img",
		Name:     nil,
		URL: func(selection *goquery.Selection) string {
			if src := selection.AttrOr("src", ""); src != "" {
				return src
			}
			return selection.AttrOr("data-src", "")
		},
	},
}
