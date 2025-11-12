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

	"github.com/lucasew/go-annotation/internal/repository"
)

type AnnotatorApp struct {
	ImagesDir          string
	Database           *sql.DB
	Config             *Config
	OffsetAdvance      int
	i18n               map[string]string
	imageRepo          *repository.ImageRepository
	annotationRepo     *repository.AnnotationRepository
}

func (a *AnnotatorApp) init() {
	if a.ImagesDir[len(a.ImagesDir)-1] == '/' {
		a.ImagesDir = a.ImagesDir[:len(a.ImagesDir)-1]
	}
	if a.OffsetAdvance == 0 {
		a.OffsetAdvance = 10
	}
	// Initialize repositories
	a.imageRepo = repository.NewImageRepository(a.Database)
	a.annotationRepo = repository.NewAnnotationRepository(a.Database)
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

type TaskWithCount struct {
	*ConfigTask
	AvailableCount int
	TotalCount     int
	CompletedCount int
	PhaseProgress  *PhaseProgress
}

type PhaseProgress struct {
	Completed            int     // Images completed in this phase
	Pending              int     // Images eligible but not yet annotated
	FilteredWrongClass   int     // Images annotated in dependency phase but with wrong class
	NotYetAnnotated      int     // Images not yet annotated in dependency phase
	Total                int     // Total images in the entire dataset
	CompletedPercent     float64 // Percentage of completed images
	PendingPercent       float64 // Percentage of pending images
	FilteredPercent      float64 // Percentage of filtered (wrong class) images
	NotYetAnnotatedPercent float64 // Percentage of not yet annotated images
}

// CountEligibleImages counts all images that are eligible for this task (regardless of annotation status)
func (a *AnnotatorApp) CountEligibleImages(ctx context.Context, taskID string) (int, error) {
	// Find stage index for this task
	stageIndex := -1
	for i, task := range a.Config.Tasks {
		if task.ID == taskID {
			stageIndex = i
			break
		}
	}
	if stageIndex == -1 {
		return 0, fmt.Errorf("task not found: %s", taskID)
	}

	task := a.Config.Tasks[stageIndex]

	// If no dependencies, all images are eligible
	if len(task.If) == 0 {
		count, err := a.imageRepo.Count(ctx)
		return int(count), err
	}

	// Get all images and filter by dependencies
	allImages, err := a.imageRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("while listing images: %w", err)
	}

	validCount := 0
	for _, img := range allImages {
		valid := true
		// Check each dependency
		for depTaskID, requiredValue := range task.If {
			// Find the stage index for the dependency task
			depStageIndex := -1
			for i, t := range a.Config.Tasks {
				if t.ID == depTaskID {
					depStageIndex = i
					break
				}
			}
			if depStageIndex == -1 {
				continue
			}

			// Check if this image has the required annotation value for the dependency
			hasValidAnnotation := false
			imageHashes, err := a.annotationRepo.GetImageHashesWithAnnotation(ctx, int64(depStageIndex), requiredValue)
			if err != nil {
				return 0, fmt.Errorf("while checking dependency: %w", err)
			}
			for _, hash := range imageHashes {
				if hash == img.SHA256 {
					hasValidAnnotation = true
					break
				}
			}

			if !hasValidAnnotation {
				valid = false
				break
			}
		}

		if valid {
			validCount++
		}
	}

	return validCount, nil
}

func (a *AnnotatorApp) CountAvailableImages(ctx context.Context, taskID string) (int, error) {
	// Find stage index for this task
	stageIndex := -1
	for i, task := range a.Config.Tasks {
		if task.ID == taskID {
			stageIndex = i
			break
		}
	}
	if stageIndex == -1 {
		return 0, fmt.Errorf("task not found: %s", taskID)
	}

	task := a.Config.Tasks[stageIndex]

	// Count images without annotation for this stage
	count, err := a.annotationRepo.CountImagesWithoutAnnotationForStage(ctx, int64(stageIndex))
	if err != nil {
		return 0, fmt.Errorf("while counting available images: %w", err)
	}

	// Handle task dependencies (If field)
	// If there are dependencies, we need to filter images that meet the criteria
	if len(task.If) > 0 {
		// Get all candidate images
		allImages, err := a.imageRepo.List(ctx)
		if err != nil {
			return 0, fmt.Errorf("while listing images: %w", err)
		}

		validCount := 0
		for _, img := range allImages {
			valid := true
			// Check each dependency
			for depTaskID, requiredValue := range task.If {
				// Find the stage index for the dependency task
				depStageIndex := -1
				for i, t := range a.Config.Tasks {
					if t.ID == depTaskID {
						depStageIndex = i
						break
					}
				}
				if depStageIndex == -1 {
					continue
				}

				// Check if this image has the required annotation value for the dependency
				hasValidAnnotation := false
				// We need to check if ANY user has annotated this image with the required value
				imageHashes, err := a.annotationRepo.GetImageHashesWithAnnotation(ctx, int64(depStageIndex), requiredValue)
				if err != nil {
					return 0, fmt.Errorf("while checking dependency: %w", err)
				}
				for _, hash := range imageHashes {
					if hash == img.SHA256 {
						hasValidAnnotation = true
						break
					}
				}

				if !hasValidAnnotation {
					valid = false
					break
				}
			}

			if valid {
				// Check if this image has annotation for current stage
				hasAnnotation, err := a.annotationRepo.CheckAnnotationExists(ctx, img.SHA256, "", int64(stageIndex))
				if err != nil {
					return 0, err
				}
				if !hasAnnotation {
					validCount++
				}
			}
		}
		return validCount, nil
	}

	return int(count), nil
}

