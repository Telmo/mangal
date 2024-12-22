package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sync"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	dbConn   *sql.DB
	dbMutex  sync.Mutex
	initOnce sync.Once
)

// obfuscateConnStr obfuscates sensitive information in connection string for logging
func obfuscateConnStr(connStr string) string {
	// Find the password part and replace it with asterisks
	parts := strings.Split(connStr, " ")
	for i, part := range parts {
		if strings.HasPrefix(part, "password=") {
			parts[i] = "password=*****"
			break
		}
	}
	return strings.Join(parts, " ")
}

// GetDB returns a singleton instance of the PostgreSQL database
func GetDB() (*sql.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbConn != nil {
		return dbConn, nil
	}

	var err error
	initOnce.Do(func() {
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			viper.GetString("database.host"),
			viper.GetInt("database.port"),
			viper.GetString("database.user"),
			viper.GetString("database.password"),
			viper.GetString("database.name"),
			viper.GetString("database.sslmode"),
		)

		// Log the connection attempt with obfuscated password
		log.Debugf("Connecting to database: %s", obfuscateConnStr(connStr))

		dbConn, err = sql.Open("postgres", connStr)
		if err != nil {
			err = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		if err := dbConn.Ping(); err != nil {
			err = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		log.Info("Successfully connected to database")
	})

	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// Migrate runs database migrations
func Migrate(db *sql.DB, force bool) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if force {
		if err := m.Force(1); err != nil {
			return fmt.Errorf("failed to force migrations: %w", err)
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// initDatabase creates the necessary tables if they don't exist
func initDatabase(db *sql.DB) error {
	// Create tables
	if err := CreateMangaTable(db); err != nil {
		return fmt.Errorf("failed to create manga table: %w", err)
	}

	return nil
}

// CreateMangaTable creates the manga table if it doesn't exist
func CreateMangaTable(db *sql.DB) error {
	// Create enum types
	_, err := db.Exec(`
		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'manga_status') THEN
				CREATE TYPE manga_status AS ENUM ('Unknown', 'Ongoing', 'Completed', 'Cancelled', 'Hiatus');
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'manga_format') THEN
				CREATE TYPE manga_format AS ENUM ('manga', 'novel', 'one_shot');
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to create enum types: %w", err)
	}

	// Create series table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS series (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description_formatted TEXT,
			description_text TEXT,
			publisher VARCHAR(255),
			status manga_status DEFAULT 'Unknown',
			year INTEGER,
			total_chapters INTEGER,
			total_issues INTEGER,
			book_type VARCHAR(50),
			comic_image TEXT,
			comic_id INTEGER,
			publication_run VARCHAR(100),
			volumes INTEGER DEFAULT 0,
			chapters INTEGER DEFAULT 0,
			average_score INTEGER,
			popularity INTEGER,
			mean_score INTEGER,
			is_licensed BOOLEAN DEFAULT false,
			updated_at BIGINT,
			banner_image TEXT,
			format manga_format,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT unique_name UNIQUE (name)
		);

		CREATE TABLE IF NOT EXISTS covers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			extra_large TEXT,
			large TEXT,
			medium TEXT,
			color VARCHAR(7),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS genres (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL,
			CONSTRAINT unique_genre UNIQUE (name)
		);

		CREATE TABLE IF NOT EXISTS series_genres (
			series_id UUID REFERENCES series(id) ON DELETE CASCADE,
			genre_id UUID REFERENCES genres(id) ON DELETE CASCADE,
			PRIMARY KEY (series_id, genre_id)
		);

		CREATE TABLE IF NOT EXISTS tags (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			CONSTRAINT unique_tag UNIQUE (name)
		);

		CREATE TABLE IF NOT EXISTS series_tags (
			series_id UUID REFERENCES series(id) ON DELETE CASCADE,
			tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (series_id, tag_id)
		);

		CREATE TABLE IF NOT EXISTS characters (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			CONSTRAINT unique_character UNIQUE (name)
		);

		CREATE TABLE IF NOT EXISTS series_characters (
			series_id UUID REFERENCES series(id) ON DELETE CASCADE,
			character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
			PRIMARY KEY (series_id, character_id)
		);

		CREATE TABLE IF NOT EXISTS urls (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS synonyms (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		if err.Error() == "pq: relation \"series\" already exists" {
			return nil
		}
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}
