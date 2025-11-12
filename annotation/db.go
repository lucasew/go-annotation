package annotation

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func GetDatabase(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrency (allows reads during writes)
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		db.Close()
		return nil, err
	}

	// Set busy timeout to 5 seconds (wait instead of immediately failing with SQLITE_BUSY)
	_, err = db.Exec("PRAGMA busy_timeout=5000")
	if err != nil {
		db.Close()
		return nil, err
	}

	// Configure connection pool for SQLite
	// SQLite can only handle one writer at a time, so we limit to 1 open connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db, nil
}
