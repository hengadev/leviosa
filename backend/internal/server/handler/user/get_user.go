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
			slog.ErrorContext(ctx, "userID string conversion to int:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		user, err := h.Repos.User.FindAccountByID(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, "find user:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err := serverutil.Encode(w, http.StatusFound, user); err != nil {
			slog.ErrorContext(ctx, "send back user:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
