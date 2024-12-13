package repository

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("ressource not found")
	ErrRessourceCreation = errors.New("ressource not created")
	ErrRessourceUpdate   = errors.New("ressource not updated")
	ErrRessourceDelete   = errors.New("ressource not deleted")
	ErrDatabase          = errors.New("database error")
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

func NewDatabaseErr(err error) error {
	return fmt.Errorf("%w: %w", ErrDatabase, err)
}
