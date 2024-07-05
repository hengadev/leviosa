package user

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

func (s *Service) UpdateAccount(ctx context.Context, userCandidate *User) error {
	if pbms := userCandidate.Valid(ctx); len(pbms) > 0 {
		return app.NewInvalidUserToUpdateErr(pbms)
	}
	// TODO: use a different function if the email needs to be modified.
	if err := s.repo.ModifyAccount(ctx, userCandidate); err != nil {
		return fmt.Errorf("add account: %w", err)
	}
	return nil
}
