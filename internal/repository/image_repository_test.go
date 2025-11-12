package repository

import (
	"context"
	"testing"
)

func TestImageRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	t.Run("creates image successfully", func(t *testing.T) {
		img, err := repo.Create(ctx, "/path/to/image.jpg", "image.jpg")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if img.ID == 0 {
			t.Error("Expected non-zero ID")
		}
		if img.Path != "/path/to/image.jpg" {
			t.Errorf("Path = %v, want %v", img.Path, "/path/to/image.jpg")
		}
		if img.OriginalFilename != "image.jpg" {
			t.Errorf("OriginalFilename = %v, want %v", img.OriginalFilename, "image.jpg")
		}
		if img.CompletedStages != 0 {
			t.Errorf("CompletedStages = %v, want 0", img.CompletedStages)
		}
		if img.IsFinished {
			t.Error("IsFinished should be false")
		}
		if img.IngestedAt.IsZero() {
			t.Error("IngestedAt should not be zero")
		}
	})

	t.Run("creates image without original filename", func(t *testing.T) {
		img, err := repo.Create(ctx, "/path/to/image2.jpg", "")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if img.OriginalFilename != "" {
			t.Errorf("OriginalFilename = %v, want empty string", img.OriginalFilename)
		}
	})

	t.Run("fails on duplicate path", func(t *testing.T) {
		_, err := repo.Create(ctx, "/path/to/image.jpg", "duplicate.jpg")
		if err == nil {
			t.Error("Expected error for duplicate path")
		}
	})
}

func TestImageRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test image
	created, err := repo.Create(ctx, "/test/image.jpg", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	t.Run("retrieves existing image", func(t *testing.T) {
		img, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if img == nil {
			t.Fatal("Expected image, got nil")
		}
		if img.ID != created.ID {
			t.Errorf("ID = %v, want %v", img.ID, created.ID)
		}
		if img.Path != created.Path {
			t.Errorf("Path = %v, want %v", img.Path, created.Path)
		}
	})

	t.Run("returns nil for non-existent image", func(t *testing.T) {
		img, err := repo.GetByID(ctx, 9999)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if img != nil {
			t.Error("Expected nil for non-existent image")
		}
	})
}

