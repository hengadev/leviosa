package models

import (
	"context"
	"time"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
	"github.com/google/uuid"
)

// TODO: add the picture in this
type User struct {
	ID             string    `json:"-"`
	Email          string    `json:"-"`
	EmailHash      string    `json:"-"`
	EncryptedEmail string    `json:"-"`
	Password       string    `json:"-"`
	PasswordHash   string    `json:"-"`
	Picture        string    `json:"-"`
	CreatedAt      time.Time `json:"-"`
	LoggedInAt     time.Time `json:"-"`
	Role           string    `json:"-"`
	BirthDate      string    `json:"-"`
	LastName       string    `json:"-"`
	FirstName      string    `json:"-"`
	Gender         string    `json:"-"`
	Telephone      string    `json:"-"`
	PostalCode     string    `json:"-"`
	City           string    `json:"-"`
	Address1       string    `json:"-"`
	Address2       string    `json:"-"`
	GoogleID       string    `json:"-"`
	AppleID        string    `json:"-"`
}

func (a *User) Create() {
	a.CreatedAt = time.Now().UTC()
}

func (a *User) Login() {
	a.LoggedInAt = time.Now().UTC()
}

// I need to use the hash things in that function
func NewUser(
	user UserSignUp,
	role Role,
) *User {
	return &User{
		ID:         uuid.NewString(),
		Email:      user.Email,
		Password:   user.Password,
		Role:       BASIC.String(),
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

func (u User) Valid(ctx context.Context) errsx.Map {
	var errs errsx.Map
	return errs
}
