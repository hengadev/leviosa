package aggregator

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
)

// RemoveCategoryWithImages removes the category, then its images. The
// category goes first: its removal is the one refused (a category that
// still has products), and a refusal must leave the images in place.
func (s *CategoryImagesService) RemoveCategoryWithImages(ctx context.Context, categoryID string) error {
	if err := s.categoryService.RemoveCategory(ctx, categoryID); err != nil {
		return err
	}

	if err := s.imageService.DeleteImages(ctx, categoryID, string(domain.CategoryType)); err != nil {
		return err
	}
	return nil
}
