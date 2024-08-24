package userHandler

import (
	"context"
	// "fmt"
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
			slog.ErrorContext(ctx, "user decoding:", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "user decoding:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// validate credentials
		userID, role, err := h.Svcs.User.ValidateCredentials(ctx, &input)
		if err != nil {
			slog.ErrorContext(ctx, "user credentials validation:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		sessionID, err := h.Svcs.Session.CreateSession(ctx, userID, role.String())
		if err != nil {
			slog.ErrorContext(ctx, "create session:", "error", err)
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
