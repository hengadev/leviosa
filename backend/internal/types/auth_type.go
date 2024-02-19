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

// Use to parse the information from the request
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Use to store user information in the database
type UserStored struct {
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

type Role string

func ConvertToRole(str string) Role {
	switch str {
	case "admin":
		return ADMIN
	case "helper":
		return HELPER
	default:
		return BASIC
	}
}

const (
	ADMIN  = Role("admin")
	BASIC  = Role("basic")
	HELPER = Role("helper")
)

// NOTE: Les roles cela va etre admin, helper, basic

// Create a new session, used in tests
func NewSession(user *User) *Session {
	return &Session{
		Id:         uuid.NewString(),
		Email:      user.Email,
		Created_at: time.Now(),
	}
}

type Session struct {
	Id         string    `json:"id"`
	Email      string    `json:"email"`
	Created_at time.Time `json:"created_at"`
}

// A function to check whether a session is expired.
func (s *Session) isExpired() bool {
	return s.Created_at.Before(time.Now())
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
