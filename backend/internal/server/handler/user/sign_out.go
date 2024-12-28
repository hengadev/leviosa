package userHandler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) Signout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get session value from cookie
		cookie, err := r.Cookie(sessionService.SessionName)
		if err != nil {
			logger.ErrorContext(ctx, "get session cookie for signout", "error", err)
			serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return

		}
		sessionID := cookie.Value

		// remove session with sessionID
		if err := a.Svcs.Session.RemoveSession(ctx, sessionID); err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidValue):
				logger.WarnContext(ctx, "invalid value in session validation")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database removing session for user query failed")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrNotFound):
				logger.WarnContext(ctx, "user session not found")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpected error removing user session")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
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
