package sessionService

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (s *Service) CreateSession(ctx context.Context, userID string, role models.Role) (string, error) {
	session, err := NewSession(userID, role)
	if err != nil {
		return "", domain.NewInvalidValueErr("invalid user ID")
	}
	if role == models.UNKNOWN {
		return "", domain.NewInvalidValueErr("invalid role: role must be different than 'UNKNOWN'")
	}
	sessionEncoded, err := json.Marshal(session)
	if err != nil {
		return "", domain.NewJSONMarshalErr(err)
	}
	err = s.Repo.CreateSession(ctx, session.ID, sessionEncoded)
	if err != nil {
		switch {
		case errors.Is(err, rp.ErrDatabase):
			return "", domain.NewQueryFailedErr(err)
		case errors.Is(err, rp.ErrContext):
			return "", err
		default:
			return "", domain.NewUnexpectTypeErr(err)
		}
	}
	return session.ID, nil
}
