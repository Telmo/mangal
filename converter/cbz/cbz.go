package cbz

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/source"
	"github.com/metafates/mangal/util"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"time"
)

type CBZ struct{}

func New() *CBZ {
	return &CBZ{}
}

func (*CBZ) Save(chapter *source.Chapter) (string, error) {
	return save(chapter, false)
}

func (*CBZ) SaveTemp(chapter *source.Chapter) (string, error) {
	return save(chapter, true)
}

func save(chapter *source.Chapter, temp bool) (path string, err error) {
	// Get the final path for the CBZ file
	path, err = chapter.Path(temp)
	if err != nil {
		return "", fmt.Errorf("failed to get chapter path: %w", err)
	}

	// Create a temporary directory for downloading pages
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("mangal-cbz-%d", time.Now().UnixNano()))
	err = filesystem.Api().MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer filesystem.Api().RemoveAll(tempDir)

	// Download pages to temp directory
	for i, page := range chapter.Pages {
		pagePath := filepath.Join(tempDir, fmt.Sprintf("%d.%s", i+1, page.Extension))
		err = filesystem.Api().WriteFile(pagePath, page.Contents.Bytes(), os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to write page %d: %w", i+1, err)
		}
	}

	// Remove any existing file/directory at the target path
	err = filesystem.Api().RemoveAll(path)
	if err != nil {
		return "", fmt.Errorf("failed to remove existing file/directory: %w", err)
	}

	// Create the CBZ file
	err = SaveTo(chapter, path)
	if err != nil {
		return "", fmt.Errorf("failed to create CBZ file: %w", err)
	}

	return path, nil
}

func SaveTo(chapter *source.Chapter, to string) error {
	// First remove any existing directory at the target path
	if err := filesystem.Api().RemoveAll(to); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	cbzFile, err := filesystem.Api().Create(to)
	if err != nil {
		return fmt.Errorf("failed to create CBZ file: %w", err)
	}

	defer util.Ignore(cbzFile.Close)

	zipWriter := zip.NewWriter(cbzFile)
	defer util.Ignore(zipWriter.Close)

	for _, page := range chapter.Pages {
		if err = addToZip(zipWriter, page.Contents, page.Filename()); err != nil {
			return fmt.Errorf("failed to add page to CBZ: %w", err)
		}
	}

	if viper.GetBool(key.MetadataComicInfoXML) {
		comicInfo := chapter.ComicInfo()
		marshalled, err := xml.MarshalIndent(comicInfo, "", "  ")
		if err == nil {
			buf := bytes.NewBuffer(marshalled)
			err = addToZip(zipWriter, buf, "ComicInfo.xml")
			if err != nil {
				return fmt.Errorf("failed to add ComicInfo.xml to CBZ: %w", err)
			}
		}
	}

	return nil
}

func addToZip(writer *zip.Writer, file io.Reader, name string) error {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	}

	headerWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(headerWriter, file)
	return err
}
