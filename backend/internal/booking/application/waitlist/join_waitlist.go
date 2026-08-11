package waitlist

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// JoinWaitlist creates a new waitlist entry in the `waiting` state for a
// client (or guest) who wants to be considered for a product/date-range
// combination once a matching slot frees up.
//
// This only creates the entry — it does not search availability, notify
// anyone, or generate a claim link. Matching freed availability against
// waiting entries and notifying them is issue 015 (waitlist-notify-and-claim),
// which is blocked and out of scope here.
//
// Flow:
//  1. Validate the product exists in the catalog
//  2. Build and validate the domain entity (guest/client identity, date range)
//  3. Encrypt and persist
func (s *WaitlistService) JoinWaitlist(ctx context.Context, request *domain.JoinWaitlistRequest) (*domain.WaitlistEntryResponse, error) {
	if request == nil {
		return nil, errs.NewInvalidValueErr("request body is required")
	}

	// 1. Verify the product exists
	if _, err := s.productService.GetProductByID(ctx, request.ProductID.String()); err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("product not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	// 2. Build the domain entity (validates identity + date range invariants).
	// Domain validation errors are wrapped with errs.ErrInvalidValue so the
	// handler's httpx.RespondWithServiceError classifies them as 400 Bad
	// Request, while still supporting errors.Is against the specific domain
	// sentinel (e.g. domain.ErrAmbiguousWaitlistIdentity) for callers that care.
	entry, err := domain.NewWaitlistEntry(
		request.ProductID,
		request.ClientID,
		request.DesiredRangeStart,
		request.DesiredRangeEnd,
		request.Notes,
		request.GuestFirstName,
		request.GuestLastName,
		request.GuestEmail,
		request.GuestPhone,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrInvalidValue, err)
	}

	// 3. Encrypt and persist
	entryEncx, err := domain.ProcessWaitlistEntryEncx(ctx, s.crypto, entry)
	if err != nil {
		return nil, fmt.Errorf("encrypt waitlist entry: %w", err)
	}

	if err := s.waitlistRepo.Create(ctx, entryEncx); err != nil {
		return nil, fmt.Errorf("create waitlist entry: %w", err)
	}

	return &domain.WaitlistEntryResponse{
		ID:                entry.ID,
		ProductID:         entry.ProductID,
		ClientID:          entry.ClientID,
		DesiredRangeStart: entry.DesiredRangeStart,
		DesiredRangeEnd:   entry.DesiredRangeEnd,
		Notes:             entry.Notes,
		Status:            entry.Status,
		CreatedAt:         entry.CreatedAt,
		GuestFirstName:    entry.GuestFirstName,
		GuestLastName:     entry.GuestLastName,
		GuestEmail:        entry.GuestEmail,
		GuestPhone:        entry.GuestPhone,
	}, nil
}
