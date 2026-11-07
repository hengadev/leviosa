package domain

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/common/validation"
	"github.com/google/uuid"
	"github.com/hengadev/errsx"
)

// CreatePractitionerRequest is an admin creating a practitioner directly: the
// account (partner role, active) and its public profile, without the
// self-service sign-up. The password is generated and returned once.
type CreatePractitionerRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Telephone  string `json:"telephone"`
	Occupation string `json:"occupation,omitempty"`
	Bio        string `json:"bio,omitempty"`
}

func (r *CreatePractitionerRequest) Valid(ctx context.Context) error {
	var errs errsx.Map

	if err := validateName(r.FirstName, "first_name"); err != nil {
		errs.Set("first_name", err)
	}
	if err := validateName(r.LastName, "last_name"); err != nil {
		errs.Set("last_name", err)
	}
	if err := validation.ValidateEmail(r.Email); err != nil {
		errs.Set("email", err)
	}
	if err := validation.ValidatePhone(r.Telephone); err != nil {
		errs.Set("telephone", err)
	}
	if len(r.Occupation) > 200 {
		errs.Set("occupation", "occupation must be 200 characters or less")
	}
	if len(r.Bio) > 1000 {
		errs.Set("bio", "bio must be 1000 characters or less")
	}

	return errs.AsError()
}

// BecomePractitionerRequest is an administrator who is also the practice's
// practitioner giving their own account the practitioner profile.
type BecomePractitionerRequest struct {
	Occupation string `json:"occupation,omitempty"`
	Bio        string `json:"bio,omitempty"`
}

func (r *BecomePractitionerRequest) Valid(ctx context.Context) error {
	var errs errsx.Map

	if len(r.Occupation) > 200 {
		errs.Set("occupation", "occupation must be 200 characters or less")
	}
	if len(r.Bio) > 1000 {
		errs.Set("bio", "bio must be 1000 characters or less")
	}

	return errs.AsError()
}

// CreatePractitionerResponse carries the generated password: it is shown to
// the admin once and stored nowhere else in clear.
type CreatePractitionerResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

// GeneratedPasswordLength is the length of the passwords generated for
// accounts an admin creates.
const GeneratedPasswordLength = 20

// Letters and digits without the look-alikes (0/O, 1/l/I), since the admin
// passes the password on by hand.
const generatedPasswordAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GeneratePassword returns a random password from crypto/rand.
func GeneratePassword() (string, error) {
	var b strings.Builder
	max := big.NewInt(int64(len(generatedPasswordAlphabet)))
	for range GeneratedPasswordLength {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(generatedPasswordAlphabet[n.Int64()])
	}
	return b.String(), nil
}
