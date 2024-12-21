package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/metafates/mangal/model"
)

// SaveMangaMetadata saves manga metadata to the SQLite database
func SaveMangaMetadata(db *sql.DB, manga *model.Manga) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Convert slices to JSON strings
	genres, err := json.Marshal(manga.Metadata.Genres)
	if err != nil {
		return fmt.Errorf("failed to marshal genres: %w", err)
	}

	tags, err := json.Marshal(manga.Metadata.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	characters, err := json.Marshal(manga.Metadata.Characters)
	if err != nil {
		return fmt.Errorf("failed to marshal characters: %w", err)
	}

	storyStaff, err := json.Marshal(manga.Metadata.Staff.Story)
	if err != nil {
		return fmt.Errorf("failed to marshal story staff: %w", err)
	}

	artStaff, err := json.Marshal(manga.Metadata.Staff.Art)
	if err != nil {
		return fmt.Errorf("failed to marshal art staff: %w", err)
	}

	translationStaff, err := json.Marshal(manga.Metadata.Staff.Translation)
	if err != nil {
		return fmt.Errorf("failed to marshal translation staff: %w", err)
	}

	letteringStaff, err := json.Marshal(manga.Metadata.Staff.Lettering)
	if err != nil {
		return fmt.Errorf("failed to marshal lettering staff: %w", err)
	}

	urls, err := json.Marshal(manga.Metadata.URLs)
	if err != nil {
		return fmt.Errorf("failed to marshal URLs: %w", err)
	}

	synonyms, err := json.Marshal(manga.Metadata.Synonyms)
	if err != nil {
		return fmt.Errorf("failed to marshal synonyms: %w", err)
	}

	metadata, err := json.Marshal(manga.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO manga (
			id, title, description, url, cover, source_id, source_name,
			status, start_year, start_month, end_year, end_month,
			total_chapters, publisher, format, volumes, average_score,
			popularity, mean_score, is_licensed, updated_at, genres, tags, characters,
			staff_story, staff_art, staff_translation, staff_lettering, urls, banner_image, synonyms, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(title) DO UPDATE SET
			id = excluded.id,
			description = excluded.description,
			url = excluded.url,
			cover = excluded.cover,
			source_id = excluded.source_id,
			source_name = excluded.source_name,
			status = excluded.status,
			start_year = excluded.start_year,
			start_month = excluded.start_month,
			end_year = excluded.end_year,
			end_month = excluded.end_month,
			total_chapters = excluded.total_chapters,
			publisher = excluded.publisher,
			format = excluded.format,
			volumes = excluded.volumes,
			average_score = excluded.average_score,
			popularity = excluded.popularity,
			mean_score = excluded.mean_score,
			is_licensed = excluded.is_licensed,
			updated_at = excluded.updated_at,
			genres = excluded.genres,
			tags = excluded.tags,
			characters = excluded.characters,
			staff_story = excluded.staff_story,
			staff_art = excluded.staff_art,
			staff_translation = excluded.staff_translation,
			staff_lettering = excluded.staff_lettering,
			urls = excluded.urls,
			banner_image = excluded.banner_image,
			synonyms = excluded.synonyms,
			metadata = excluded.metadata
		`,
		manga.ID, manga.Title, manga.Description, manga.URL,
		manga.Metadata.Cover.ExtraLarge, manga.SourceID, manga.SourceName,
		manga.Metadata.Status,
		manga.Metadata.StartDate.Year, manga.Metadata.StartDate.Month,
		manga.Metadata.EndDate.Year, manga.Metadata.EndDate.Month,
		manga.Metadata.Chapters,
		manga.Metadata.Publisher,
		manga.Metadata.Format,
		manga.Metadata.Volumes,
		manga.Metadata.AverageScore,
		manga.Metadata.Popularity,
		manga.Metadata.MeanScore,
		manga.Metadata.IsLicensed,
		manga.Metadata.UpdatedAt,
		string(genres),
		string(tags),
		string(characters),
		string(storyStaff),
		string(artStaff),
		string(translationStaff),
		string(letteringStaff),
		string(urls),
		manga.Metadata.BannerImage,
		string(synonyms),
		string(metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to insert/update manga: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetMangaMetadata retrieves manga metadata from the SQLite database
func GetMangaMetadata(db *sql.DB, id string) (*model.Manga, error) {
	manga := &model.Manga{
		ID: id,
	}

	var (
		genresJson         string
		tagsJson          string
		charactersJson    string
		storyStaffJson    string
		artStaffJson      string
		translationStaffJson string
		letteringStaffJson string
		urlsJson          string
		synonymsJson      string
		metadataJson      string
	)

	err := db.QueryRow(`
		SELECT title, description, url, cover, source_id, source_name,
		status, start_year, start_month, end_year, end_month,
		total_chapters, publisher, format, volumes, average_score,
		popularity, mean_score, is_licensed, updated_at, genres, tags, characters,
		staff_story, staff_art, staff_translation, staff_lettering, urls, banner_image, synonyms, metadata
		FROM manga WHERE id = ?`, id,
	).Scan(
		&manga.Title,
		&manga.Description,
		&manga.URL,
		&manga.Metadata.Cover.ExtraLarge,
		&manga.SourceID,
		&manga.SourceName,
		&manga.Metadata.Status,
		&manga.Metadata.StartDate.Year,
		&manga.Metadata.StartDate.Month,
		&manga.Metadata.EndDate.Year,
		&manga.Metadata.EndDate.Month,
		&manga.Metadata.Chapters,
		&manga.Metadata.Publisher,
		&manga.Metadata.Format,
		&manga.Metadata.Volumes,
		&manga.Metadata.AverageScore,
		&manga.Metadata.Popularity,
		&manga.Metadata.MeanScore,
		&manga.Metadata.IsLicensed,
		&manga.Metadata.UpdatedAt,
		&genresJson,
		&tagsJson,
		&charactersJson,
		&storyStaffJson,
		&artStaffJson,
		&translationStaffJson,
		&letteringStaffJson,
		&urlsJson,
		&manga.Metadata.BannerImage,
		&synonymsJson,
		&metadataJson,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get manga metadata: %w", err)
	}

	// Unmarshal JSON strings to slices
	if err := json.Unmarshal([]byte(genresJson), &manga.Metadata.Genres); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genres: %w", err)
	}
	if err := json.Unmarshal([]byte(tagsJson), &manga.Metadata.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}
	if err := json.Unmarshal([]byte(charactersJson), &manga.Metadata.Characters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal characters: %w", err)
	}
	if err := json.Unmarshal([]byte(storyStaffJson), &manga.Metadata.Staff.Story); err != nil {
		return nil, fmt.Errorf("failed to unmarshal story staff: %w", err)
	}
	if err := json.Unmarshal([]byte(artStaffJson), &manga.Metadata.Staff.Art); err != nil {
		return nil, fmt.Errorf("failed to unmarshal art staff: %w", err)
	}
	if err := json.Unmarshal([]byte(translationStaffJson), &manga.Metadata.Staff.Translation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal translation staff: %w", err)
	}
	if err := json.Unmarshal([]byte(letteringStaffJson), &manga.Metadata.Staff.Lettering); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lettering staff: %w", err)
	}
	if err := json.Unmarshal([]byte(urlsJson), &manga.Metadata.URLs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URLs: %w", err)
	}
	if err := json.Unmarshal([]byte(synonymsJson), &manga.Metadata.Synonyms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal synonyms: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJson), &manga.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return manga, nil
}

// InitMangaDB initializes the database tables
func InitMangaDB(db *sql.DB) error {
	return CreateMangaTable(db)
}
