package user

import (
	"errors"
	"fmt"
)

type Password string

const passwordMinLength = 8

func ValidatePassword(p string) error {
	if len(p) < passwordMinLength {
		return errors.New(fmt.Sprintf("password too short, expect at least %d caracter", passwordMinLength))
	}
	return nil
}
func NewPassword(p string) (Password, error) {
	if err := ValidatePassword(p); err != nil {
		return "", err
	}
	return Password(p), nil
}

func (p Password) String() string {
	return string(p)
}
