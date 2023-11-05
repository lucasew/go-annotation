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

func stringOr(str, or string) string {
	if str != "" {
		return str
	} else {
		return or
	}
}

func pathParts(path string) []string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func (a *AnnotatorApp) GetTask(taskID string) *ConfigTask {
	for _, currentTask := range a.Config.Tasks {
		if currentTask.ID == taskID {
			return currentTask
		}
	}
	return nil
}

func (a *AnnotatorApp) GetHTTPHandler() http.Handler {
	a.init()
	mux := http.NewServeMux()
	mux.HandleFunc("/help/", func(w http.ResponseWriter, r *http.Request) {
		itemPath := pathParts(r.URL.Path)
		var markdownBuilder strings.Builder
		title := "Help"
		fmt.Fprintf(&markdownBuilder, "# [<](/) Project help\n")
		fmt.Fprintf(&markdownBuilder, "## Description\n")
		fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(a.Config.Meta.Description, "(No description provided)"), "\n", "\n>"))
		if len(itemPath) == 1 {
			fmt.Fprintf(&markdownBuilder, "## Phases\n\n")
			for _, task := range a.Config.Tasks {
				fmt.Fprintf(&markdownBuilder, "### [%s](/help/%s)\n", task.ShortName, task.ID)
				fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(task.Name, "(No description provided)"), "\n", "\n>"))
			}
		} else if len(itemPath) == 2 {
			helpTask := itemPath[1]
			task := a.GetTask(helpTask)
			if task == nil {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}
			fmt.Fprintf(&markdownBuilder, "## Phase: %s\n", task.ShortName)
			fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(task.Name, "(No description provided)"), "\n", "\n>"))

			fmt.Fprintf(&markdownBuilder, "### Possible choices\n")
			for k, v := range task.Classes {
				fmt.Fprintf(&markdownBuilder, "#### %s (%s)\n", stringOr(v.Name, "(No name)"), k)

				fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(v.Description, "(No description provided)"), "\n", "\n>"))
				if v.Examples != nil && len(v.Examples) > 0 {

					fmt.Fprintf(&markdownBuilder, "##### Examples\n")
					for _, example := range v.Examples {
						fmt.Fprintf(&markdownBuilder, "![](/asset/%s)", example)
					}
				}
			}
		} else {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		ExecTemplate(w, TemplateContent{Title: title, Content: markdownBuilder.String()})
	})

	mux.HandleFunc("/annotate/", func(w http.ResponseWriter, r *http.Request) {
		var markdownBuilder strings.Builder
		itemPath := pathParts(r.URL.Path)

		if len(itemPath) == 2 {
			fmt.Fprintf(&markdownBuilder, "**Two items**")
		}
		if len(itemPath) == 3 {
			fmt.Fprintf(&markdownBuilder, "**Three items**")
		}

		fmt.Fprintf(&markdownBuilder, "[<](/)")
		for _, part := range itemPath {
			fmt.Fprintf(&markdownBuilder, "\n\n- **%s**\n", part)
		}
		ExecTemplate(w, TemplateContent{Title: "annotation", Content: markdownBuilder.String()})
	})
	mux.HandleFunc("/asset/", func(w http.ResponseWriter, r *http.Request) {
		itemPath := pathParts(r.URL.Path)
		if len(itemPath) != 2 {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		image_id := itemPath[1]
		log.Printf("http: fetching asset id %s", image_id)

		rows, err := a.Database.QueryContext(r.Context(), "select filename from images where sha256 = ? limit 1", image_id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error: while querying for the filename of hash %s: %s", image_id, err)
			return
		}
		if !rows.Next() {
			log.Printf("http: asset id %s was not found in the database", image_id)
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		var filename string
		rows.Scan(&filename)
		log.Printf("http: asset id %s is %s!", image_id, filename)
		// TODO: fetch image filename from database
		f, err := os.Open(path.Join(a.ImagesDir, filename))
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var markdownBuilder strings.Builder
		fmt.Fprintf(&markdownBuilder, "# Welcome to go-annotator\n")
		fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(a.Config.Meta.Description, "\n", "\n>"))
		fmt.Fprintf(&markdownBuilder, "\n\n")
		fmt.Fprintf(&markdownBuilder, "[Annotation instructions](/help) ")
		fmt.Fprintf(&markdownBuilder, "[Continue annotations](/annotate)")

		ExecTemplate(w, TemplateContent{Title: "Welcome", Content: markdownBuilder.String()})
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
insert into images (filename, sha256) values (?, ?) on conflict(sha256) do update set filename=excluded.filename
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
    user text,
    value text,
    sure int, -- 0 = not sure, 1 = sure
    foreign key(image) references images(sha256)
);

`, task.ID))
		if err != nil {
			return fmt.Errorf("while creating task database for task '%s': %w", task.ID, err)
		}
		_, err = tx.Exec(fmt.Sprintf(`
insert into task_%s (image) select sha256 image from images
`, task.ID))

		if err != nil {
			return fmt.Errorf("while seeding task database for task '%s': %w", task.ID, err)
		}
	}

	log.Printf("PrepareDatabase: success! commiting transaction to the database")
	return tx.Commit()
}
