package domain

import (
	"context"
	"time"
)

// Image represents an image to be annotated
type Image struct {
	ID               int64
	Path             string
	OriginalFilename string
	IngestedAt       time.Time
	CompletedStages  int
	IsFinished       bool
}

// ImageRepository defines the interface for image storage operations
type ImageRepository interface {
	// Create creates a new image record
	Create(ctx context.Context, path, originalFilename string) (*Image, error)

	// GetByID retrieves an image by its ID
	GetByID(ctx context.Context, id int64) (*Image, error)

	// GetByPath retrieves an image by its path
	GetByPath(ctx context.Context, path string) (*Image, error)

	// List retrieves all images
	List(ctx context.Context) ([]*Image, error)

	// ListNotFinished retrieves images that haven't been fully annotated
	ListNotFinished(ctx context.Context, limit int) ([]*Image, error)

	// UpdateCompletionStatus updates the completion status of an image
	UpdateCompletionStatus(ctx context.Context, id int64, completedStages int, isFinished bool) error

	// Count returns the total number of images
	Count(ctx context.Context) (int64, error)

	// CountPending returns the number of images not yet finished
	CountPending(ctx context.Context) (int64, error)

	// Delete removes an image by ID
	Delete(ctx context.Context, id int64) error
}
