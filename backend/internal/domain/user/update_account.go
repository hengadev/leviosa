package user

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (s *Service) UpdateAccount(ctx context.Context, userCandidate *User) error {
	if pbms := userCandidate.Valid(ctx); len(pbms) > 0 {
		return serverutil.FormatError(pbms, "user")
	}
	// TODO: use a different function if the email needs to be modified.
	if err := s.repo.ModifyAccount(ctx, userCandidate); err != nil {
		return fmt.Errorf("add account: %w", err)
	}
	return nil
}
