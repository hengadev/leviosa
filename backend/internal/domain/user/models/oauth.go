package models

import (
	"context"
	"time"

	"github.com/GaryHY/event-reservation-app/pkg/errsx"
	"github.com/google/uuid"
)

type OAuthUser interface {
	mustBeOAuthUser()
	ToUser() *User
}

type GoogleUser struct {
	GoogleID    string `json:"googleID"`
	Name        string `json:"name"`
	GivenName   string `json:"given_name"`
	FamilyName  string `json:"family_name"`
	Picture     string `json:"picture"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
	Birthday    string `json:"birthday"`
	PhoneNumber string `json:"phoneNumber"`
	Address     struct {
		Formatted     string `json:"formatted"`
		StreetAddress string `json:"streetAddress"`
		Locality      string `json:"locality"`
		Region        string `json:"region"`
		PostalCode    string `json:"postalCode"`
		Country       string `json:"country"`
	} `json:"address"`
}

func (g GoogleUser) mustBeOAuthUser() {}

type AppleUser struct {
	AppleID string `json:"apple_id"`
}

func (a AppleUser) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}

func (a AppleUser) mustBeOAuthUser() {}

func (a AppleUser) ToUser() *User {
	return &User{}
}

// ToUser create a new user from a GoogleUser
func (g GoogleUser) ToUser() *User {
	birthdate, _ := time.Parse("2006-01-02", g.Birthday)
	return &User{
		ID:         uuid.NewString(),
		Email:      g.Email,
		BirthDate:  birthdate,
		LastName:   g.FamilyName,
		FirstName:  g.GivenName,
		Gender:     g.Gender,
		Telephone:  g.PhoneNumber,
		PostalCode: g.Address.PostalCode,
		City:       g.Address.Locality,
		Address1:   g.Address.StreetAddress,
		GoogleID:   g.GoogleID,
	}
}

// ToUser create a new user from a GoogleUser
func (g GoogleUser) ToUserPending() *UserPending {
	return &UserPending{
		Email:     "",
		LastName:  g.FamilyName,
		FirstName: g.GivenName,
		GoogleID:  g.GoogleID,
		AppleID:   "",
	}
}

func (g GoogleUser) Valid(ctx context.Context) (problems errsx.Map) {
	var errs errsx.Map
	return errs
}
