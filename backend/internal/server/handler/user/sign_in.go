package userHandler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (h *Handler) Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// parse the request body
		input, pbms, err := serverutil.DecodeValid[userService.Credentials](r)
		if len(pbms) > 0 {
			slog.ErrorContext(ctx, "failed to authenticate the user, bad request", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to authenticate the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// validate credentials
		userID, role, err := h.Svcs.User.ValidateCredentials(ctx, &input)
		if err != nil {
			slog.ErrorContext(ctx, "failed to validate user credentials", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// create session
		sessionID, err := h.Svcs.Session.CreateSession(ctx, userID, role.String())
		if err != nil {
			slog.ErrorContext(ctx, "failed to create session", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessionService.SessionName,
			Value:    sessionID,
			Expires:  time.Now().Add(sessionService.SessionDuration),
			HttpOnly: true,
		})
	})
}
