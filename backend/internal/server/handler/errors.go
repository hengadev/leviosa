package errsrv

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("The requested resource was not found.")
	ErrInternal          = errors.New("An internal error occurred:")
	ErrBadRequest        = errors.New("This is bad request.")
	ErrServiceUnvailable = errors.New("Service unavailable.")
)

func NewInternalErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrInternal, err)
}

func NewBadRequestErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrBadRequest, err)
}

func NewServiceUnavailableErr(err error) string {
	return fmt.Sprintf("%s: %s", ErrServiceUnvailable, err)
}

var (
	ErrGet    = errors.New("find")
	ErrCreate = errors.New("create")
	ErrUpdate = errors.New("modify")
	ErrDelete = errors.New("remove")
)

func NewGetErr(domain string, err error) string {
	return fmt.Sprintf("%s %s: %s", ErrGet, domain, err.Error())
}

func NewCreateErr(domain string, err error) string {
	return fmt.Sprintf("%s %s: %s", ErrCreate, domain, err.Error())
}
func NewUpdateErr(domain string, err error) string {
	return fmt.Sprintf("%s %s: %s", ErrUpdate, domain, err.Error())
}

func NewDeleteErr(domain string, err error) string {
	return fmt.Sprintf("%s %s: %s", ErrDelete, domain, err.Error())
}
