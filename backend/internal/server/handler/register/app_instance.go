package register

import (
	"github.com/GaryHY/leviosa/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func NewHandler(appCtx *app.App) *AppInstance {
	return &AppInstance{appCtx}
}
