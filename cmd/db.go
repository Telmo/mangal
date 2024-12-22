package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/metafates/mangal/db"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/model"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Commands for managing the PostgreSQL database, including initialization, migrations, and data operations.`,
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the PostgreSQL database",
	Long:  `Initializes the database schema by running migrations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Initializing database...")

		// Connect to the database
		database, err := db.GetDB()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to connect to database: %v", err))
			return nil // Return nil to prevent cobra from showing the error again
		}

		// Run migrations
		force, _ := cmd.Flags().GetBool("force")
		if err := db.Migrate(database, force); err != nil {
			log.Error(fmt.Sprintf("Failed to run migrations: %v", err))
			if database != nil {
				database.Close()
			}
			return nil
		}

		// Close database connection after migrations
		if database != nil {
			database.Close()
		}

		log.Info("Database initialized successfully")
		return nil
	},
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Applies any pending database migrations to update the schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Running database migrations...")

		database, err := db.GetDB()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to connect to database: %v", err))
			return nil
		}
		defer func() {
			if database != nil {
				database.Close()
			}
		}()

		force, _ := cmd.Flags().GetBool("force")
		if err := db.Migrate(database, force); err != nil {
			log.Error(fmt.Sprintf("Failed to run migrations: %v", err))
			return nil
		}

		log.Info("Database migrations completed successfully")
		return nil
	},
}

