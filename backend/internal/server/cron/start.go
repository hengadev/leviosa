package cron

import (
	"context"
	"fmt"
	"log/slog"
)

// A function to just set the cron job
func (a *AppInstance) Start(ctx context.Context) error {
	jobs := []struct {
		name     string
		schedule string
		job      func() error
	}{
		// sec min hour dayofmonth month dayofweek
		{
			"parsing event for mail handling",
			"0 0 6 * * *",
			parseEvent(
				ctx,
				a.App.Svcs.Mail.SendRegistrationReminderEmail,
			),
		},
		{
			"handling vote closing",
			"0 0 6 * * *",
			checkCloseVote(ctx),
		},
		{
			"handling database replication",
			"0 0 24 * * *",
			backupDatabase(
				ctx,
				a.App.Svcs.Media.Repo.AddDatabaseReplica,
			),
		},
	}
	for _, job := range jobs {
		if _, err := a.cron.AddFunc(job.schedule, wrapJobWithErrorHandling(ctx, job.name, job.job, a.logger)); err != nil {
			return fmt.Errorf("failed to add cron job: %w", err)
		}
	}

	// start cron scheduler
	a.cron.Start()

	// wait for context cancellation
	<-ctx.Done()
	a.cron.Stop()
	return ctx.Err()
}

// wrapJobWithErrorHandling wraps a job function with error handling
func wrapJobWithErrorHandling(ctx context.Context, name string, job func() error, logger *slog.Logger) func() {
	return func() {
		if err := job(); err != nil {
			logger.WarnContext(ctx, fmt.Sprintf("cron job %s failed: %v\n", name, err))
		}
	}
}
