package cron

import (
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/server/app"

	"github.com/robfig/cron"
)

type Handler struct {
	*app.App
}

func NewHandler(handler *app.App) *Handler {
	return &Handler{handler}
}

// A function to just set the cron job friend
func (h *Handler) Start() error {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return fmt.Errorf("get location from time module: %w", err)
	}
	cronJob := cron.NewWithLocation(loc)
	defer cronJob.Stop()

	// sec min hour dayofmonth month dayofweek
	cronJob.AddFunc("0 1 * * * *", testPrint("Gary"))
	cronJob.AddFunc("0 0 6 * * *", parseEvent)
	cronJob.AddFunc("0 0 6 * * *", checkCloseVote)
	cronJob.AddFunc("0 0 6 * * *", backupDatabase)
	cronJob.Start()
	select {}
}

func testPrint(name string) func() {
	return func() {
		fmt.Printf("Je teste la fonction avec le name: %s", name)
	}
}
