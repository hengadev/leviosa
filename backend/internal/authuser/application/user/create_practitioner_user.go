package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
)

// CreatePractitionerUser creates the account of a practitioner an admin adds:
// active, with the partner role and the given password. It refuses an email
// that already has an account.
func (s *UserService) CreatePractitionerUser(ctx context.Context, request *domain.CreatePractitionerRequest, password string) (uuid.UUID, error) {
	if err := request.Valid(ctx); err != nil {
		return uuid.Nil, errs.NewInvalidValueErr(err.Error())
	}

	email := strings.TrimSpace(request.Email)
	emailBytes, err := encx.SerializeValue(email)
	if err != nil {
		return uuid.Nil, errs.NewInvalidValueErr(fmt.Sprintf("failed to serialize email: %v", err))
	}
	exists, err := s.repo.ExistsByEmailHash(ctx, s.crypto.HashBasic(ctx, emailBytes))
	if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
		return uuid.Nil, fmt.Errorf("check email availability: %w", err)
	}
	if exists {
		return uuid.Nil, errs.NewConflictErr(errors.New("an account already exists with this email"))
	}

	user := &domain.User{
		ID:        uuid.New(),
		State:     domain.Active,
		Email:     email,
		Password:  password,
		FirstName: strings.TrimSpace(request.FirstName),
		LastName:  strings.TrimSpace(request.LastName),
		Role:      identity.PartnerStr,
		Telephone: strings.TrimSpace(request.Telephone),
		CreatedAt: time.Now(),
	}

	userEncx, err := domain.ProcessUserEncx(ctx, s.crypto, user)
	if err != nil {
		return uuid.Nil, errs.NewNotEncryptedErr("practitioner user", err)
	}
	if err := s.repo.CreateUser(ctx, userEncx); err != nil {
		if errors.Is(err, errs.ErrUniqueViolation) {
			return uuid.Nil, errs.NewConflictErr(errors.New("an account already exists with this email"))
		}
		return uuid.Nil, fmt.Errorf("create practitioner user: %w", err)
	}

	return user.ID, nil
}
