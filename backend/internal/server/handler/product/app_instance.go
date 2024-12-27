package productHandler

import (
	"github.com/GaryHY/event-reservation-app/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func NewHandler(appCtx *app.App) *AppInstance {
	return &AppInstance{appCtx}
}
