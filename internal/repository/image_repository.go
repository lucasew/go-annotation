package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lucasew/go-annotation/internal/domain"
	"github.com/lucasew/go-annotation/internal/sqlc"
)

// ImageRepository implements domain.ImageRepository using SQLC
type ImageRepository struct {
	queries *sqlc.Queries
}

// NewImageRepository creates a new ImageRepository
func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{
		queries: sqlc.New(db),
	}
}

// NewImageRepositoryWithTx creates a new ImageRepository with a transaction
func NewImageRepositoryWithTx(tx *sql.Tx) *ImageRepository {
	return &ImageRepository{
		queries: sqlc.New(tx),
	}
}

// Create creates a new image record
func (r *ImageRepository) Create(ctx context.Context, path, originalFilename string) (*domain.Image, error) {
	params := sqlc.CreateImageParams{
		Path: path,
		OriginalFilename: sql.NullString{
			String: originalFilename,
			Valid:  originalFilename != "",
		},
	}

	img, err := r.queries.CreateImage(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainImage(img), nil
}

// GetByID retrieves an image by its ID
func (r *ImageRepository) GetByID(ctx context.Context, id int64) (*domain.Image, error) {
	img, err := r.queries.GetImage(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return toDomainImage(img), nil
}

// GetByPath retrieves an image by its path
func (r *ImageRepository) GetByPath(ctx context.Context, path string) (*domain.Image, error) {
	img, err := r.queries.GetImageByPath(ctx, path)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return toDomainImage(img), nil
}

// List retrieves all images
func (r *ImageRepository) List(ctx context.Context) ([]*domain.Image, error) {
	images, err := r.queries.ListImages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Image, len(images))
	for i, img := range images {
		result[i] = toDomainImage(img)
	}

	return result, nil
}

// ListNotFinished retrieves images that haven't been fully annotated
func (r *ImageRepository) ListNotFinished(ctx context.Context, limit int) ([]*domain.Image, error) {
	images, err := r.queries.ListImagesNotFinished(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Image, len(images))
	for i, img := range images {
		result[i] = toDomainImage(img)
	}

	return result, nil
}

// UpdateCompletionStatus updates the completion status of an image
func (r *ImageRepository) UpdateCompletionStatus(ctx context.Context, id int64, completedStages int, isFinished bool) error {
	params := sqlc.UpdateImageCompletionStatusParams{
		ID:              id,
		CompletedStages: int64(completedStages),
		IsFinished:      isFinished,
	}

	return r.queries.UpdateImageCompletionStatus(ctx, params)
}

// Count returns the total number of images
func (r *ImageRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountImages(ctx)
}

// CountPending returns the number of images not yet finished
func (r *ImageRepository) CountPending(ctx context.Context) (int64, error) {
	return r.queries.CountPendingImages(ctx)
}

// Delete removes an image by ID
func (r *ImageRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteImage(ctx, id)
}

// toDomainImage converts a sqlc.Image to domain.Image
func toDomainImage(img sqlc.Image) *domain.Image {
	return &domain.Image{
		ID:               img.ID,
		Path:             img.Path,
		OriginalFilename: img.OriginalFilename.String,
		IngestedAt:       img.IngestedAt,
		CompletedStages:  int(img.CompletedStages),
		IsFinished:       img.IsFinished,
	}
}

// Verify that ImageRepository implements domain.ImageRepository
var _ domain.ImageRepository = (*ImageRepository)(nil)
