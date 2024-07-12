package app

import (
	"errors"
	"fmt"
)

type pbms map[string]string

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

// session errors
func NewInvalidSessionErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidSession, err)
}
