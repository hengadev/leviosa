package cron

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// Une fonction pour realiser des actions a l'approche de certaines dates.
func parseEvent(
	ctx context.Context,
	emailSenderFn func(context.Context, *models.User, string, int) errsx.Map,
) func() error {
	return func() error {
		// TODO:
		// 1. get list of future registrations that you need to send a reminder email for
		// 2. get list of all users concerned with emails and  the corresponding dates
		// 3. send emails with right templates
		// if errs := emailSenderFn(ctx); len(errs) > 0 {
		// 	return errs
		// }
		return nil
	}
}
