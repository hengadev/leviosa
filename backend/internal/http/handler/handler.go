package handler

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("The requested resource was not found.")
	ErrForbidden = errors.New("You do not have permission to perform this action.")
	ErrInternal  = errors.New("An internal error occurred :")
)

func NewInternalErr(err error) string {
	return fmt.Sprintf("%w: w", ErrInternal, err)
}

func NewHandlerErr(err error) error {
	return fmt.Errorf("%w: %w", err)
}
