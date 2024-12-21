package database

import (
	"database/sql"
	"fmt"

	"github.com/metafates/mangal/model"
)

// SaveMangaMetadata stores manga metadata in the SQLite database
func SaveMangaMetadata(db *sql.DB, manga *model.Manga) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert manga
	_, err = tx.Exec(`
		INSERT INTO manga (
			id, title, description, url, cover, source_id, source_name,
			status, start_year, start_month, end_year, end_month,
			total_chapters, publisher
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
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
			publisher = excluded.publisher
	`,
		manga.ID,
		manga.Title,
		manga.Description,
		manga.URL,
		manga.Metadata.Cover.ExtraLarge,
		manga.SourceID,
		manga.SourceName,
		manga.Metadata.Status,
		manga.Metadata.StartDate.Year,
		manga.Metadata.StartDate.Month,
		manga.Metadata.EndDate.Year,
		manga.Metadata.EndDate.Month,
		manga.Metadata.Chapters,
		manga.Metadata.Publisher,
	)
	if err != nil {
		return fmt.Errorf("failed to insert manga: %w", err)
	}

	// Delete existing genres
	_, err = tx.Exec("DELETE FROM manga_genres WHERE manga_id = ?", manga.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing genres: %w", err)
	}

	// Insert genres
	for _, genre := range manga.Metadata.Genres {
		// Insert into genres table if not exists
		_, err = tx.Exec("INSERT OR IGNORE INTO genres (id, name) VALUES (?, ?)", genre, genre)
		if err != nil {
			return fmt.Errorf("failed to insert genre: %w", err)
		}

		// Insert into manga_genres
		_, err = tx.Exec("INSERT INTO manga_genres (manga_id, genre_id) VALUES (?, ?)", manga.ID, genre)
		if err != nil {
			return fmt.Errorf("failed to insert manga genre: %w", err)
		}
	}

	// Delete existing staff
	_, err = tx.Exec("DELETE FROM manga_staff WHERE manga_id = ?", manga.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing staff: %w", err)
	}

	// Insert staff (both Story and Art)
	for _, story := range manga.Metadata.Staff.Story {
		// Insert into staff table if not exists
		_, err = tx.Exec("INSERT OR IGNORE INTO staff (id, name) VALUES (?, ?)", story, story)
		if err != nil {
			return fmt.Errorf("failed to insert staff: %w", err)
		}

		// Insert into manga_staff
		_, err = tx.Exec("INSERT INTO manga_staff (manga_id, staff_id, role) VALUES (?, ?, ?)", manga.ID, story, "Story")
		if err != nil {
			return fmt.Errorf("failed to insert manga staff: %w", err)
		}
	}

	for _, art := range manga.Metadata.Staff.Art {
		// Insert into staff table if not exists
		_, err = tx.Exec("INSERT OR IGNORE INTO staff (id, name) VALUES (?, ?)", art, art)
		if err != nil {
			return fmt.Errorf("failed to insert staff: %w", err)
		}

		// Insert into manga_staff
		_, err = tx.Exec("INSERT INTO manga_staff (manga_id, staff_id, role) VALUES (?, ?, ?)", manga.ID, art, "Art")
		if err != nil {
			return fmt.Errorf("failed to insert manga staff: %w", err)
		}
	}

	// Delete existing characters
	_, err = tx.Exec("DELETE FROM manga_characters WHERE manga_id = ?", manga.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing characters: %w", err)
	}

	// Insert characters
	for _, character := range manga.Metadata.Characters {
		// Insert into characters table if not exists
		_, err = tx.Exec("INSERT OR IGNORE INTO characters (id, name) VALUES (?, ?)", character, character)
		if err != nil {
			return fmt.Errorf("failed to insert character: %w", err)
		}

		// Insert into manga_characters
		_, err = tx.Exec("INSERT INTO manga_characters (manga_id, character_id) VALUES (?, ?)", manga.ID, character)
		if err != nil {
			return fmt.Errorf("failed to insert manga character: %w", err)
		}
	}

	// Delete existing tags
	_, err = tx.Exec("DELETE FROM manga_tags WHERE manga_id = ?", manga.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert tags
	for _, tag := range manga.Metadata.Tags {
		// Insert into tags table if not exists
		_, err = tx.Exec("INSERT OR IGNORE INTO tags (id, name) VALUES (?, ?)", tag, tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}

		// Insert into manga_tags
		_, err = tx.Exec("INSERT INTO manga_tags (manga_id, tag_id) VALUES (?, ?)", manga.ID, tag)
		if err != nil {
			return fmt.Errorf("failed to insert manga tag: %w", err)
		}
	}

	return tx.Commit()
}

