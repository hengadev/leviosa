package app

import (
	"errors"
	"fmt"
)

// here we deal with domain errors

// database error
// validation error (email, password etc...)
// ressource not found error

var (
	//general
	ErrForbidden   = errors.New("forbidden")
	ErrQueryFailed = errors.New("database query execution failed")
	// account errors
	ErrInvalidInput = errors.New("invalid input")
	// auth errors
	ErrAuth = errors.New("invalid authentication")
	// user errors
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidUser       = errors.New("invalid user")
	ErrInvalidUserUpdate = errors.New("invalid user field")
	// session errors
	ErrInvalidSession = errors.New("invalid session")
)

	//general
	ErrMarshalJSON = errors.New("json marshalling")
)

func NewJSONMarshalErr(err error) error {
	return fmt.Errorf("%w: %w", ErrMarshalJSON, err)
}

func NewQueryFailedErr(err error) error {
	return fmt.Errorf("%w: %w", ErrQueryFailed, err)
}

func NewUserNotFoundErr(err error) error {
	return fmt.Errorf("%w: %w", ErrUserNotFound, err)
}

func NewInvalidUserErr(err error) error {
// user errors
// user errors
	return fmt.Errorf("%w: %w", ErrInvalidUser, err)
}

// session errors
func NewInvalidSessionErr(err error) error {
	return fmt.Errorf("%w: %w", ErrInvalidSession, err)
}
