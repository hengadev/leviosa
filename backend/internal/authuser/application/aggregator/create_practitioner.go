package aggregator

import (
	"context"
	"fmt"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// CreatePractitioner is an admin adding a practitioner in one step: the
// account (active, partner role, generated password), the verified partner
// profile, and the practice room their availabilities need. When a step
// after the account fails, the account is deleted again (the profile goes
// with it), so a retry starts clean. The password is returned once.
func (s *AuthAggregatorService) CreatePractitioner(ctx context.Context, adminID uuid.UUID, request *domain.CreatePractitionerRequest) (*domain.CreatePractitionerResponse, error) {
	if err := request.Valid(ctx); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}
	if s.bookingClient == nil {
		return nil, fmt.Errorf("booking client not configured")
	}

	password, err := domain.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}

	userID, err := s.user.CreatePractitionerUser(ctx, request, password)
	if err != nil {
		return nil, err
	}

	partnerID, err := s.createPractitionerProfile(ctx, adminID, userID, request)
	if err != nil {
		s.undoPractitioner(ctx, userID)
		return nil, err
	}

	return &domain.CreatePractitionerResponse{
		UserID:    userID,
		PartnerID: partnerID,
		Email:     strings.TrimSpace(request.Email),
		Password:  password,
	}, nil
}

func (s *AuthAggregatorService) createPractitionerProfile(ctx context.Context, adminID, userID uuid.UUID, request *domain.CreatePractitionerRequest) (uuid.UUID, error) {
	partner, err := s.partner.CreatePartner(ctx, userID, strings.TrimSpace(request.Bio), "", nil, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create partner profile: %w", err)
	}

	if occupation := strings.TrimSpace(request.Occupation); occupation != "" {
		if _, err := s.partner.UpdatePartner(ctx, partner.ID, &domain.UpdatePartnerRequest{Occupation: &occupation}); err != nil {
			return uuid.Nil, fmt.Errorf("set partner occupation: %w", err)
		}
	}

	if _, err := s.partner.VerifyPartner(ctx, partner.ID, adminID); err != nil {
		return uuid.Nil, fmt.Errorf("verify partner: %w", err)
	}

	if err := s.bookingClient.EnsurePracticeRoom(ctx, userID); err != nil {
		return uuid.Nil, fmt.Errorf("give the practitioner the practice room: %w", err)
	}

	return partner.ID, nil
}

// undoPractitioner deletes a half-created practitioner. A failure here is
// only logged: the caller already returns the error that caused it.
func (s *AuthAggregatorService) undoPractitioner(ctx context.Context, userID uuid.UUID) {
	if err := s.user.DeleteUser(ctx, userID); err != nil {
		if logger, logErr := ctxutil.GetLoggerFromContext(ctx); logErr == nil {
			logger.ErrorContext(ctx, "create practitioner: could not delete the half-created account",
				"user_id", userID, "error", err)
		}
	}
}