// GetPhaseProgressStats calculates comprehensive progress statistics for a task
func (a *AnnotatorApp) GetPhaseProgressStats(ctx context.Context, taskID string) (*PhaseProgress, error) {
	// Get total images in the entire dataset
	totalCount, err := a.imageRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("while counting total images: %w", err)
	}

	// Get eligible images (that pass filters from previous phases)
	eligibleCount, err := a.CountEligibleImages(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("while counting eligible images: %w", err)
	}

	// Get available images (eligible but not yet annotated)
	availableCount, err := a.CountAvailableImages(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("while counting available images: %w", err)
	}

	// Calculate completed and pending
	completed := eligibleCount - availableCount
	if completed < 0 {
		completed = 0
	}
	pending := availableCount

	total := int(totalCount)
	notEligible := total - eligibleCount

	// Now differentiate between filtered (annotated with wrong class) and not yet annotated
	var filteredWrongClass, notYetAnnotated int

	// Find task and check if it has dependencies
	stageIndex := -1
	for i, task := range a.Config.Tasks {
		if task.ID == taskID {
			stageIndex = i
			break
		}
	}

	if stageIndex != -1 {
		task := a.Config.Tasks[stageIndex]

		// If task has dependencies, analyze the not-eligible images
		if len(task.If) > 0 {
			// Get all images
			allImages, err := a.imageRepo.List(ctx)
			if err != nil {
				return nil, fmt.Errorf("while listing images: %w", err)
			}

			// Get images that passed the filter (eligible)
			eligibleHashes := make(map[string]bool)
			for _, img := range allImages {
				valid := true
				for depTaskID, requiredValue := range task.If {
					depStageIndex := -1
					for i, t := range a.Config.Tasks {
						if t.ID == depTaskID {
							depStageIndex = i
							break
						}
					}
					if depStageIndex == -1 {
						continue
					}

					imageHashes, err := a.annotationRepo.GetImageHashesWithAnnotation(ctx, int64(depStageIndex), requiredValue)
					if err != nil {
						return nil, fmt.Errorf("while checking dependency: %w", err)
					}

					hasValidAnnotation := false
					for _, hash := range imageHashes {
						if hash == img.SHA256 {
							hasValidAnnotation = true
							break
						}
					}

					if !hasValidAnnotation {
						valid = false
						break
					}
				}

				if valid {
					eligibleHashes[img.SHA256] = true
				}
			}

			// Check not-eligible images to see if they were annotated in dependency phase
			for _, img := range allImages {
				if !eligibleHashes[img.SHA256] {
					// This image is not eligible - check if it was annotated in dependency phase
					annotatedInDep := false
					for depTaskID := range task.If {
						depStageIndex := -1
						for i, t := range a.Config.Tasks {
							if t.ID == depTaskID {
								depStageIndex = i
								break
							}
						}
						if depStageIndex == -1 {
							continue
						}

						// Check if this image has ANY annotation in the dependency phase
						hasAnnotation, err := a.annotationRepo.CheckAnnotationExists(ctx, img.SHA256, "", int64(depStageIndex))
						if err == nil && hasAnnotation {
							annotatedInDep = true
							break
						}
					}

					if annotatedInDep {
						filteredWrongClass++
					} else {
						notYetAnnotated++
					}
				}
			}
		} else {
			// No dependencies, so all not-eligible images are "not yet annotated"
			notYetAnnotated = notEligible
		}
	}

	// Calculate percentages
	var completedPercent, pendingPercent, filteredPercent, notYetAnnotatedPercent float64
	if total > 0 {
		completedPercent = float64(completed) / float64(total) * 100
		pendingPercent = float64(pending) / float64(total) * 100
		filteredPercent = float64(filteredWrongClass) / float64(total) * 100
		notYetAnnotatedPercent = float64(notYetAnnotated) / float64(total) * 100
	}

	return &PhaseProgress{
		Completed:              completed,
		Pending:                pending,
		FilteredWrongClass:     filteredWrongClass,
		NotYetAnnotated:        notYetAnnotated,
		Total:                  total,
		CompletedPercent:       completedPercent,
		PendingPercent:         pendingPercent,
		FilteredPercent:        filteredPercent,
		NotYetAnnotatedPercent: notYetAnnotatedPercent,
	}, nil
}

