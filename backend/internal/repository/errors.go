package repository

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrNotCreated = errors.New("not created")
	ErrNotUpdated = errors.New("not updated")
	ErrNotDeleted = errors.New("not deleted")
	ErrDatabase   = errors.New("database error")
	ErrInternal   = errors.New("internal error")
	ErrContext    = errors.New("context related error")
)

func NewContextError(err error) error {
	return fmt.Errorf("%w: %w", ErrContext, err)
}

func NewInternalError(err error) error {
	return fmt.Errorf("%w: %w", ErrNotFound, err)
}

func NewNotFoundError(err error, domainName string) error {
	return fmt.Errorf("%s %w: %w", domainName, ErrNotFound, err)
}

func NewNotCreatedErr(err error, domainName string) error {
	return fmt.Errorf("%s %w: %w", domainName, ErrNotCreated, err)
}

func NewNotUpdatedErr(err error, domainName string) error {
	return fmt.Errorf("%s %w: %w", domainName, ErrNotUpdated, err)
}

func NewNotDeletedErr(err error, domainName string) error {
	return fmt.Errorf("%s %w: %w", domainName, ErrNotDeleted, err)
}

func NewDatabaseErr(err error) error {
	return fmt.Errorf("%w: %w", ErrDatabase, err)
}
