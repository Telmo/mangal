package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		return "", err
	}

	log.Info("downloading " + chapter.Manga.Name + " - " + chapter.Name + " to : " + path)

	if viper.GetBool(key.DownloaderRedownloadExisting) {
		log.Info("chapter already downloaded, deleting and redownloading")
		err = filesystem.Api().Remove(path)
		if err != nil {
			log.Warn(err)
		}
	} else {
		log.Info("checking if chapter is already downloaded")
		if chapter.IsDownloaded() {
			log.Info("chapter already downloaded, skipping")
			return path, nil
		}
	}

	progress("Getting pages")
	pages, err := chapter.Source().PagesOf(chapter)
	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Info("found " + fmt.Sprintf("%d", len(pages)) + " pages")

	err = chapter.DownloadPages(false, progress)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if viper.GetBool(key.MetadataFetchAnilist) {
		err := chapter.Manga.PopulateMetadata(progress)
		if err != nil {
			log.Warn(err)
		}
	}

	if viper.GetBool(key.MetadataSeriesJSON) {
		path, err := chapter.Manga.Path(false)
		if err != nil {
			log.Warn(err)
		} else {
			path = filepath.Join(path, "series.json")
			progress("Generating series.json")
			seriesJSON := chapter.Manga.SeriesJSON()
			buf := &bytes.Buffer{}
			encoder := json.NewEncoder(buf)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(seriesJSON)
			if err != nil {
				log.Warn(err)
			} else {
				err = filesystem.Api().WriteFile(path, buf.Bytes(), os.ModePerm)
				if err != nil {
					log.Warn(err)
				}
			}
		}
	}

	if viper.GetBool(key.DownloaderDownloadCover) {
		coverDir, err := chapter.Manga.Path(false)
		if err == nil {
			_ = chapter.Manga.DownloadCover(false, coverDir, progress)
		}
	}

	log.Info("getting " + viper.GetString(key.FormatsUse) + " converter")
	progress(fmt.Sprintf(
		"Converting %d pages to %s %s",
		len(pages),
		style.Fg(color.Yellow)(viper.GetString(key.FormatsUse)),
		style.Faint(chapter.SizeHuman())),
	)
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
