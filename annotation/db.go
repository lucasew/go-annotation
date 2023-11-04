package annotation

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func GetDatabase(filename string) (*sql.DB, error) {
	return sql.Open("sqlite", filename)
}
