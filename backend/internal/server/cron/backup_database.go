package cron

import (
	"context"
	"errors"
	"mime/multipart"
	"os"

	"github.com/GaryHY/event-reservation-app/internal/domain"
)

// Une fonction pour backer mes BDD. Faire les migrations en gros.
func backupDatabase(ctx context.Context, addDatabaseReplicaFn func(context.Context, multipart.File) error) func() error {
	return func() error {
		// TODO: find the right path for that file
		file, err := os.Open("db.db")
		if err != nil {
			// TODO: do something with the error, so that I can return it and use the logger where I define the cron job
		}
		if err := addDatabaseReplicaFn(ctx, file); err != nil {
			// TODO: do a better error handling in here
			switch {
			case errors.Is(err, domain.ErrNotCreated):
				return err
			}
		}
		return nil
	}
}
