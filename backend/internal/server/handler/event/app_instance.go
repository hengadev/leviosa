package eventHandler

import (
	"github.com/GaryHY/leviosa/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func New(appCtx *app.App) *AppInstance {
	return &AppInstance{appCtx}
}
