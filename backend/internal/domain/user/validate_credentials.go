package userService

import (
	"context"
	"errors"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) ValidateCredentials(ctx context.Context, creds *Credentials) error {
	hashedPassword, err := s.repo.GetHashedPasswordByEmail(ctx, creds.Email)
	switch {
	case errors.Is(err, rp.ErrNotFound):
		return domain.NewNotFoundErr(err)
	case errors.Is(err, rp.ErrDatabase):
		return err
	case err != nil:
		return domain.NewUnexpectTypeErr(err)
	}
	if err = CompareHashAndPassword(hashedPassword, creds.Password); err != nil {
		return fmt.Errorf("password comparison failed: provided password does not match the stored hash: %w", err)
	}
	return nil
}

// ValidatePassword is a helper function that implements the logic behind verifying if the hashed password corresponds to thee password value sent
func CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
