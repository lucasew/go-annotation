package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create schema
	schema := `
CREATE TABLE images (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT UNIQUE NOT NULL,
  original_filename TEXT,
  ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_stages INTEGER DEFAULT 0,
  is_finished BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_images_is_finished ON images(is_finished);
CREATE INDEX idx_images_completed_stages ON images(completed_stages);

CREATE TABLE annotations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  image_id INTEGER NOT NULL,
  username TEXT NOT NULL,
  stage_index INTEGER NOT NULL,
  option_value TEXT NOT NULL,
  annotated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(image_id, username, stage_index),
  FOREIGN KEY(image_id) REFERENCES images(id) ON DELETE CASCADE
);

CREATE INDEX idx_annotations_image_id ON annotations(image_id);
CREATE INDEX idx_annotations_username ON annotations(username);
CREATE INDEX idx_annotations_stage ON annotations(stage_index);
`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// CleanupTestDB closes the test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("failed to close test database: %v", err)
	}
}

// MustExec executes a SQL statement and fails the test if it errors
func MustExec(t *testing.T, db *sql.DB, query string, args ...interface{}) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("failed to exec query: %v", err)
	}
}
