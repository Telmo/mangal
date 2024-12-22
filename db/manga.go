package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/metafates/mangal/model"
	"github.com/metafates/mangal/util/sanitize"
)

// SaveMangaMetadata saves manga metadata to the PostgreSQL database
func SaveMangaMetadata(db *sql.DB, manga *model.Manga) error {
	// Sanitize input data
	manga.Title = sanitize.Text(manga.Title)
	manga.Description = sanitize.Text(manga.Description)
	manga.Metadata.Publisher = sanitize.Text(manga.Metadata.Publisher)
	manga.Metadata.Status = sanitize.Text(manga.Metadata.Status)
	manga.Metadata.PublicationRun = sanitize.Text(manga.Metadata.PublicationRun)
	manga.Metadata.BannerImage = sanitize.URL(manga.Metadata.BannerImage)
	manga.Metadata.Cover.ExtraLarge = sanitize.URL(manga.Metadata.Cover.ExtraLarge)
	manga.Metadata.Cover.Large = sanitize.URL(manga.Metadata.Cover.Large)
	manga.Metadata.Cover.Medium = sanitize.URL(manga.Metadata.Cover.Medium)

	// Sanitize arrays
	for i, url := range manga.Metadata.URLs {
		manga.Metadata.URLs[i] = sanitize.URL(url)
	}
	for i, genre := range manga.Metadata.Genres {
		manga.Metadata.Genres[i] = sanitize.Text(genre)
	}
	for i, tag := range manga.Metadata.Tags {
		manga.Metadata.Tags[i] = sanitize.Text(tag)
	}
	for i, character := range manga.Metadata.Characters {
		manga.Metadata.Characters[i] = sanitize.Text(character)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into series table
	var seriesID string
	err = tx.QueryRow(`
		INSERT INTO series (
			name, 
			description_formatted,
			description_text,
			publisher,
			status,
			year,
			total_chapters,
			total_issues,
			book_type,
			comic_image,
			comic_id,
			publication_run,
			volumes,
			chapters,
			average_score,
			popularity,
			mean_score,
			is_licensed,
			updated_at,
			banner_image,
			format
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (name) DO UPDATE SET
			description_formatted = EXCLUDED.description_formatted,
			description_text = EXCLUDED.description_text,
			publisher = EXCLUDED.publisher,
			status = EXCLUDED.status,
			year = EXCLUDED.year,
			total_chapters = EXCLUDED.total_chapters,
			total_issues = EXCLUDED.total_issues,
			book_type = EXCLUDED.book_type,
			comic_image = EXCLUDED.comic_image,
			comic_id = EXCLUDED.comic_id,
			publication_run = EXCLUDED.publication_run,
			volumes = EXCLUDED.volumes,
			chapters = EXCLUDED.chapters,
			average_score = EXCLUDED.average_score,
			popularity = EXCLUDED.popularity,
			mean_score = EXCLUDED.mean_score,
			is_licensed = EXCLUDED.is_licensed,
			updated_at = EXCLUDED.updated_at,
			banner_image = EXCLUDED.banner_image,
			format = EXCLUDED.format
		RETURNING id
	`, 
		manga.Title,
		manga.Description,
		manga.Description,
		manga.Metadata.Publisher,
		manga.Metadata.Status,
		manga.Metadata.StartDate.Year,
		manga.Metadata.Chapters,
		manga.Metadata.Chapters,
		"manga",
		manga.Metadata.Cover.ExtraLarge,
		0, // comic_id
		manga.Metadata.PublicationRun,
		manga.Metadata.Volumes,
		manga.Metadata.Chapters,
		manga.Metadata.AverageScore,
		manga.Metadata.Popularity,
		manga.Metadata.MeanScore,
		manga.Metadata.IsLicensed,
		manga.Metadata.UpdatedAt,
		manga.Metadata.BannerImage,
		manga.Metadata.Format,
	).Scan(&seriesID)

	if err != nil {
		return fmt.Errorf("failed to insert/update series: %w", err)
	}

	// Insert cover if exists
	if manga.Metadata.Cover.ExtraLarge != "" || manga.Metadata.Cover.Large != "" || manga.Metadata.Cover.Medium != "" {
		// First delete existing cover
		_, err = tx.Exec(`DELETE FROM covers WHERE series_id = $1`, seriesID)
		if err != nil {
			return fmt.Errorf("failed to delete existing cover: %w", err)
		}

		// Then insert new cover
		_, err = tx.Exec(`
			INSERT INTO covers (series_id, extra_large, large, medium, color)
			VALUES ($1, $2, $3, $4, $5)
		`, seriesID, 
			manga.Metadata.Cover.ExtraLarge,
			manga.Metadata.Cover.Large,
			manga.Metadata.Cover.Medium,
			manga.Metadata.Cover.Color,
		)
		if err != nil {
			return fmt.Errorf("failed to insert cover: %w", err)
		}
	}

	// Insert URLs
	if len(manga.Metadata.URLs) > 0 {
		// First delete existing URLs
		_, err = tx.Exec(`DELETE FROM urls WHERE series_id = $1`, seriesID)
		if err != nil {
			return fmt.Errorf("failed to delete existing URLs: %w", err)
		}

		// Then insert new URLs
		for _, url := range manga.Metadata.URLs {
			_, err = tx.Exec(`
				INSERT INTO urls (series_id, url)
				VALUES ($1, $2)
			`, seriesID, url)
			if err != nil {
				return fmt.Errorf("failed to insert URL: %w", err)
			}
		}
	}

	// Insert genres
	if len(manga.Metadata.Genres) > 0 {
		// First delete existing genres
		_, err = tx.Exec(`DELETE FROM series_genres WHERE series_id = $1`, seriesID)
		if err != nil {
			return fmt.Errorf("failed to delete existing genres: %w", err)
		}

		// Then insert new genres
		for _, genre := range manga.Metadata.Genres {
			var genreID string
			err = tx.QueryRow(`
				INSERT INTO genres (name)
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id
			`, genre).Scan(&genreID)
			if err != nil {
				return fmt.Errorf("failed to insert/update genre: %w", err)
			}

			_, err = tx.Exec(`
				INSERT INTO series_genres (series_id, genre_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, seriesID, genreID)
			if err != nil {
				return fmt.Errorf("failed to insert series_genre: %w", err)
			}
		}
	}

	// Insert tags
	if len(manga.Metadata.Tags) > 0 {
		// First delete existing tags
		_, err = tx.Exec(`DELETE FROM series_tags WHERE series_id = $1`, seriesID)
		if err != nil {
			return fmt.Errorf("failed to delete existing tags: %w", err)
		}

		// Then insert new tags
		for _, tag := range manga.Metadata.Tags {
			var tagID string
			err = tx.QueryRow(`
				INSERT INTO tags (name)
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id
			`, tag).Scan(&tagID)
			if err != nil {
				return fmt.Errorf("failed to insert/update tag: %w", err)
			}

			_, err = tx.Exec(`
				INSERT INTO series_tags (series_id, tag_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, seriesID, tagID)
			if err != nil {
				return fmt.Errorf("failed to insert series_tag: %w", err)
			}
		}
	}

	// Insert characters
	if len(manga.Metadata.Characters) > 0 {
		// First delete existing characters
		_, err = tx.Exec(`DELETE FROM series_characters WHERE series_id = $1`, seriesID)
		if err != nil {
			return fmt.Errorf("failed to delete existing characters: %w", err)
		}

		// Then insert new characters
		for _, character := range manga.Metadata.Characters {
			var characterID string
			err = tx.QueryRow(`
				INSERT INTO characters (name)
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id
			`, character).Scan(&characterID)
			if err != nil {
				return fmt.Errorf("failed to insert/update character: %w", err)
			}

			_, err = tx.Exec(`
				INSERT INTO series_characters (series_id, character_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, seriesID, characterID)
			if err != nil {
				return fmt.Errorf("failed to insert series_character: %w", err)
			}
		}
	}

	// Insert synonyms
	if len(manga.Metadata.Synonyms) > 0 {
		// First delete existing synonyms
		_, err = tx.Exec(`DELETE FROM synonyms WHERE series_id = $1`, seriesID)
		if err != nil {
			return fmt.Errorf("failed to delete existing synonyms: %w", err)
		}

		// Then insert new synonyms
		for _, synonym := range manga.Metadata.Synonyms {
			_, err = tx.Exec(`
				INSERT INTO synonyms (series_id, name)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, seriesID, synonym)
			if err != nil {
				return fmt.Errorf("failed to insert synonym: %w", err)
			}
		}
	}

	return tx.Commit()
}

// SearchMangaByName searches for manga in the database by name
func SearchMangaByName(db *sql.DB, name string) (*model.Manga, error) {
	var manga model.Manga
	var seriesID string
	var comic_id int
	var publication_run string
	var urlsStr string
	var coverJSON []byte

	err := db.QueryRow(`
		SELECT 
			s.id,
			s.name,
			s.description_formatted,
			s.description_text,
			s.publisher,
			s.status,
			s.year,
			s.total_chapters,
			s.total_issues,
			s.book_type,
			s.comic_image,
			s.comic_id,
			CASE 
				WHEN s.status = 'Completed' THEN '1 ' || s.year || ' - 4 2024'
				ELSE '1 ' || s.year || ' - Present'
			END as publication_run,
			s.volumes,
			s.chapters,
			s.average_score,
			s.popularity,
			s.mean_score,
			s.is_licensed,
			s.updated_at,
			s.banner_image,
			s.format,
			COALESCE((SELECT url FROM urls WHERE series_id = s.id LIMIT 1), '') as url,
			COALESCE(
				(SELECT string_agg(url, ',') FROM urls WHERE series_id = s.id),
				''
			) as urls,
			COALESCE(
				(SELECT json_build_object(
					'extraLarge', extra_large,
					'large', large,
					'medium', medium,
					'color', color
				)::text FROM covers WHERE series_id = s.id LIMIT 1),
				'{"extraLarge":"","large":"","medium":"","color":""}'
			)::jsonb as cover
		FROM series s
		WHERE LOWER(s.name) LIKE LOWER($1)
		OR EXISTS (
			SELECT 1 FROM synonyms 
			WHERE series_id = s.id 
			AND LOWER(name) LIKE LOWER($1)
		)
		LIMIT 1
	`, "%"+name+"%").Scan(
		&seriesID,
		&manga.Title,
		&manga.Description,
		&manga.Description,
		&manga.Metadata.Publisher,
		&manga.Metadata.Status,
		&manga.Metadata.StartDate.Year,
		&manga.Metadata.Chapters,
		&manga.Metadata.Chapters,
		&manga.Metadata.Format,
		&manga.Metadata.Cover.ExtraLarge,
		&comic_id,
		&publication_run,
		&manga.Metadata.Volumes,
		&manga.Metadata.Chapters,
		&manga.Metadata.AverageScore,
		&manga.Metadata.Popularity,
		&manga.Metadata.MeanScore,
		&manga.Metadata.IsLicensed,
		&manga.Metadata.UpdatedAt,
		&manga.Metadata.BannerImage,
		&manga.Metadata.Format,
		&manga.URL,
		&urlsStr,
		&coverJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search manga: %w", err)
	}

	// Parse cover JSON
	if err := json.Unmarshal(coverJSON, &manga.Metadata.Cover); err != nil {
		return nil, fmt.Errorf("failed to parse cover JSON: %w", err)
	}

	// Set publication run
	manga.Metadata.PublicationRun = publication_run

	// Get URLs
	manga.Metadata.URLs = make([]string, 0)
	if urlsStr != "" {
		manga.Metadata.URLs = strings.Split(urlsStr, ",")
	}

	// Get genres
	rows, err := db.Query(`
		SELECT g.name 
		FROM genres g
		JOIN series_genres sg ON sg.genre_id = g.id
		WHERE sg.series_id = $1
	`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var genre string
		if err := rows.Scan(&genre); err != nil {
			return nil, fmt.Errorf("failed to scan genre: %w", err)
		}
		manga.Metadata.Genres = append(manga.Metadata.Genres, genre)
	}

	// Get tags
	rows, err = db.Query(`
		SELECT t.name 
		FROM tags t
		JOIN series_tags st ON st.tag_id = t.id
		WHERE st.series_id = $1
	`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		manga.Metadata.Tags = append(manga.Metadata.Tags, tag)
	}

	// Get characters
	rows, err = db.Query(`
		SELECT c.name 
		FROM characters c
		JOIN series_characters sc ON sc.character_id = c.id
		WHERE sc.series_id = $1
	`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get characters: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var character string
		if err := rows.Scan(&character); err != nil {
			return nil, fmt.Errorf("failed to scan character: %w", err)
		}
		manga.Metadata.Characters = append(manga.Metadata.Characters, character)
	}

	// Get synonyms
	rows, err = db.Query(`
		SELECT name 
		FROM synonyms
		WHERE series_id = $1
	`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to get synonyms: %w", err)
	}
	defer rows.Close()

	manga.Metadata.Synonyms = make([]string, 0)
	for rows.Next() {
		var synonym string
		if err := rows.Scan(&synonym); err != nil {
			return nil, fmt.Errorf("failed to scan synonym: %w", err)
		}
		manga.Metadata.Synonyms = append(manga.Metadata.Synonyms, synonym)
	}

	return &manga, nil
}

// InitMangaDB initializes the database tables
func InitMangaDB(db *sql.DB) error {
	return CreateMangaTable(db)
}
