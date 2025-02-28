package register

import (
	"github.com/hengadev/leviosa/internal/server/app"
)

type AppInstance struct {
	*app.App
}

func NewHandler(appCtx *app.App) *AppInstance {
	return &AppInstance{appCtx}
}
