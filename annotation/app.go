package annotation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type AnnotatorApp struct {
	ImagesDir string
	Database  *sql.DB
	Config    *Config
}

func (a *AnnotatorApp) init() {
	if a.ImagesDir[len(a.ImagesDir)-1] == '/' {
		a.ImagesDir = a.ImagesDir[:len(a.ImagesDir)-1]
	}
}

func (a *AnnotatorApp) GetHTTPHandler() http.Handler {
	a.init()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Template.Execute(w, nil)
	})

	mux.HandleFunc("/asset/", func(w http.ResponseWriter, r *http.Request) {
		itemPath := strings.Split(r.URL.Path, "/")
		if len(itemPath) != 3 {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		image_id := itemPath[2]
		log.Printf("http: fetching asset %s", image_id)
		// TODO: fetch image filename from database
		f, err := os.Open(path.Join(a.ImagesDir, image_id))
		defer f.Close()
		if errors.Is(err, os.ErrNotExist) {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error: http: while serving image asset: %s", err)
			return
		}
		io.Copy(w, f)
	})
	log.Printf("images dir: %s", a.ImagesDir)

	var handler http.Handler = mux
	handler = HTTPLogger(handler)
	return handler
}

func (a *AnnotatorApp) PrepareDatabase(ctx context.Context) error {
	a.init()
	log.Printf("PrepareDatabase: starting transaction")
	tx, err := a.Database.BeginTx(ctx, &sql.TxOptions{})
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
	err = filepath.WalkDir(a.ImagesDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == a.ImagesDir {
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
	if err != nil {
		return err
	}
	log.Printf("PrepareDatabase: create index for hashes to image paths") // TODO: check later if it's actually premature optimization
	_, err = tx.Exec(`
        create unique index if not exists images_hash_idx on images(sha256)
    `)
	if err != nil {
		return fmt.Errorf("while creating index for image hashes: %w", err)
	}

	log.Printf("PrepareDatabase: creating tables for each task")
	for _, task := range a.Config.Tasks {
		_, err := tx.Exec(fmt.Sprintf(`
create table if not exists task_%s (
    image text not null,
    user text not null,
    value text,
    sure int, -- 0 = not sure, 1 = sure
    foreign key(image) references images(sha256)
)`, task.ID))
		if err != nil {
			return fmt.Errorf("while creating task database for task '%s': %w", task.ID, err)
		}
	}

	log.Printf("PrepareDatabase: success! commiting transaction to the database")
	return tx.Commit()
}
