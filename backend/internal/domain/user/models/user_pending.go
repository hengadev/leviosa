package models

import (
	"context"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// the user that is send to admin for validation
type UserPending struct {
	EmailHash string `json:"email_hash"`
	LastName  string `json:"lastname"`
	FirstName string `json:"firstname"`
	GoogleID  string `json:"google_id"`
	AppleID   string `json:"apple_id"`
}

// the admin receive this when validating the user
type UserPendingResponse struct {
	EmailHash string `json:"email"`
	Role      string `json:"role"`
}

func NewPrependingUser(user *UserSignUp) *User {
	return &User{
		Email:      user.Email,
		Password:   user.Password,
		BirthDate:  user.BirthDate,
		LastName:   user.LastName,
		FirstName:  user.FirstName,
		Gender:     user.Gender,
		Telephone:  user.Telephone,
		PostalCode: user.PostalCode,
		City:       user.City,
		Address1:   user.Address1,
		Address2:   user.Address2,
	}
}

func (u UserPending) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}

func (u UserPendingResponse) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}
