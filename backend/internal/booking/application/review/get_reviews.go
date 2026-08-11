package review

import (
	"context"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/google/uuid"
)

// defaultReviewPageSize is used when the caller does not specify a positive limit.
const defaultReviewPageSize = 20

// GetPublicReviewsForPartner returns non-hidden reviews and an aggregate rating for
// a partner. The aggregate is computed over the full non-hidden set, independent of
// pagination.
func (s *service) GetPublicReviewsForPartner(ctx context.Context, partnerID uuid.UUID, limit, offset int) (*domain.PublicReviewListResponse, error) {
	if limit <= 0 {
		limit = defaultReviewPageSize
	}

	reviews, err := s.reviewRepo.ListByPartnerID(ctx, partnerID, ports.ReviewFilter{
		IncludeHidden: false,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list partner reviews: %w", err)
	}

	avg, count, err := s.reviewRepo.AggregateByPartnerID(ctx, partnerID)
	if err != nil {
		return nil, fmt.Errorf("aggregate partner reviews: %w", err)
	}

	return buildPublicReviewListResponse(reviews, avg, count), nil
}

// GetPublicReviewsForProduct returns non-hidden reviews and an aggregate rating for
// a product. The aggregate is computed over the full non-hidden set, independent of
// pagination.
func (s *service) GetPublicReviewsForProduct(ctx context.Context, productID uuid.UUID, limit, offset int) (*domain.PublicReviewListResponse, error) {
	if limit <= 0 {
		limit = defaultReviewPageSize
	}

	reviews, err := s.reviewRepo.ListByProductID(ctx, productID, ports.ReviewFilter{
		IncludeHidden: false,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list product reviews: %w", err)
	}

	avg, count, err := s.reviewRepo.AggregateByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("aggregate product reviews: %w", err)
	}

	return buildPublicReviewListResponse(reviews, avg, count), nil
}

func buildPublicReviewListResponse(reviews []*domain.Review, avg float64, count int) *domain.PublicReviewListResponse {
	responses := make([]domain.ReviewResponse, len(reviews))
	for i, r := range reviews {
		responses[i] = domain.ReviewResponseFromReview(r)
	}
	return &domain.PublicReviewListResponse{
		AggregateRating: avg,
		ReviewCount:     count,
		Reviews:         responses,
	}
}
