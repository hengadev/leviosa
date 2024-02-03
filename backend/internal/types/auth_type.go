package types

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/mail"
	"time"
)

const (
	SessionDuration   = time.Duration(30 * time.Minute)
	SessionCookieName = "session_token"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NOTE: Dans le tuto tout cela est private (struct + fields)
type AuthUser struct {
	Email        string `json:"email"`
	HashPassword string `json:"hashpassword"`
}

func NewSession(user User) *Session {
	return &Session{
		Email:      user.Email,
		Created_at: time.Now().Format(time.RFC822),
		Expiry:     SessionDuration,
	}
}

type Session struct {
	Email      string
	Created_at string
	Expiry     time.Duration
}

func (s *Session) IsExpired() bool {
	createdTime, err := time.Parse(time.RFC822Z, s.Created_at)
	if err != nil {
		log.Fatal("Cannot parse the time - ", err)
	}
	return createdTime.Add(s.Expiry).Before(time.Now())
}

// NOTE: Is using the standard library enough ?
func (u User) ValidateEmail() bool {
	_, err := mail.ParseAddress(u.Email)
	return err == nil
}

// Password should be X long with
func (u User) ValidatePassword() bool {
	return true
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
