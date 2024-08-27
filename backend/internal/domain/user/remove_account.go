package userService

import (
	"context"
	"fmt"
)

func (s *Service) DeleteUser(ctx context.Context, id int) error {
	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("remove account: %w", err)
	}
	return nil
}
