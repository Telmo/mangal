package pdf

import (
	"fmt"
	"io"
	"os"

	"github.com/metafates/mangal/filesystem"
	"github.com/metafates/mangal/source"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type PDF struct{}

// New creates a new PDF converter
func New() *PDF {
	return &PDF{}
}

func (p *PDF) Save(chapter *source.Chapter) (string, error) {
	return p.save(chapter, false)
}

func (p *PDF) SaveTemp(chapter *source.Chapter) (string, error) {
	return p.save(chapter, true)
}

func (p *PDF) save(chapter *source.Chapter, temp bool) (string, error) {
	path, err := chapter.Path(temp)
	if err != nil {
		return "", fmt.Errorf("failed to get chapter path: %w", err)
	}

	file, err := filesystem.Api().Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := p.Convert(file, chapter); err != nil {
		return "", err
	}

	return path, nil
}

func (p *PDF) Convert(w io.Writer, chapter *source.Chapter) error {
	if err := pagesToPDF(w, chapter.Pages); err != nil {
		return fmt.Errorf("failed to convert pages to PDF: %w", err)
	}

	return nil
}

func (p *PDF) Extension() string {
	return "pdf"
}

func (p *PDF) String() string {
	return "PDF"
}

func (p *PDF) Name() string {
	return "pdf"
}

// pagesToPDF will convert images to PDF and write to w
func pagesToPDF(w io.Writer, pages []*source.Page) error {
	// Create a temporary directory for the images
	tempDir, err := os.MkdirTemp("", "mangal-pdf-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temporary output file
	tempOutput, err := os.CreateTemp("", "mangal-pdf-output-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp output file: %w", err)
	}
	defer os.Remove(tempOutput.Name())
	defer tempOutput.Close()

	// Write pages to temporary files
	var imageFiles []string
	for i, page := range pages {
		if page.Contents == nil {
			continue
		}

		tempFile, err := os.CreateTemp(tempDir, fmt.Sprintf("page-%d-*.%s", i, page.Extension))
		if err != nil {
			return fmt.Errorf("failed to create temp file for page %d: %w", i, err)
		}
		defer os.Remove(tempFile.Name())

		if _, err := io.Copy(tempFile, page.Contents); err != nil {
			return fmt.Errorf("failed to write page %d to temp file: %w", i, err)
		}
		tempFile.Close()

		imageFiles = append(imageFiles, tempFile.Name())
	}

	// Convert images to PDF
	if err := api.ImportImagesFile(imageFiles, tempOutput.Name(), nil, nil); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	// Copy the PDF to the output writer
	if _, err := tempOutput.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek temp output file: %w", err)
	}

	if _, err := io.Copy(w, tempOutput); err != nil {
		return fmt.Errorf("failed to copy PDF to output: %w", err)
	}

	return nil
}
