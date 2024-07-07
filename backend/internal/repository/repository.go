package repository

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrRessourceCreation = errors.New("ressource not created")
	ErrRessourceUpdate   = errors.New("ressource not updated")
	ErrRessourceDelete   = errors.New("ressource not deleted")
	ErrBadQuery          = errors.New("bad query")
	ErrRows              = errors.New("rows error")
	ErrScan              = errors.New("scan error")
)

func NewNotFoundError(err error) error {
	return fmt.Errorf("%w: %w", ErrNotFound, err)
}

func NewRessourceCreationErr(err error) error {
	return fmt.Errorf("%w: %w", ErrRessourceCreation, err)
}

func NewRessourceUpdateErr(err error) error {
	return fmt.Errorf("%w: %w", ErrRessourceUpdate, err)
}

func NewRessourceDeleteErr(err error) error {
	return fmt.Errorf("%w: %w", ErrRessourceDelete, err)
}
func NewBadQueryErr(err error) error {
	return fmt.Errorf("%w: %w", ErrBadQuery, err)
}

func NewErrRow(err error) error {
	return fmt.Errorf("%w: %w", ErrRows, err)
}

func NewErrScan(err error) error {
	return fmt.Errorf("%w: %w", ErrScan, err)
}
