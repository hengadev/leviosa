package reviewRepository

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

func (r *Repository) ListByProductID(ctx context.Context, productID uuid.UUID, filter ports.ReviewFilter) ([]*domain.Review, error) {
	return r.listByColumn(ctx, "product_id", productID, filter)
}
