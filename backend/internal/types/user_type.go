package types

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"time"
)

const (
	SessionDuration   = time.Duration(30 * time.Minute)
	SessionCookieName = "session_token"
)

// Use to parse the information from the request from the /signup endpoint
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Use to parse the information from the request from the /signin endpoint
type UserForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Role       string `json:"role"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	Gender     string `json:"gender"`
	BirthDate  string `json:"birthdate"`
	Telephone  string `json:"telephone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCard string `json:"postalcard"`
}

// Use to store user information in the database
func NewUserStored(user *UserForm) *UserStored {
	return &UserStored{
		Id:         uuid.NewString(),
		Email:      user.Email,
		Password:   user.Password,
		Role:       "basic",
		LastName:   user.LastName,
		FirstName:  user.FirstName,
		Gender:     user.Gender,
		BirthDate:  user.BirthDate,
		Telephone:  user.Telephone,
		Address:    user.Address,
		City:       user.City,
		PostalCard: user.PostalCard,
	}
}

type UserStored struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	LastName   string `json:"lastname"`
	FirstName  string `json:"firstname"`
	Gender     string `json:"gender"`
	BirthDate  string `json:"birthdate"`
	Telephone  string `json:"telephone"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCard string `json:"postalcard"`
}

func (u *User) ValidateEmail() bool {
	_, err := mail.ParseAddress(u.Email)
	return err == nil
}

// TODO: Add the logic to verify if the password are good enough
func (u *User) ValidatePassword() bool {
	return true
}

func (u *UserStored) ValidateEmail() bool {
	_, err := mail.ParseAddress(u.Email)
	return err == nil
}

// TODO: Add the logic to verify if the password are good enough
func (u *UserStored) ValidatePassword() bool {
	return true
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
