package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/metafates/mangal/color"
	"github.com/metafates/mangal/converter"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/history"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/style"
	"github.com/spf13/viper"
)

// Download the chapter using given source.
func Download(chapter *source.Chapter, progress func(string)) (string, error) {
	path, err := chapter.Path(false)
	if err != nil {
		return "", fmt.Errorf("failed to get chapter path: %w", err)
	}

	log.Info("downloading " + chapter.Manga.Name + " - " + chapter.Name + " to: " + path)

	if viper.GetBool(key.DownloaderRedownloadExisting) {
		log.Info("chapter already downloaded, deleting and redownloading")
		if err := filesystem.Api().Remove(path); err != nil {
			log.Warn("failed to delete existing chapter: " + err.Error())
		}
	} else if chapter.IsDownloaded() {
		log.Info("chapter already downloaded, skipping")
		return path, nil
	}

	progress("Getting pages")
	pages, err := chapter.Source().PagesOf(chapter)
	if err != nil {
		return "", fmt.Errorf("failed to get pages: %w", err)
	}
	log.Info(fmt.Sprintf("found %d pages", len(pages)))

	if err := chapter.DownloadPages(false, progress); err != nil {
		return "", fmt.Errorf("failed to download pages: %w", err)
	}

	// Run metadata and cover downloads concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 3) // Buffer for potential errors

	// Populate metadata first
	if viper.GetBool(key.MetadataFetchAnilist) {
		if err := chapter.Manga.PopulateMetadata(progress); err != nil {
			log.Warn(fmt.Sprintf("failed to populate metadata: %v", err))
			errChan <- err
		}
	}

	// Then write series.json after metadata is populated
	if viper.GetBool(key.MetadataSeriesJSON) {
		metadataPath, err := chapter.Manga.Path(false)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to get metadata path: %v", err))
			errChan <- err
		} else {
			jsonPath := filepath.Join(metadataPath, "series.json")
			progress("Generating series.json")
			
			seriesJSON := chapter.Manga.SeriesJSON()
			buf := &bytes.Buffer{}
			encoder := json.NewEncoder(buf)
			encoder.SetIndent("", "  ")
			
			if err := encoder.Encode(seriesJSON); err != nil {
				log.Warn(fmt.Sprintf("failed to encode series JSON: %v", err))
				errChan <- err
			} else {
				if err := filesystem.Api().WriteFile(jsonPath, buf.Bytes(), os.ModePerm); err != nil {
					log.Warn(fmt.Sprintf("failed to write series JSON: %v", err))
					errChan <- err
				}
			}
		}
	}

	// Download cover in parallel since it doesn't depend on metadata
	if viper.GetBool(key.DownloaderDownloadCover) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := chapter.Manga.DownloadCover(false, "", progress); err != nil {
				log.Warn(fmt.Sprintf("failed to download cover: %v", err))
				errChan <- err
			}
		}()
	}

	// Wait for cover download to finish
	wg.Wait()
	close(errChan)

	// Check for any errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return "", fmt.Errorf("encountered %d errors during metadata/cover operations", len(errs))
	}

	log.Info("getting " + viper.GetString(key.FormatsUse) + " converter")
	progress(fmt.Sprintf(
		"Converting %d pages to %s %s",
		len(pages),
		style.Fg(color.Yellow)(viper.GetString(key.FormatsUse)),
		style.Faint(chapter.SizeHuman()),
	))

	conv, err := converter.Get(viper.GetString(key.FormatsUse))
	if err != nil {
		log.Error(err)
		return "", err
	}

	log.Info("converting " + viper.GetString(key.FormatsUse))
	path, err = conv.Save(chapter)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if viper.GetBool(key.HistorySaveOnDownload) {
		go func() {
			err = history.Save(chapter)
			if err != nil {
				log.Warn(err)
			} else {
				log.Info("history saved")
			}
		}()
	}

	log.Info("downloaded without errors")
	progress("Downloaded")
	return path, nil
}
