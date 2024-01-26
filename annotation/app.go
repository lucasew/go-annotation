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
	"sort"
	"strings"

	"math/rand"
)

type AnnotatorApp struct {
	ImagesDir     string
	Database      *sql.DB
	Config        *Config
	OffsetAdvance int
	i18n          map[string]string
}

func (a *AnnotatorApp) init() {
	if a.ImagesDir[len(a.ImagesDir)-1] == '/' {
		a.ImagesDir = a.ImagesDir[:len(a.ImagesDir)-1]
	}
	if a.OffsetAdvance == 0 {
		a.OffsetAdvance = 10
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

type AnnotationStep struct {
	TaskID    string
	ImageID   string
	ImageName string
}

func (a *AnnotatorApp) NextAnnotationStep(ctx context.Context, taskID string) (*AnnotationStep, error) {
	if taskID == "" {
		for _, task := range a.Config.Tasks {
			step, err := a.NextAnnotationStep(ctx, task.ID)
			if err != nil {
				return nil, err
			}
			if step == nil {
				continue
			}
			return step, nil
		}
		return nil, nil
	}
	task := a.GetTask(taskID)
	filters := ""
	for criteriaTask, criteriaValue := range task.If {
		filters = fmt.Sprintf("and (image in (select image from task_%s where value = '%s' and sure = 1 ))", criteriaTask, criteriaValue)
	}

	ret := AnnotationStep{TaskID: taskID}
	rows, err := a.Database.QueryContext(ctx, fmt.Sprintf("select image from task_%s where value is NULL %s", task.ID, filters))
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("while fetching pending tasks: %w", err)
	}
	imageIDs := []string{}
	for i := 0; i < a.OffsetAdvance; i++ {
		if !rows.Next() {
			break
		}
		var imageID string
		err = rows.Scan(&imageID)
		if err != nil {
			return nil, err
		}
		imageIDs = append(imageIDs, imageID)
	}
	if len(imageIDs) == 0 {
		rows, err = a.Database.QueryContext(ctx, fmt.Sprintf("select image from task_%s where sure != 1 %s", task.ID, filters))
		defer rows.Close()
		if err != nil {
			return nil, fmt.Errorf("while fetching doubtful tasks: %w", err)
		}
		for i := 0; i < a.OffsetAdvance; i++ {
			if !rows.Next() {
				break
			}
			var imageID string
			err = rows.Scan(&imageID)
			if err != nil {
				return nil, err
			}
			imageIDs = append(imageIDs, imageID)
		}
	}
	if len(imageIDs) > 0 {
		ret.ImageID = imageIDs[rand.Intn(len(imageIDs))]
		return &ret, nil
	}
	return nil, nil
}

func (a *AnnotatorApp) GetFilenameFromHash(ctx context.Context, hash string) (filename string, err error) {
	rows, err := a.Database.QueryContext(ctx, "select filename from images where sha256 = ?", hash)
	defer rows.Close()
	if err != nil {
		return "", err
	}
	if !rows.Next() {
		return "", fmt.Errorf("No filename found for hash")
	}
	err = rows.Scan(&filename)
	return
}

type AnnotationResponse struct {
	ImageID string
	TaskID  string
	User    string
	Value   string
	Sure    bool
}

func (a *AnnotatorApp) SubmitAnnotation(ctx context.Context, annotation AnnotationResponse) error {
	tx, err := a.Database.BeginTx(ctx, &sql.TxOptions{})
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("while starting transaction: %w", err)
	}
	task := a.GetTask(annotation.TaskID)
	if task == nil {
		return fmt.Errorf("no such task") // did you check for the task before calling this?
	}
	_, err = tx.Exec(
		fmt.Sprintf("update task_%s set user = ?, value = ?, sure = ? where image == ?", task.ID),
		annotation.User, annotation.Value, annotation.Sure, annotation.ImageID,
	)
	if err != nil {
		return fmt.Errorf("while updating value of the annotation: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("while commiting transaction: %w", err)
	}
	return nil
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
		fmt.Fprintf(&markdownBuilder, "# [<](/) %s\n", i("Project help"))
		fmt.Fprintf(&markdownBuilder, "## %s\n", i("Description"))
		fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(a.Config.Meta.Description, i("(No description provided)")), "\n", "\n>"))
		if len(itemPath) == 1 {
			fmt.Fprintf(&markdownBuilder, "## %s\n\n", i("Phases"))
			for _, task := range a.Config.Tasks {
				fmt.Fprintf(&markdownBuilder, "### [%s](/help/%s)\n", task.ShortName, task.ID)
				fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(task.Name, i("(No description provided)")), "\n", "\n>"))
			}
		} else if len(itemPath) == 2 {
			helpTask := itemPath[1]
			task := a.GetTask(helpTask)
			if task == nil {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}
			fmt.Fprintf(&markdownBuilder, "## %s: %s\n", i("Phase"), task.ShortName)
			fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(task.Name, "(No description provided)"), "\n", "\n>"))

			fmt.Fprintf(&markdownBuilder, "### %s\n", i("Possible choices"))
			for k, v := range task.Classes {
				fmt.Fprintf(&markdownBuilder, "#### %s (%s)\n", i(stringOr(v.Name, "(No name)")), k)

				fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(stringOr(v.Description, i("(No description provided)")), "\n", "\n>"))
				if v.Examples != nil && len(v.Examples) > 0 {

					fmt.Fprintf(&markdownBuilder, "##### %s \n", i("Examples"))
					for _, example := range v.Examples {
						fmt.Fprintf(&markdownBuilder, "\n\n![](/asset/%s)", example)
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

		if len(itemPath) != 3 {
			step, err := a.NextAnnotationStep(r.Context(), "")
			if err != nil {
				log.Printf("error in annotate when getting next step from scratch: %s", err)
				w.WriteHeader(500)
				return
			}
			if step == nil {
				fmt.Fprintf(&markdownBuilder, "# %s!\n", i("Congratulations"))
				fmt.Fprintf(&markdownBuilder, "%s!\n\n", i("All annotations are done"))
				fmt.Fprintf(&markdownBuilder, "[%s](/)\n", i("Go to the main page"))
				ExecTemplate(w, TemplateContent{Title: i("All annotations are done!"), Content: markdownBuilder.String()})
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/annotate/%s/%s", step.TaskID, step.ImageID), http.StatusSeeOther)

			return
		}

		taskID := itemPath[1]
		imageID := itemPath[2]
		task := a.GetTask(taskID)
		if task == nil {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		filename, _ := a.GetFilenameFromHash(r.Context(), imageID)

		if r.Method == http.MethodPost {
			log.Printf("POST")
			r.ParseForm()
			if !(r.Form.Has("selectedClass") && r.Form.Has("sure")) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			selectedClass := r.FormValue("selectedClass")
			_, isClassValid := task.Classes[selectedClass]
			log.Printf("Selected class: %s empty=%v valid=%v", selectedClass, selectedClass == "", isClassValid)
			sure := r.FormValue("sure") == "on"
			log.Printf("Sure: %v", sure)
			user, _, _ := r.BasicAuth()
			err := a.SubmitAnnotation(r.Context(), AnnotationResponse{
				ImageID: imageID,
				TaskID:  taskID,
				User:    user,
				Value:   selectedClass,
				Sure:    sure,
			})
			if err != nil {
				log.Printf("error while submitting annotation: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			step, err := a.NextAnnotationStep(r.Context(), taskID)
			if err != nil {
				log.Printf("error while getting next step: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if step == nil {
				step, err = a.NextAnnotationStep(r.Context(), "")
				if err != nil {
					log.Printf("error while getting next step at the end of task: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			if step == nil {
				w.Header().Add("HX-Redirect", "/")
			} else if step.TaskID != taskID {
				w.Header().Add("HX-Redirect", fmt.Sprintf("/help/%s", step.TaskID))
			} else {
				w.Header().Add("HX-Redirect", fmt.Sprintf("/annotate/%s/%s", taskID, step.ImageID))
			}
			return
		}
		classNames := make([]string, 0, len(task.Classes))
		for class := range task.Classes {
			classNames = append(classNames, class)
		}
		sort.Sort(sort.StringSlice(classNames))

		fmt.Fprintf(&markdownBuilder, "# [<](/) %s [???](/help/%s) \n", task.Name, task.ID)
		thisURL := fmt.Sprintf("/annotate/%s/%s", taskID, imageID)

		fmt.Fprintf(&markdownBuilder, `<div style="display: flex; flex-wrap: wrap; flex: 1; justify-content: space-between">`)
		for _, class := range classNames {
			classMeta := task.Classes[class]
			fmt.Fprintf(&markdownBuilder, `<button style="display: flex; flex: 1" hx-post="%s" data_selectedClass="%s" data_sure="on" hx-vals='js:{selectedClass: event.target.attributes.data_selectedClass.value, sure: event.target.attributes.data_sure.value}'>%s</button>`, thisURL, class, i(classMeta.Name))
		}
		fmt.Fprintf(&markdownBuilder, `<button style="display: flex; flex: 1" hx-post="%s" data_selectedClass="" data_sure="off" hx-vals='js:{selectedClass: event.attributes.data_selectedClass.value, sure: event.attributes.data_sure.value}'>???</button>`, thisURL)

		fmt.Fprintf(&markdownBuilder, `</div>`)

		fmt.Fprintf(&markdownBuilder, `<p id="image_id" hx-on:click="navigator.clipboard.writeText(this.innerText); alert('%s')" style="font-family: monospace; overflow-x: hidden; text-align: center;">%s</p>`, i("Copied to clipboard!"), filename)

		fmt.Fprintf(&markdownBuilder, "\n\n![](/asset/%s)", imageID)

		ExecTemplate(w, TemplateContent{Title: i("annotation"), Content: markdownBuilder.String()})
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
		fmt.Fprintf(&markdownBuilder, "# %s\n", i("Welcome to go-annotator"))
		fmt.Fprintf(&markdownBuilder, "> %s\n\n", strings.ReplaceAll(a.Config.Meta.Description, "\n", "\n>"))
		fmt.Fprintf(&markdownBuilder, "\n\n")
		fmt.Fprintf(&markdownBuilder, "[%s](/help) ", i("Annotation instructions"))
		fmt.Fprintf(&markdownBuilder, "[%s](/annotate)", i("Continue annotations"))

		ExecTemplate(w, TemplateContent{Title: i("Welcome"), Content: markdownBuilder.String()})
	})

	log.Printf("images dir: %s", a.ImagesDir)

	var handler http.Handler = mux
	handler = HTTPLogger(handler)
	handler = a.authenticationMiddleware(handler)
	return handler
}

func (a *AnnotatorApp) authenticationMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			var item *ConfigAuth = nil
			item, ok = a.Config.Authentication[username]
			if ok {
				if password == item.Password {
					log.Printf("auth for user %s: success", username)
					handler.ServeHTTP(w, r)
					return
				}
				log.Printf("auth for user %s: bad password", username)
			} else {

				log.Printf("auth for user %s: no such user", username)
			}
		}
		log.Printf("auth: not ok")
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		w.WriteHeader(http.StatusUnauthorized)
	})
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
    image text not null unique,
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
insert or ignore into task_%s (image) select sha256 image from images
`, task.ID))

		if err != nil {
			return fmt.Errorf("while seeding task database for task '%s': %w", task.ID, err)
		}
	}

	log.Printf("PrepareDatabase: success! commiting transaction to the database")
	return tx.Commit()
}
