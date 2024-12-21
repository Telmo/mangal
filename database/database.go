package database

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/metafates/mangal/where"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

// GetDB returns a singleton instance of the SQLite database
func GetDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		dbPath := where.Database()
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return
		}

		err = initDatabase(db)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// initDatabase creates the necessary tables if they don't exist
func initDatabase(db *sql.DB) error {
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

	// Create manga_genres table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_genres (
			manga_id TEXT,
			genre_id TEXT,
			PRIMARY KEY (manga_id, genre_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
			FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_genres table: %w", err)
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

	// Create manga_staff table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_staff (
			manga_id TEXT,
			staff_id TEXT,
			role TEXT,
			PRIMARY KEY (manga_id, staff_id, role),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
			FOREIGN KEY (staff_id) REFERENCES staff(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_staff table: %w", err)
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

	// Create manga_characters table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_characters (
			manga_id TEXT,
			character_id TEXT,
			PRIMARY KEY (manga_id, character_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_characters table: %w", err)
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

	// Create manga_tags table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS manga_tags (
			manga_id TEXT,
			tag_id TEXT,
			PRIMARY KEY (manga_id, tag_id),
			FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create manga_tags table: %w", err)
	}

	return nil
}
