package user

import (
	"errors"
	"fmt"
)

type Password string

const passwordMinLength = 8

func NewPassword(p string) (Password, error) {
	// TODO: do some validation on the password
	if len(p) < passwordMinLength {
		return "", errors.New(fmt.Sprintf("password too short, expect at least %d caracter", passwordMinLength))
	}
	return Password(p), nil
}

func (p Password) String() string {
	return string(p)
}
