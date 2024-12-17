package userHandler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/google/uuid"
)

func (a *AppInstance) Signout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get session value from cookie
		sessionID := r.Cookies()[0].Value
		if err := uuid.Validate(sessionID); err != nil {
			slog.ErrorContext(ctx, "get sessionID from cookie:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusBadRequest)
			return
		}
		// remove session with sessionID
		if err := a.Svcs.Session.RemoveSession(ctx, sessionID); err != nil {
			slog.ErrorContext(ctx, "remove session:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send cookie to inform client that cookie is no longer valid
		http.SetCookie(w, &http.Cookie{
			Name:    sessionService.SessionName,
			Value:   "",
			Expires: time.Now(),
		})
	})
}
