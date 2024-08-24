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

func (h *Handler) UpdateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		userID := ctx.Value(mw.UserIDKey).(string)

		user, pbms, err := serverutil.DecodeValid[userService.User](r)
		if len(pbms) > 0 {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// modify user
		if err = h.Svcs.User.UpdateAccount(ctx, &user, userID); err != nil {
			slog.ErrorContext(ctx, "failed to modify the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
