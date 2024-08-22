package userHandler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (h *Handler) GetUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		userIDstr := ctx.Value(mw.SessionIDKey).(string)
		userID, err := strconv.Atoi(userIDstr)
		if err != nil {
			slog.ErrorContext(ctx, "failed to convert string userID to int", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// userID := ctx.Value(mw.SessionIDKey).(int)
		user, err := h.Repos.User.FindAccountByID(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err := serverutil.Encode(w, http.StatusFound, user); err != nil {
			slog.ErrorContext(ctx, "failed to send the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
