package userService

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Source struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Etag       string `json:"etag"`
	UpdateTime string `json:"updateTime"`
}

type Metadata struct {
	Source        Source `json:"source"`
	Primary       bool   `json:"primary"`
	Verified      bool   `json:"verified"`
	SourcePrimary bool   `json:"sourcePrimary"`
}

type Date struct {
	Month int `json:"month"`
	Day   int `json:"day"`
	Year  int `json:"year"`
}

type Birthday struct {
	Metadata Metadata `json:"metadata"`
	Date     Date     `json:"date"`
	Text     string   `json:"text"`
}

type Gender struct {
	FormattedValue string   `json:"formattedValue"`
	Value          string   `json:"value"`
	Metadata       Metadata `json:"metadata"`
}

type PhoneNumber struct {
	Metadata      Metadata `json:"metadata"`
	Value         string   `json:"value"`
	CanonicalForm string   `json:"canonicalForm"`
	Type          string   `json:"type"`
	FormattedType string   `json:"formattedType"`
}

type OAuthUser interface {
	ToUser() (*User, error)
	GetEmail() string
	HasVerifiedEmail() bool
}

func DecodeOAuthUser(r *http.Request, provider string) OAuthUser {
	switch provider {
	case "google":
		var googleUser GoogleUser
		json.NewDecoder(r.Body).Decode(&googleUser)
		return &googleUser
	case "apple":
		var appleUser AppleUser
		json.NewDecoder(r.Body).Decode(&appleUser)
		return &appleUser
	}
	return nil
}

type GoogleUser struct {
	Sub           string        `json:"sub"`
	Name          string        `json:"name"`
	GivenName     string        `json:"given_name"`
	FamilyName    string        `json:"family_name"`
	Picture       string        `json:"picture"`
	Email         string        `json:"email"`
	EmailVerified bool          `json:"email_verified"`
	ResourceName  string        `json:"resourceName"`
	Etag          string        `json:"etag"`
	Genders       []Gender      `json:"genders"`
	Birthdays     []Birthday    `json:"birthdays"`
	PhoneNumbers  []PhoneNumber `json:"phoneNumbers"`
}

func (g *GoogleUser) ToUser() (*User, error) {
	var user User
	email, pbms := NewEmail(g.Email)
	if len(pbms) > 0 {
		return nil, pbms
	}
	user.LastName = g.FamilyName
	user.FirstName = g.GivenName
	user.Email = email.String()
	date := g.Birthdays[0].Date
	year := date.Year
	month, err := convIntToMonth(date.Month)
	if err != nil {
		fmt.Println("month conversion:", err)
	}
	day, err := convIntToDay(date.Day, date.Month)
	if err != nil {
		fmt.Println("day conversion:", err)
	}
	birthday := fmt.Sprintf("%d-%s-%s", year, month, day)
	if err != nil {
		fmt.Println("birthday formatting:", err)
	}
	user.BirthDate = birthday
	user.Gender = convToGender(g.Genders[0].Value)
	user.Telephone = g.PhoneNumbers[0].Value
	user.OAuthProvider = "google"
	user.OAuthID = g.Sub
	return &user, nil
}

func (g *GoogleUser) GetEmail() string {
	return g.Email
}

func (g *GoogleUser) HasVerifiedEmail() bool {
	return g.EmailVerified
}

type AppleUser struct {
}

func (a *AppleUser) ToUser() (*User, error) {
	var user *User
	return user, nil
}

func (a *AppleUser) GetEmail() string {
	return ""
}

func (a *AppleUser) HasVerifiedEmail() bool {
	return true
}

// TODO: remove that, I no longer user the oauth handlers
type Response struct {
	Birthdays     []Birthday    `json:"birthdays"`
	Genders       []Gender      `json:"genders"`
	PhoneNumbers  []PhoneNumber `json:"phoneNumbers"`
	RessourceName string        `json:"ressourceName"`
	Etag          string        `json:"etag"`
}
