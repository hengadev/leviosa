package errsrv

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("The requested resource was not found.")
	ErrForbidden         = errors.New("You do not have permission to perform this action.")
	ErrInternal          = errors.New("An internal error occurred.")
	ErrBadRequest        = errors.New("This is bad request.")
	ErrServiceUnvailable = errors.New("Service unavailable.")
)

func NewInternalErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrInternal, err)
}

func NewBadRequestErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrBadRequest, err)
}

func NewForbiddenErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrForbidden, err)
}

func NewServiceUnavailableErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrServiceUnvailable, err)
}
