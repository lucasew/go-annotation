package annotation

import (
    _ "modernc.org/sqlite"
    "io/fs"
    "database/sql"
    "fmt"
    "context"
    "path/filepath"
    "log"
)

func GetDatabase(filename string) (*sql.DB, error) {
    return sql.Open("sqlite", filename)
}

func PrepareDatabase(ctx context.Context, db *sql.DB, config *Config, imageFolder string) error {
    log.Printf("PrepareDatabase: starting transaction")
    tx, err := db.BeginTx(ctx, &sql.TxOptions{})
    defer tx.Rollback()
    if err != nil {
        return fmt.Errorf("while starting database setup transaction: %w", err)
    }
    log.Printf("PrepareDatabase: setting up images table")
    _, err = tx.Exec(`
create table if not exists images (
sha256 text unique primary key,
filename text not null
)
    `)
    if err != nil {
        return fmt.Errorf("while creating table 'images': %w", err)
    }
    log.Printf("PrepareDatabase: populating images table")
    err = filepath.WalkDir(imageFolder, func(path string, info fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if path == imageFolder {
            return nil
        }
        if info.IsDir() {
            return fmt.Errorf("while checking if item '%s' is a file: datasets must be organized in a flat folder structure. Hint: use the 'ingest' subcommand.", path)
        }
        log.Printf("PrepareDatabase: populating images table: %s", path)
        _, err = DecodeImage(path)
        if err != nil {
            return fmt.Errorf("while checking if item '%s' is an image: %w", path, err)
        }
        fileHash, err := HashFile(path)
        if err != nil {
            return fmt.Errorf("while hashing item '%s': %w", path, err)
        }
        _, err = tx.Exec(`
insert into images (sha256, filename) values (?, ?) on conflict(sha256) do update set filename=excluded.filename
        `, info.Name(), fileHash)
        return err
    })
    if err != nil { return err }

    log.Printf("PrepareDatabase: success! commiting transaction to the database")
    return tx.Commit()
}
