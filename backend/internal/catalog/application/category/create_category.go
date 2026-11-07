package category

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

func (s *CategoryService) CreateCategory(ctx context.Context, request *domain.CreateCategoryRequest) (string, error) {
	// The name is kept as typed; uniqueness ignores case (catalog_categories_name_lower_key).
	request.Name = strings.TrimSpace(request.Name)
	if err := request.Valid(ctx); err != nil {
		return "", errs.NewInvalidValueErr(err.Error())
	}

	now := time.Now()

	category := &domain.Category{
		ID:          uuid.New(),
		Name:        request.Name,
		Description: request.Description,
		Status:      domain.Draft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	categoryID, err := s.repo.AddCategory(ctx, category)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrUniqueViolation): // This is the core change: handling the DB unique constraint directly
			return "", errs.NewAlreadyExistsError(err, "category with this name")
		default:
			return "", fmt.Errorf("create category: %w", err)
		}
	}
	return categoryID, nil
}
