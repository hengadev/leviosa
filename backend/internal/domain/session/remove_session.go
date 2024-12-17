package sessionService

import (
	"context"
	"errors"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) RemoveSession(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return domain.NewNotFoundErr(errors.New("empty session ID"))
	}
	err := s.Repo.RemoveSession(ctx, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrDatabase):
			return domain.NewQueryFailedErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return err
		default:
			return domain.NewUnexpectTypeErr(err)
		}
	}
	return nil
}
