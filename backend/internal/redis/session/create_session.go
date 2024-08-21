package sessionRepository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
)

func (s *Repository) CreateSession(ctx context.Context, sess *session.Session) error {
	if sess == nil {
		return fmt.Errorf("nil session")
	}
	if sess.IsZero() {
		return fmt.Errorf("zero session")
	}
	sessionEncoded, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	err = s.Client.Set(ctx, sess.ID, sessionEncoded, session.SessionDuration).Err()
	if err != nil {
		return rp.NewRessourceCreationErr(err)
	}
	return nil
}
