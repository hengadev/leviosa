package types

import (
	"golang.org/x/crypto/bcrypt"
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

func (u User) ValidateEmail() bool {
	return true
}

// Password should be X long with
func (u User) ValidatePassword() bool {
	return true
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
