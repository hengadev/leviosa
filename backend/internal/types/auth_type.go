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
	// Role     string `json:"role"`
	// Telephone string `json:"telephone"`
	// Address   string `json:"address"`
	// Gender   string `json:"gender"`
	// BirthDate      string `json:"birthdate"`
	// LastName      string `json:"lastname"`
	// FirstName      string `json:"firstname"`
}

// Use to store user information in the database
type UserSignUp struct {
	Id        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	Telephone string `json:"telephone"`
	Address   string `json:"address"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birthdate"`
	LastName  string `json:"lastname"`
	FirstName string `json:"firstname"`
}

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

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
