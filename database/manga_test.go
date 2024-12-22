package database

import (
	"testing"

	"github.com/metafates/mangal/db"
	"github.com/metafates/mangal/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSaveMangaMetadata(t *testing.T) {
	Convey("Given a test database setup", t, func() {
		database, err := db.GetDB()
		So(err, ShouldBeNil)
		defer database.Close()

		err = db.Migrate(database, true)
		So(err, ShouldBeNil)

		Convey("When saving manga metadata", func() {
			manga := &model.Manga{
				Title:       "Test Manga",
				Description: "Test Description",
				Metadata: model.MangaMetadata{
					Publisher:  "Test Publisher",
					Status:    "Ongoing",
					StartDate: model.Date{Year: 2024},
					Chapters:  10,
					Cover: model.Cover{
						ExtraLarge: "https://example.com/cover-xl.jpg",
						Large:     "https://example.com/cover-l.jpg",
						Medium:    "https://example.com/cover-m.jpg",
						Color:     "#FF0000",
					},
					URLs: []string{"https://example.com/manga/1"},
					Genres: []string{"Action", "Adventure"},
					Tags: []string{"Magic", "Fantasy"},
					Characters: []string{"Hero", "Villain"},
					Volumes: 2,
					AverageScore: 85,
					Popularity: 1000,
					MeanScore: 80,
					IsLicensed: true,
					BannerImage: "https://example.com/banner.jpg",
					Format: "MANGA",
				},
			}

			err := db.SaveMangaMetadata(database, manga)

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)

				Convey("And we should be able to retrieve it", func() {
					retrieved, err := db.SearchMangaByName(database, "Test Manga")
					So(err, ShouldBeNil)
					So(retrieved, ShouldNotBeNil)
					So(retrieved.Title, ShouldEqual, manga.Title)
					So(retrieved.Description, ShouldEqual, manga.Description)
					So(retrieved.Metadata.Publisher, ShouldEqual, manga.Metadata.Publisher)
					So(retrieved.Metadata.Status, ShouldEqual, manga.Metadata.Status)
					So(retrieved.Metadata.StartDate.Year, ShouldEqual, manga.Metadata.StartDate.Year)
					So(retrieved.Metadata.Chapters, ShouldEqual, manga.Metadata.Chapters)
					So(retrieved.Metadata.Cover.ExtraLarge, ShouldEqual, manga.Metadata.Cover.ExtraLarge)
					So(retrieved.Metadata.Cover.Large, ShouldEqual, manga.Metadata.Cover.Large)
					So(retrieved.Metadata.Cover.Medium, ShouldEqual, manga.Metadata.Cover.Medium)
					So(retrieved.Metadata.Cover.Color, ShouldEqual, manga.Metadata.Cover.Color)
					So(retrieved.Metadata.URLs, ShouldResemble, manga.Metadata.URLs)
					So(retrieved.Metadata.Genres, ShouldResemble, manga.Metadata.Genres)
					So(retrieved.Metadata.Tags, ShouldResemble, manga.Metadata.Tags)
					So(retrieved.Metadata.Characters, ShouldResemble, manga.Metadata.Characters)
					So(retrieved.Metadata.Volumes, ShouldEqual, manga.Metadata.Volumes)
					So(retrieved.Metadata.AverageScore, ShouldEqual, manga.Metadata.AverageScore)
					So(retrieved.Metadata.Popularity, ShouldEqual, manga.Metadata.Popularity)
					So(retrieved.Metadata.MeanScore, ShouldEqual, manga.Metadata.MeanScore)
					So(retrieved.Metadata.IsLicensed, ShouldEqual, manga.Metadata.IsLicensed)
					So(retrieved.Metadata.BannerImage, ShouldEqual, manga.Metadata.BannerImage)
					So(retrieved.Metadata.Format, ShouldEqual, manga.Metadata.Format)
				})
			})
		})
	})
}

func TestSearchMangaByName(t *testing.T) {
	Convey("Given a test database with a manga", t, func() {
		database, err := db.GetDB()
		So(err, ShouldBeNil)
		defer database.Close()

		err = db.Migrate(database, true)
		So(err, ShouldBeNil)

		testManga := &model.Manga{
			Title:       "Test Manga",
			Description: "Test Description",
			Metadata: model.MangaMetadata{
				Publisher:  "Test Publisher",
				Status:    "Ongoing",
				StartDate: model.Date{Year: 2024},
				Chapters:  10,
				Cover: model.Cover{
					ExtraLarge: "https://example.com/cover-xl.jpg",
					Large:     "https://example.com/cover-l.jpg",
					Medium:    "https://example.com/cover-m.jpg",
					Color:     "#FF0000",
				},
				URLs: []string{"https://example.com/manga/1"},
				Genres: []string{"Action", "Adventure"},
				Tags: []string{"Magic", "Fantasy"},
				Characters: []string{"Hero", "Villain"},
				Volumes: 2,
				AverageScore: 85,
				Popularity: 1000,
				MeanScore: 80,
				IsLicensed: true,
				BannerImage: "https://example.com/banner.jpg",
				Format: "MANGA",
			},
		}
		err = db.SaveMangaMetadata(database, testManga)
		So(err, ShouldBeNil)

		Convey("When searching with exact match", func() {
			manga, err := db.SearchMangaByName(database, "Test Manga")

			Convey("Then it should find the manga", func() {
				So(err, ShouldBeNil)
				So(manga, ShouldNotBeNil)
				So(manga.Title, ShouldEqual, "Test Manga")
			})
		})

		Convey("When searching with partial match", func() {
			manga, err := db.SearchMangaByName(database, "Test")

			Convey("Then it should find the manga", func() {
				So(err, ShouldBeNil)
				So(manga, ShouldNotBeNil)
				So(manga.Title, ShouldEqual, "Test Manga")
			})
		})

		Convey("When searching with case insensitive match", func() {
			manga, err := db.SearchMangaByName(database, "test manga")

			Convey("Then it should find the manga", func() {
				So(err, ShouldBeNil)
				So(manga, ShouldNotBeNil)
				So(manga.Title, ShouldEqual, "Test Manga")
			})
		})

		Convey("When searching for nonexistent manga", func() {
			manga, err := db.SearchMangaByName(database, "Nonexistent Manga")

			Convey("Then it should return nil without error", func() {
				So(err, ShouldBeNil)
				So(manga, ShouldBeNil)
			})
		})
	})
}