// GetMangaMetadata retrieves manga metadata from the SQLite database
func GetMangaMetadata(db *sql.DB, id string) (*model.Manga, error) {
	manga := &model.Manga{ID: id}
	
	// Get main manga data
	err := db.QueryRow(`
		SELECT title, description, url, cover, source_id, source_name,
		       status, start_year, start_month, end_year, end_month,
		       total_chapters, publisher
		FROM manga WHERE id = ?`, id,
	).Scan(
		&manga.Title,
		&manga.Description,
		&manga.URL,
		&manga.Cover,
		&manga.SourceID,
		&manga.SourceName,
		&manga.Metadata.Status,
		&manga.Metadata.StartDate.Year,
		&manga.Metadata.StartDate.Month,
		&manga.Metadata.EndDate.Year,
		&manga.Metadata.EndDate.Month,
		&manga.Metadata.Chapters,
		&manga.Metadata.Publisher,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}

	// Get genres
	rows, err := db.Query("SELECT genre_id FROM manga_genres WHERE manga_id = ?", id)
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

	// Get staff
	rows, err = db.Query("SELECT staff_id, role FROM manga_staff WHERE manga_id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var staff string
		var role string
		if err := rows.Scan(&staff, &role); err != nil {
			return nil, fmt.Errorf("failed to scan staff: %w", err)
		}
		if role == "Story" {
			manga.Metadata.Staff.Story = append(manga.Metadata.Staff.Story, staff)
		} else if role == "Art" {
			manga.Metadata.Staff.Art = append(manga.Metadata.Staff.Art, staff)
		}
	}

	// Get characters
	rows, err = db.Query("SELECT character_id FROM manga_characters WHERE manga_id = ?", id)
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

	// Get tags
	rows, err = db.Query("SELECT tag_id FROM manga_tags WHERE manga_id = ?", id)
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

	return manga, nil
}

// InitDB initializes the database tables
func InitDB(db *sql.DB) error {
	// Create manga table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS manga (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			url TEXT,
			cover TEXT,
			source_id TEXT,
			source_name TEXT,
			status TEXT,
			start_year INTEGER,
			start_month INTEGER,
			end_year INTEGER,
			end_month INTEGER,
			total_chapters INTEGER,
			publisher TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga table: %w", err)
	}

	// Create manga_genres table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_genres (
			manga_id TEXT,
			genre_id TEXT,
			PRIMARY KEY (manga_id, genre_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_genres table: %w", err)
	}

	// Create manga_staff table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_staff (
			manga_id TEXT,
			staff_id TEXT,
			role TEXT,
			PRIMARY KEY (manga_id, staff_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_staff table: %w", err)
	}

	// Create manga_characters table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_characters (
			manga_id TEXT,
			character_id TEXT,
			PRIMARY KEY (manga_id, character_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_characters table: %w", err)
	}

	// Create manga_tags table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_tags (
			manga_id TEXT,
			tag_id TEXT,
			PRIMARY KEY (manga_id, tag_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_tags table: %w", err)
	}

	// Create genres table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS genres (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create genres table: %w", err)
	}

	// Create staff table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS staff (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create staff table: %w", err)
	}

	// Create characters table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS characters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create characters table: %w", err)
	}

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	return nil
}
