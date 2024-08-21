package user

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (s *Service) UpdateAccount(ctx context.Context, userCandidate *User, userID string) error {
	if pbms := userCandidate.Valid(ctx); len(pbms) > 0 {
		return serverutil.FormatError(pbms, "user")
	}
	err := s.repo.ModifyAccount(
		ctx,
		userCandidate,
		map[string]any{"id": userID},
		"Email",
		"Password",
	)
	if err != nil {
		return fmt.Errorf("add account: %w", err)
	}
	return nil
}
