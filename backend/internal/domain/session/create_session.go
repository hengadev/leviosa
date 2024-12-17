package sessionService

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Service) CreateSession(ctx context.Context, userID string, role userService.Role) (string, error) {
	session := NewSession(userID, role)
	sessionEncoded, err := json.Marshal(session)
	if err != nil {
		return "", domain.NewJSONMarshalErr(err)
	}
	err = s.Repo.CreateSession(ctx, session.ID, sessionEncoded)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrDatabase):
			return "", domain.NewQueryFailedErr(err)
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return "", err
		default:
			return "", domain.NewUnexpectTypeErr(err)
		}
	}
	return session.ID, nil
}
