package app

import (
	"errors"
	"fmt"
)

var (
	//general
	ErrForbidden = errors.New("forbidden")
	// account errors
	ErrInvalidInput = errors.New("invalid input")
	// auth errors
	ErrAuth = errors.New("invalid auth")
	// user errors
	ErrInvalidUser       = errors.New("invalid user")
	ErrInvalidUserUpdate = errors.New("invalid user field")
	// session errors
	ErrInvalidSession = errors.New("invalid session")
)

// user errors
func NewInvalidInputErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidInput, err)
}

func NewAuthErr(err error) error {
	return fmt.Errorf("%w: %w", ErrAuth, err)
}

// user errors
func NewInvalidUserErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidUser, err)
}

func NewInvalidUserToUpdateErr(pbms map[string]string) error {
	errs := "["
	for k, v := range pbms {
		errs += fmt.Sprintf("{%s: %s}", k, v)
	}
	errs += "]"
	return fmt.Errorf("%w: %w", ErrInvalidUserUpdate, errors.New(errs))
}

// session errors
func NewInvalidSessionErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidSession, err)
}
