package aggregator

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// BecomePractitioner gives an administrator's own account the practitioner
// side, for a practice whose owner both runs the admin and sees the
// clients: a verified profile and the practice room, while the account
// keeps the administrator role. Each step already done is skipped, so a
// retry after a failure finishes the job instead of failing on the profile.
func (s *AuthAggregatorService) BecomePractitioner(ctx context.Context, adminID uuid.UUID, request *domain.BecomePractitionerRequest) (uuid.UUID, error) {
	if err := request.Valid(ctx); err != nil {
		return uuid.Nil, errs.NewInvalidValueErr(err.Error())
	}
	if s.bookingClient == nil {
		return uuid.Nil, fmt.Errorf("booking client not configured")
	}

	existing, err := s.partner.GetPartnerByUserID(ctx, adminID)
	switch {
	case errors.Is(err, errs.ErrRepositoryNotFound):
		return s.createPractitionerProfile(ctx, adminID, adminID, request.Bio, request.Occupation)
	case err != nil:
		return uuid.Nil, fmt.Errorf("look up the practitioner profile: %w", err)
	}

	if existing.StripeAccountStatus != domain.StripeAccountStatusActive || !existing.StripeOnboardingComplete {
		if _, err := s.partner.VerifyPartner(ctx, existing.ID, adminID); err != nil {
			return uuid.Nil, fmt.Errorf("verify partner: %w", err)
		}
	}
	if err := s.bookingClient.EnsurePracticeRoom(ctx, adminID); err != nil {
		return uuid.Nil, fmt.Errorf("give the practitioner the practice room: %w", err)
	}
	return existing.ID, nil
}