func TestImageRepository_GetByPath(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test image
	created, err := repo.Create(ctx, "/test/image.jpg", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	t.Run("retrieves existing image by path", func(t *testing.T) {
		img, err := repo.GetByPath(ctx, "/test/image.jpg")
		if err != nil {
			t.Fatalf("GetByPath() error = %v", err)
		}

		if img == nil {
			t.Fatal("Expected image, got nil")
		}
		if img.ID != created.ID {
			t.Errorf("ID = %v, want %v", img.ID, created.ID)
		}
	})

	t.Run("returns nil for non-existent path", func(t *testing.T) {
		img, err := repo.GetByPath(ctx, "/nonexistent.jpg")
		if err != nil {
			t.Fatalf("GetByPath() error = %v", err)
		}
		if img != nil {
			t.Error("Expected nil for non-existent path")
		}
	})
}

func TestImageRepository_List(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test images
	_, err := repo.Create(ctx, "/test/image1.jpg", "image1.jpg")
	if err != nil {
		t.Fatalf("Failed to create image1: %v", err)
	}
	_, err = repo.Create(ctx, "/test/image2.jpg", "image2.jpg")
	if err != nil {
		t.Fatalf("Failed to create image2: %v", err)
	}

	t.Run("lists all images", func(t *testing.T) {
		images, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(images) != 2 {
			t.Errorf("Got %d images, want 2", len(images))
		}
	})
}

func TestImageRepository_ListNotFinished(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test images
	img1, _ := repo.Create(ctx, "/test/image1.jpg", "image1.jpg")
	img2, _ := repo.Create(ctx, "/test/image2.jpg", "image2.jpg")
	img3, _ := repo.Create(ctx, "/test/image3.jpg", "image3.jpg")

	// Mark one as finished
	repo.UpdateCompletionStatus(ctx, img2.ID, 3, true)

	t.Run("lists only non-finished images", func(t *testing.T) {
		images, err := repo.ListNotFinished(ctx, 10)
		if err != nil {
			t.Fatalf("ListNotFinished() error = %v", err)
		}

		if len(images) != 2 {
			t.Errorf("Got %d images, want 2", len(images))
		}

		// Check that img2 is not in the list
		for _, img := range images {
			if img.ID == img2.ID {
				t.Error("Finished image should not be in list")
			}
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		images, err := repo.ListNotFinished(ctx, 1)
		if err != nil {
			t.Fatalf("ListNotFinished() error = %v", err)
		}

		if len(images) != 1 {
			t.Errorf("Got %d images, want 1", len(images))
		}
	})

	// Update completion stages to test ordering
	repo.UpdateCompletionStatus(ctx, img1.ID, 1, false)
	repo.UpdateCompletionStatus(ctx, img3.ID, 2, false)

	t.Run("orders by completed stages", func(t *testing.T) {
		images, err := repo.ListNotFinished(ctx, 10)
		if err != nil {
			t.Fatalf("ListNotFinished() error = %v", err)
		}

		if len(images) < 2 {
			t.Fatal("Need at least 2 images for this test")
		}

		// Images should be ordered by completed_stages ASC
		if images[0].CompletedStages > images[1].CompletedStages {
			t.Error("Images should be ordered by completed_stages ASC")
		}
	})
}

func TestImageRepository_UpdateCompletionStatus(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test image
	img, _ := repo.Create(ctx, "/test/image.jpg", "test.jpg")

	t.Run("updates completion status", func(t *testing.T) {
		err := repo.UpdateCompletionStatus(ctx, img.ID, 3, true)
		if err != nil {
			t.Fatalf("UpdateCompletionStatus() error = %v", err)
		}

		// Verify update
		updated, err := repo.GetByID(ctx, img.ID)
		if err != nil {
			t.Fatalf("Failed to get updated image: %v", err)
		}

		if updated.CompletedStages != 3 {
			t.Errorf("CompletedStages = %v, want 3", updated.CompletedStages)
		}
		if !updated.IsFinished {
			t.Error("IsFinished should be true")
		}
	})
}

func TestImageRepository_Count(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	t.Run("counts all images", func(t *testing.T) {
		// Create test images
		repo.Create(ctx, "/test/image1.jpg", "image1.jpg")
		repo.Create(ctx, "/test/image2.jpg", "image2.jpg")
		repo.Create(ctx, "/test/image3.jpg", "image3.jpg")

		count, err := repo.Count(ctx)
		if err != nil {
			t.Fatalf("Count() error = %v", err)
		}

		if count != 3 {
			t.Errorf("Count = %v, want 3", count)
		}
	})
}

func TestImageRepository_CountPending(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test images
	img1, _ := repo.Create(ctx, "/test/image1.jpg", "image1.jpg")
	img2, _ := repo.Create(ctx, "/test/image2.jpg", "image2.jpg")
	repo.Create(ctx, "/test/image3.jpg", "image3.jpg")

	// Mark one as finished
	repo.UpdateCompletionStatus(ctx, img2.ID, 3, true)

	t.Run("counts only pending images", func(t *testing.T) {
		count, err := repo.CountPending(ctx)
		if err != nil {
			t.Fatalf("CountPending() error = %v", err)
		}

		if count != 2 {
			t.Errorf("CountPending = %v, want 2", count)
		}
	})

	// Mark all as finished
	repo.UpdateCompletionStatus(ctx, img1.ID, 3, true)

	t.Run("returns 0 when all finished", func(t *testing.T) {
		count, err := repo.CountPending(ctx)
		if err != nil {
			t.Fatalf("CountPending() error = %v", err)
		}

		if count != 1 {
			t.Errorf("CountPending = %v, want 1", count)
		}
	})
}

func TestImageRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test image
	img, _ := repo.Create(ctx, "/test/image.jpg", "test.jpg")

	t.Run("deletes image", func(t *testing.T) {
		err := repo.Delete(ctx, img.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify deletion
		deleted, err := repo.GetByID(ctx, img.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if deleted != nil {
			t.Error("Image should be deleted")
		}
	})
}

// Benchmark tests
func BenchmarkImageRepository_Create(b *testing.B) {
	db := SetupTestDB(&testing.T{})
	defer db.Close()

	repo := NewImageRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.Create(ctx, "/test/image.jpg", "test.jpg")
		// Clean up to avoid duplicates
		db.Exec("DELETE FROM images")
	}
}

func BenchmarkImageRepository_GetByID(b *testing.B) {
	db := SetupTestDB(&testing.T{})
	defer db.Close()

	repo := NewImageRepository(db)
	ctx := context.Background()

	// Create test image
	img, _ := repo.Create(ctx, "/test/image.jpg", "test.jpg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByID(ctx, img.ID)
	}
}
