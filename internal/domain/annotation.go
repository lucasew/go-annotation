package domain

import (
	"context"
	"time"
)

// Annotation represents a single annotation of an image by a user
type Annotation struct {
	ID          int64
	ImageID     int64
	Username    string
	StageIndex  int
	OptionValue string
	AnnotatedAt time.Time
}

// AnnotationWithImage extends Annotation with image information
type AnnotationWithImage struct {
	Annotation
	ImagePath            string
	ImageOriginalFilename string
}

// AnnotationStats provides statistics about annotations
type AnnotationStats struct {
	AnnotatedImages  int64
	TotalAnnotations int64
	TotalUsers       int64
}

// AnnotationRepository defines the interface for annotation storage operations
type AnnotationRepository interface {
	// Create creates or updates an annotation (upsert)
	Create(ctx context.Context, imageID int64, username string, stageIndex int, optionValue string) (*Annotation, error)

	// Get retrieves a specific annotation
	Get(ctx context.Context, imageID int64, username string, stageIndex int) (*Annotation, error)

	// GetForImage retrieves all annotations for a specific image
	GetForImage(ctx context.Context, imageID int64) ([]*Annotation, error)

	// GetByUser retrieves annotations by a specific user (paginated)
	GetByUser(ctx context.Context, username string, limit, offset int) ([]*AnnotationWithImage, error)

	// GetByImageAndUser retrieves all annotations for an image by a specific user
	GetByImageAndUser(ctx context.Context, imageID int64, username string) ([]*Annotation, error)

	// CountByUser returns the total number of annotations by a user
	CountByUser(ctx context.Context, username string) (int64, error)

	// ListPendingImagesForUserAndStage finds images that need annotation by a user for a specific stage
	ListPendingImagesForUserAndStage(ctx context.Context, username string, stageIndex int, limit int) ([]*Image, error)

	// Exists checks if an annotation exists
	Exists(ctx context.Context, imageID int64, username string, stageIndex int) (bool, error)

	// Delete removes an annotation by ID
	Delete(ctx context.Context, id int64) error

	// DeleteForImage removes all annotations for an image
	DeleteForImage(ctx context.Context, imageID int64) error

	// GetStats returns overall annotation statistics
	GetStats(ctx context.Context) (*AnnotationStats, error)
}
