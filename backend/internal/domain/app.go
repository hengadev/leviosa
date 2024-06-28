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
	// session errors
	ErrInvalidUser = errors.New("invalid user")
)

// user errors
func NewInvalidInputErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidInput, err)
}

func NewAuthErr(err error) error {
	return fmt.Errorf("%w: %w", ErrAuth, err)
}

// session errors
func NewInvalidUserErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidUser, err)
}
