package cron

import (
	"log/slog"

	"github.com/GaryHY/event-reservation-app/internal/server/app"

	"github.com/robfig/cron/v3"
)

type AppInstance struct {
	App    *app.App
	cron   *cron.Cron
	logger *slog.Logger
}

func New(appCtx *app.App, logger *slog.Logger) *AppInstance {
	return &AppInstance{
		App:    appCtx,
		cron:   cron.New(), // Initialize cron in constructor
		logger: logger,
	}
}
