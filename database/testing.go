package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/metafates/mangal/db"
	_ "github.com/lib/pq"
)

// TestDB represents a test database connection
type TestDB struct {
	*sql.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Get test database connection string from environment variable
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/mangal_test?sslmode=disable"
	}

	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return &TestDB{DB: db}
}

// Close closes the test database connection
func (tdb *TestDB) Close(t *testing.T) {
	t.Helper()
	if err := tdb.DB.Close(); err != nil {
		t.Errorf("Failed to close test database connection: %v", err)
	}
}

// SetupTestDB sets up a test database with the schema
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	// Get database connection
	testDB, err := db.GetDB()
	if err != nil {
		t.Fatalf("Failed to get test database connection: %v", err)
	}

	// Run migrations
	if err := db.Migrate(testDB, true); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestDB{DB: testDB}
}

// TearDownTestDB tears down the test database
func TearDownTestDB(t *testing.T, tdb *TestDB) {
	t.Helper()

	// Drop all tables
	_, err := tdb.Exec(`
		DROP TABLE IF EXISTS synonyms;
		DROP TABLE IF EXISTS urls;
		DROP TABLE IF EXISTS series_characters;
		DROP TABLE IF EXISTS characters;
		DROP TABLE IF EXISTS series_tags;
		DROP TABLE IF EXISTS tags;
		DROP TABLE IF EXISTS series_genres;
		DROP TABLE IF EXISTS genres;
		DROP TABLE IF EXISTS covers;
		DROP TABLE IF EXISTS series;
	`)
	if err != nil {
		t.Errorf("Failed to drop tables: %v", err)
	}

	// Close the connection
	tdb.Close(t)
}

// CleanTestData cleans all test data from the database
func CleanTestData(t *testing.T, tdb *TestDB) {
	t.Helper()

	// Clean all tables
	tables := []string{
		"synonyms",
		"urls",
		"series_characters",
		"characters",
		"series_tags",
		"tags",
		"series_genres",
		"genres",
		"covers",
		"series",
	}

	for _, table := range tables {
		_, err := tdb.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Errorf("Failed to clean table %s: %v", table, err)
		}
	}
}
