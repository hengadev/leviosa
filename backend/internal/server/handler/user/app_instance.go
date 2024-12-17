package userHandler

import (
	"github.com/GaryHY/event-reservation-app/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func New(appCtx *app.App) *AppInstance {
	return &AppInstance{appCtx}
}
