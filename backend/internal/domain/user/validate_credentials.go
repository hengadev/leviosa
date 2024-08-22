package userService

import (
	"context"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) ValidateCredentials(ctx context.Context, creds *Credentials) (int, Role, error) {
	fail := func(err error) (int, Role, error) {
		return 0, UNKNOWN, app.NewAuthErr(err)
	}
	// get credential from database
	userID, password, role, err := s.repo.GetCredentials(ctx, creds)
	if err != nil {
		fail(err)
	}
	// check if userID and role are valid
	if userID == 0 || role == UNKNOWN {
		fail(fmt.Errorf("invalid userID and role from database"))
	}
	// check if same password
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(creds.Password)); err != nil {
		return fail(err)
	}
	return userID, role, nil
}