var dbInsertCmd = &cobra.Command{
	Use:   "insert [path to series.json]",
	Short: "Insert manga data from a JSON file",
	Long:  `Inserts manga metadata from a series.json file into the database.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read and parse the JSON file
		jsonPath := args[0]
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to read JSON file: %v", err))
			return nil
		}

		var seriesJSON model.SeriesJSON
		if err := json.Unmarshal(data, &seriesJSON); err != nil {
			log.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
			return nil
		}

		// Convert SeriesJSON to Manga
		manga := model.SeriesJSONToManga(&seriesJSON)

		// Connect to database
		database, err := db.GetDB()
		if err != nil {
			log.Error(fmt.Sprintf("Failed to connect to database: %v", err))
			return nil
		}
		defer func() {
			if database != nil {
				database.Close()
			}
		}()

		// Save manga metadata
		if err := db.SaveMangaMetadata(database, manga); err != nil {
			log.Error(fmt.Sprintf("Failed to save manga metadata: %v", err))
			return nil
		}

		log.Info("Successfully inserted manga metadata")
		return nil
	},
}

var dbSearchCmd = &cobra.Command{
	Use:   "search [name]",
	Short: "Search for manga by name",
	Long:  `Search for manga in the database by name and output the data as JSON.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get database connection
		dbConn, err := db.GetDB()
		if err != nil {
			log.Error("Failed to get database connection:", err)
			return nil
		}
		defer func() {
			if dbConn != nil {
				dbConn.Close()
			}
		}()

		// Search for manga
		manga, err := db.SearchMangaByName(dbConn, args[0])
		if err != nil {
			log.Error("Failed to search manga:", err)
			return nil
		}

		if manga == nil {
			log.Info("No manga found")
			return nil
		}

		// Create output structure matching series.json
		output := struct {
			Metadata struct {
				Type                 string   `json:"type"`
				Name                 string   `json:"name"`
				DescriptionFormatted string   `json:"descriptionFormatted"`
				DescriptionText      string   `json:"descriptionText"`
				Publisher            string   `json:"publisher"`
				Status               string   `json:"status"`
				Year                 int      `json:"year"`
				TotalChapters        int      `json:"totalChapters"`
				TotalIssues          int      `json:"totalIssues"`
				BookType             string   `json:"bookType"`
				ComicImage           string   `json:"comicImage"`
				ComicID              int      `json:"comicID"`
				PublicationRun       string   `json:"publicationRun"`
				Genres               []string `json:"genres"`
				Tags                 []string `json:"tags"`
				Characters           []string `json:"characters"`
				Staff                struct {
					Story       []string `json:"story"`
					Art         []string `json:"art"`
					Translation []string `json:"translation"`
					Lettering   []string `json:"lettering"`
				} `json:"staff"`
				Volumes      int      `json:"volumes"`
				Chapters     int      `json:"chapters"`
				AverageScore int      `json:"averageScore"`
				Popularity   int      `json:"popularity"`
				MeanScore    int      `json:"meanScore"`
				IsLicensed   bool     `json:"isLicensed"`
				UpdatedAt    int      `json:"updatedAt"`
				URLs         []string `json:"urls"`
				BannerImage  string   `json:"bannerImage"`
				Cover        struct {
					ExtraLarge string `json:"extraLarge"`
					Large      string `json:"large"`
					Medium     string `json:"medium"`
					Color      string `json:"color"`
				} `json:"cover"`
				Synonyms []string `json:"synonyms"`
				Format   string   `json:"format"`
			} `json:"metadata"`
		}{}

		// Fill output structure from manga data
		output.Metadata.Type = "comicSeries"
		output.Metadata.Name = manga.Title
		output.Metadata.DescriptionFormatted = manga.Description
		output.Metadata.DescriptionText = manga.Description
		output.Metadata.Publisher = manga.Metadata.Publisher
		output.Metadata.Status = manga.Metadata.Status
		output.Metadata.Year = manga.Metadata.StartDate.Year
		output.Metadata.TotalChapters = manga.Metadata.Chapters
		output.Metadata.TotalIssues = manga.Metadata.Chapters
		output.Metadata.BookType = "manga"
		output.Metadata.ComicImage = manga.Metadata.Cover.ExtraLarge
		output.Metadata.ComicID = 0
		output.Metadata.PublicationRun = fmt.Sprintf("1 %d - Present", manga.Metadata.StartDate.Year)
		output.Metadata.Genres = manga.Metadata.Genres
		output.Metadata.Tags = manga.Metadata.Tags
		output.Metadata.Characters = manga.Metadata.Characters
		output.Metadata.Staff.Story = []string{}
		output.Metadata.Staff.Art = []string{}
		output.Metadata.Staff.Translation = []string{}
		output.Metadata.Staff.Lettering = []string{}
		output.Metadata.Volumes = manga.Metadata.Volumes
		output.Metadata.Chapters = manga.Metadata.Chapters
		output.Metadata.AverageScore = manga.Metadata.AverageScore
		output.Metadata.Popularity = manga.Metadata.Popularity
		output.Metadata.MeanScore = manga.Metadata.MeanScore
		output.Metadata.IsLicensed = manga.Metadata.IsLicensed
		output.Metadata.UpdatedAt = manga.Metadata.UpdatedAt
		output.Metadata.URLs = manga.Metadata.URLs
		output.Metadata.BannerImage = manga.Metadata.BannerImage
		output.Metadata.Cover = struct {
			ExtraLarge string `json:"extraLarge"`
			Large      string `json:"large"`
			Medium     string `json:"medium"`
			Color      string `json:"color"`
		}{
			ExtraLarge: manga.Metadata.Cover.ExtraLarge,
			Large:      manga.Metadata.Cover.Large,
			Medium:     manga.Metadata.Cover.Medium,
			Color:      manga.Metadata.Cover.Color,
		}
		output.Metadata.Synonyms = manga.Metadata.Synonyms
		output.Metadata.Format = manga.Metadata.Format

		// Output as JSON
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Error("Failed to marshal JSON:", err)
			return nil
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbInsertCmd)
	dbCmd.AddCommand(dbSearchCmd)

	// Add force flag to init command
	dbInitCmd.Flags().BoolP("force", "f", false, "Force clean database state if dirty")
	dbMigrateCmd.Flags().BoolP("force", "f", false, "Force clean database state if dirty")
}