func (a *AnnotatorApp) NextAnnotationStep(ctx context.Context, taskID string) (*AnnotationStep, error) {
	// If no task specified, try each task in order
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

	// Find stage index for this task
	stageIndex := -1
	for i, task := range a.Config.Tasks {
		if task.ID == taskID {
			stageIndex = i
			break
		}
	}
	if stageIndex == -1 {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	task := a.Config.Tasks[stageIndex]

	// Get images without annotation for this stage
	allImages, err := a.imageRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("while listing images: %w", err)
	}

	// Filter images based on dependencies and annotation status
	var candidateImages []string
	for _, img := range allImages {
		// Check if image already has annotation for this stage
		hasAnnotation, err := a.annotationRepo.CheckAnnotationExists(ctx, img.SHA256, "", int64(stageIndex))
		if err != nil {
			return nil, err
		}
		if hasAnnotation {
			continue // Skip images that already have annotation
		}

		// Check task dependencies (If field)
		valid := true
		if len(task.If) > 0 {
			for depTaskID, requiredValue := range task.If {
				// Find the stage index for the dependency task
				depStageIndex := -1
				for i, t := range a.Config.Tasks {
					if t.ID == depTaskID {
						depStageIndex = i
						break
					}
				}
				if depStageIndex == -1 {
					continue
				}

				// Check if this image has the required annotation value for the dependency
				imageHashes, err := a.annotationRepo.GetImageHashesWithAnnotation(ctx, int64(depStageIndex), requiredValue)
				if err != nil {
					return nil, fmt.Errorf("while checking dependency: %w", err)
				}

				hasValidAnnotation := false
				for _, hash := range imageHashes {
					if hash == img.SHA256 {
						hasValidAnnotation = true
						break
					}
				}

				if !hasValidAnnotation {
					valid = false
					break
				}
			}
		}

		if valid {
			candidateImages = append(candidateImages, img.SHA256)
			// Limit candidates to OffsetAdvance for performance
			if len(candidateImages) >= a.OffsetAdvance {
				break
			}
		}
	}

	// No images available
	if len(candidateImages) == 0 {
		return nil, nil
	}

	// Randomly select one image SHA256
	selectedSHA256 := candidateImages[rand.Intn(len(candidateImages))]

	// Get image details
	selectedImage, err := a.imageRepo.GetBySHA256(ctx, selectedSHA256)
	if err != nil {
		return nil, fmt.Errorf("while getting image details: %w", err)
	}

	return &AnnotationStep{
		TaskID:    taskID,
		ImageID:   selectedSHA256,
		ImageName: selectedImage.Filename,
	}, nil
}

func (a *AnnotatorApp) GetImageFilename(ctx context.Context, sha256 string) (filename string, err error) {
	// Get image from repository using SHA256 hash
	img, err := a.imageRepo.GetBySHA256(ctx, sha256)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("image not found: %s", sha256)
		}
		return "", err
	}

	return img.Filename, nil
}

type AnnotationResponse struct {
	ImageID string
	TaskID  string
	User    string
	Value   string
	Sure    bool
}

