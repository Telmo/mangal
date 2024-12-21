package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/metafates/mangal/config"
	"github.com/metafates/mangal/constant"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/source"
	"github.com/samber/lo"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func init() {
	filesystem.SetMemMapFs()
	lo.Must0(config.Setup())
	viper.Set(key.FormatsUse, constant.FormatPDF)
}

func SampleChapter() (*source.Chapter, error) {
	tmpDir, err := os.MkdirTemp("", "mangal-pdf-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	pages := make([]*source.Page, 2)
	for i := range pages {
		// Create a new PNG image with a gradient
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		for y := 0; y < 100; y++ {
			for x := 0; x < 100; x++ {
				c := color.RGBA{
					R: uint8((x * 255) / 100),
					G: uint8((y * 255) / 100),
					B: 0,
					A: 255,
				}
				img.Set(x, y, c)
			}
		}

		// Create a buffer to store the PNG image
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode PNG: %w", err)
		}

		// Write the PNG image to a temporary file
		pagePath := filepath.Join(tmpDir, fmt.Sprintf("page-%d.png", i+1))
		if err := os.WriteFile(pagePath, buf.Bytes(), 0644); err != nil {
			return nil, fmt.Errorf("failed to write PNG file: %w", err)
		}

		// Create a new buffer for the page contents
		contentsBuf := bytes.NewBuffer(buf.Bytes())

		pages[i] = &source.Page{
			URL:       pagePath,
			Index:     uint16(i),
			Extension: "png",
			Contents:  contentsBuf,
		}
	}

	manga := &source.Manga{
		Name:     "Test Manga",
		URL:      "https://example.com/test-manga",
		ID:       "test-manga",
		Index:    1,
		Chapters: make([]*source.Chapter, 0),
	}

	chapter := &source.Chapter{
		Name:   "Test Chapter",
		URL:    "https://example.com/test-chapter",
		Index:  1,
		ID:     "test-chapter",
		Pages:  pages,
		Manga:  manga,
		Volume: "1",
	}

	manga.Chapters = append(manga.Chapters, chapter)

	return chapter, nil
}

func TestPDF(t *testing.T) {
	Convey("Given a chapter", t, func() {
		chapter, err := SampleChapter()
		So(err, ShouldBeNil)

		Convey("When converting to PDF", func() {
			pdf := &PDF{}
			result, err := pdf.Save(chapter)
			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And the PDF file should exist", func() {
					_, err := os.Stat(result)
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
