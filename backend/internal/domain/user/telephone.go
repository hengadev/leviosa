package userService

import (
	"errors"
	"regexp"
)

var invalidTelephone = regexp.MustCompile(`^0\d{9}$`)

type Telephone string

func ValidateTelephone(telephone string) error {
	if matches := invalidTelephone.FindAllString(telephone, -1); len(matches) == 0 {
		return errors.New("invalid telephone number")
	}
	return nil
}

func NewTelephone(telephone string) (Telephone, error) {
	if err := ValidateTelephone(telephone); err != nil {
		return "", err
	}
	return Telephone(telephone), nil
}

func (t Telephone) String() string {
	return string(t)
}