func (a *AnnotatorApp) SubmitAnnotation(ctx context.Context, annotation AnnotationResponse) error {
	// Find stage index for this task
	stageIndex := -1
	for i, task := range a.Config.Tasks {
		if task.ID == annotation.TaskID {
			stageIndex = i
			break
		}
	}
	if stageIndex == -1 {
		return fmt.Errorf("no such task: %s", annotation.TaskID)
	}

	// ImageID is already the SHA256 hash, use it directly
	_, err := a.annotationRepo.Create(ctx, annotation.ImageID, annotation.User, stageIndex, annotation.Value)
	if err != nil {
		return fmt.Errorf("while creating annotation: %w", err)
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

// ClassButton represents a class button with keyboard shortcut
type ClassButton struct {
	ID   string
	Name string
	Key  string
}

func (a *AnnotatorApp) GetHTTPHandler() http.Handler {
	a.init()
	mux := http.NewServeMux()

	// Home page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		data := map[string]interface{}{
			"Title":       i("Welcome to go-annotator"),
			"ProjectName": i("Welcome to go-annotator"),
			"Description": a.Config.Meta.Description,
		}

		err := RenderPage(w, "home.html", data)
		if err != nil {
			log.Printf("error rendering home template: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Help pages
	mux.HandleFunc("/help/", func(w http.ResponseWriter, r *http.Request) {
		itemPath := pathParts(r.URL.Path)
		title := "Help"

		var tasks []TaskWithCount = nil
		var currentTask *ConfigTask = nil

		if len(itemPath) == 1 {
			// Only populate tasks for the timeline view (no markdown for tasks)
			tasks = make([]TaskWithCount, 0, len(a.Config.Tasks))

			for _, task := range a.Config.Tasks {
				availableCount, err := a.CountAvailableImages(r.Context(), task.ID)
				if err != nil {
					log.Printf("error counting available images for task %s: %s", task.ID, err)
					availableCount = 0
				}

				totalEligible, err := a.CountEligibleImages(r.Context(), task.ID)
				if err != nil {
					log.Printf("error counting eligible images for task %s: %s", task.ID, err)
					totalEligible = availableCount // fallback to available
				}

				completedCount := totalEligible - availableCount
				if completedCount < 0 {
					completedCount = 0
				}

				// Get comprehensive phase progress stats
				phaseProgress, err := a.GetPhaseProgressStats(r.Context(), task.ID)
				if err != nil {
					log.Printf("error getting phase progress for task %s: %s", task.ID, err)
					phaseProgress = &PhaseProgress{}
				}

				tasks = append(tasks, TaskWithCount{
					ConfigTask:     task,
					AvailableCount: availableCount,
					TotalCount:     totalEligible,
					CompletedCount: completedCount,
					PhaseProgress:  phaseProgress,
				})
			}
		} else if len(itemPath) == 2 {
			helpTask := itemPath[1]
			task := a.GetTask(helpTask)
			if task == nil {
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}
			currentTask = task

			// Get progress stats for this specific task
			phaseProgress, err := a.GetPhaseProgressStats(r.Context(), helpTask)
			if err != nil {
				log.Printf("error getting phase progress for task %s: %s", helpTask, err)
				phaseProgress = &PhaseProgress{}
			}

			// Get available count to check if there are images to annotate
			availableCount, err := a.CountAvailableImages(r.Context(), helpTask)
			if err != nil {
				log.Printf("error counting available images for task %s: %s", helpTask, err)
				availableCount = 0
			}

			tasks = []TaskWithCount{
				{
					ConfigTask:     task,
					AvailableCount: availableCount,
					TotalCount:     phaseProgress.Completed + phaseProgress.Pending,
					CompletedCount: phaseProgress.Completed,
					PhaseProgress:  phaseProgress,
				},
			}
		} else {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		data := map[string]interface{}{
			"Title":       title,
			"Description": a.Config.Meta.Description,
			"Task":        currentTask,
			"Tasks":       tasks,
		}

		err := RenderPage(w, "help.html", data)
		if err != nil {
			log.Printf("error rendering help template: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Annotate pages
	mux.HandleFunc("/annotate/", func(w http.ResponseWriter, r *http.Request) {
		itemPath := pathParts(r.URL.Path)

		if len(itemPath) != 3 {
			taskID := r.URL.Query().Get("task")
			step, err := a.NextAnnotationStep(r.Context(), taskID)
			if err != nil {
				log.Printf("error in annotate when getting next step from scratch: %s", err)
				w.WriteHeader(500)
				return
			}
			if step == nil {
				data := map[string]interface{}{
					"Title": i("All annotations are done!"),
				}
				err := RenderPage(w, "complete.html", data)
				if err != nil {
					log.Printf("error rendering complete template: %s", err)
				}
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
		imageFilename, _ := a.GetImageFilename(r.Context(), imageID)

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

		// Build classes with keyboard shortcuts
		classNames := make([]string, 0, len(task.Classes))
		for class := range task.Classes {
			classNames = append(classNames, class)
		}
		sort.Sort(sort.StringSlice(classNames))

		classes := []ClassButton{}
		keyIndex := 1
		for _, className := range classNames {
			classMeta := task.Classes[className]
			key := ""
			if keyIndex <= 9 {
				key = fmt.Sprintf("%d", keyIndex)
				keyIndex++
			}
			classes = append(classes, ClassButton{
				ID:   className,
				Name: i(classMeta.Name),
				Key:  key,
			})
		}

		// Get comprehensive progress information
		phaseProgress, err := a.GetPhaseProgressStats(r.Context(), taskID)
		if err != nil {
			log.Printf("error getting phase progress: %s", err)
			// Fallback to empty progress
			phaseProgress = &PhaseProgress{}
		}

		data := map[string]interface{}{
			"Title":         i("annotation"),
			"TaskID":        taskID,
			"TaskName":      task.Name,
			"ImageID":       imageID,
			"ImageFilename": imageFilename,
			"Classes":       classes,
			"PhaseProgress": phaseProgress,
			// Keep old Progress for backward compatibility
			"Progress": map[string]interface{}{
				"AvailableCount": phaseProgress.Pending,
				"TotalCount":     phaseProgress.Completed + phaseProgress.Pending,
				"CompletedCount": phaseProgress.Completed,
			},
		}

		err = RenderPage(w, "annotate.html", data)
		if err != nil {
			log.Printf("error rendering annotate template: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Asset handler - serves images by SHA256 hash
	mux.HandleFunc("/asset/", func(w http.ResponseWriter, r *http.Request) {
		itemPath := pathParts(r.URL.Path)
		if len(itemPath) != 2 {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		sha256 := itemPath[1]
		log.Printf("http: fetching asset %s", sha256)

		// Get image filename from repository
		filename, err := a.GetImageFilename(r.Context(), sha256)
		if err != nil {
			log.Printf("http: asset %s was not found: %s", sha256, err)
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}

		log.Printf("http: asset %s is %s!", sha256, filename)
		fullPath := path.Join(a.ImagesDir, filename)
		f, err := os.Open(fullPath)
		if errors.Is(err, os.ErrNotExist) {
			http.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error: http: while serving image asset: %s", err)
			return
		}
		defer f.Close()
		io.Copy(w, f)
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

	log.Printf("PrepareDatabase: running database migrations")
	// Run the migration SQL to create the new schema
	migrationSQL := `
-- Images table stores information about images to be annotated
-- Uses SHA256 hash as primary key for content-based addressing
CREATE TABLE IF NOT EXISTS images (
  sha256 TEXT PRIMARY KEY,
  filename TEXT NOT NULL,
  ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Annotations table stores all annotations
-- Uses username directly from YAML config (no FK to users table)
CREATE TABLE IF NOT EXISTS annotations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  image_sha256 TEXT NOT NULL,
  username TEXT NOT NULL,
  stage_index INTEGER NOT NULL,
  option_value TEXT NOT NULL,
  annotated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(image_sha256, username, stage_index),
  FOREIGN KEY(image_sha256) REFERENCES images(sha256) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_annotations_image_sha256 ON annotations(image_sha256);
CREATE INDEX IF NOT EXISTS idx_annotations_username ON annotations(username);
CREATE INDEX IF NOT EXISTS idx_annotations_stage ON annotations(stage_index);
	`

	_, err := a.Database.ExecContext(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("while running migrations: %w", err)
	}

	log.Printf("PrepareDatabase: ingesting images from directory")
	err = filepath.WalkDir(a.ImagesDir, func(fullPath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fullPath == a.ImagesDir {
			return nil
		}
		if info.IsDir() {
			return fmt.Errorf("while checking if item '%s' is a file: datasets must be organized in a flat folder structure. Hint: use the 'ingest' subcommand.", fullPath)
		}

		log.Printf("PrepareDatabase: ingesting image: %s", fullPath)

		// Verify it's an image
		_, err = DecodeImage(fullPath)
		if err != nil {
			return fmt.Errorf("while checking if item '%s' is an image: %w", fullPath, err)
		}

		// Hash the file to get SHA256
		fileHash, err := HashFile(fullPath)
		if err != nil {
			return fmt.Errorf("while hashing image '%s': %w", fullPath, err)
		}

		// Use repository to create image (with upsert behavior via ON CONFLICT)
		_, err = a.imageRepo.Create(ctx, fileHash, info.Name())
		if err != nil {
			// Ignore duplicate errors (hash already exists)
			if !strings.Contains(err.Error(), "UNIQUE constraint") {
				return fmt.Errorf("while inserting image '%s': %w", fullPath, err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("PrepareDatabase: success! Database is ready")
	return nil
}
