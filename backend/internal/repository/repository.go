package repository

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrRessourceCreation = errors.New("ressource not created")
)

func NewNotFoundError(err error) error {
	return fmt.Errorf("%w: %w", ErrNotFound, err)
}

func NewRessourceCreationErr(err error) error {
	return fmt.Errorf("%w: %w", ErrRessourceCreation, err)
}
