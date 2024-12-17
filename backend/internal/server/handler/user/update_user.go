package userHandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) UpdateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		userID := ctx.Value(mw.UserIDKey).(int)

		// use a custom valid for the updtate thing
		user, err := serverutil.Decode[userService.User](r.Body)
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// modify user
		if err = a.Svcs.User.UpdateAccount(ctx, &user, userID); err != nil {
			slog.ErrorContext(ctx, "failed to modify the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
