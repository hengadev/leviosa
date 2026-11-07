package aggregator

import (
	"context"
	"errors"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/ports"
)

type fakeCategoryRemover struct {
	ports.CategoryService
	err error
}

func (f *fakeCategoryRemover) RemoveCategory(context.Context, string) error { return f.err }

type fakeImageDeleter struct {
	ports.ImageParentService
	deleted bool
}

func (f *fakeImageDeleter) DeleteImages(context.Context, string, string) error {
	f.deleted = true
	return nil
}

func TestRemoveCategoryWithImages(t *testing.T) {
	t.Run("a refused removal keeps the images", func(t *testing.T) {
		refused := errors.New("category has products")
		images := &fakeImageDeleter{}
		svc := NewCategoryAggregatorService(&fakeCategoryRemover{err: refused}, images)

		if err := svc.RemoveCategoryWithImages(context.Background(), "id"); !errors.Is(err, refused) {
			t.Fatalf("err = %v, want %v", err, refused)
		}
		if images.deleted {
			t.Fatal("images deleted although the category was not removed")
		}
	})

	t.Run("a removed category loses its images", func(t *testing.T) {
		images := &fakeImageDeleter{}
		svc := NewCategoryAggregatorService(&fakeCategoryRemover{}, images)

		if err := svc.RemoveCategoryWithImages(context.Background(), "id"); err != nil {
			t.Fatalf("err = %v", err)
		}
		if !images.deleted {
			t.Fatal("images kept after the category was removed")
		}
	})
}
