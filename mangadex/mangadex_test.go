package mangadex

import (
	"os"
	"testing"
)

func TestSearchManga(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "Search for The Beginning After The End",
			query:   "The Beginning After The End",
			wantErr: false,
		},
		{
			name:    "Search for non-existent manga",
			query:   "ThisMangaDoesNotExist12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga, err := client.SearchManga(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchManga() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && manga == nil {
				t.Error("SearchManga() returned nil manga when no error was expected")
			}
		})
	}
}

func TestSearchAndSave(t *testing.T) {
	tempFile := "/tmp/test_series.json"
	defer os.Remove(tempFile)

	tests := []struct {
		name       string
		query      string
		outputPath string
		wantErr    bool
	}{
		{
			name:       "Search and save The Beginning After The End",
			query:      "The Beginning After The End",
			outputPath: tempFile,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manga, err := SearchAndSave(tt.query, tt.outputPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchAndSave() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if manga == nil {
					t.Error("SearchAndSave() returned nil manga when no error was expected")
					return
				}

				// Check if file exists
				if _, err := os.Stat(tt.outputPath); os.IsNotExist(err) {
					t.Error("SearchAndSave() did not create the output file")
				}
			}
		})
	}
}
