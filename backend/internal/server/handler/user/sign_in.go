package userHandler

import (
	"context"
	"errors"
	"net/http"
	"time"

	app "github.com/GaryHY/event-reservation-app/internal/domain"
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
			h.Logger.ErrorContext(ctx, "user decoding:", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			h.Logger.ErrorContext(ctx, "user decoding:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		isLocked, err := h.Repos.Throttler.IsLocked(ctx, input.Email)
		if isLocked {
			h.Logger.ErrorContext(ctx, "user locked, too much attempts")
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			h.Logger.ErrorContext(ctx, "check if user is locked")
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}

		// validate credentials
		err = h.Svcs.User.ValidateCredentials(ctx, &input)
		switch {
		case errors.Is(err, app.ErrUserNotFound):
			h.Logger.ErrorContext(ctx, "invalid email", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		case errors.Is(err, app.ErrQueryFailed):
			h.Logger.ErrorContext(ctx, "database error", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		case err != nil:
			attemptErr := h.Svcs.Throttler.RegisterAttempt(ctx, input.Email)
			h.Logger.ErrorContext(ctx, "invalid password", "error", errors.Join(err, attemptErr))
			http.Error(w, errsrv.NewBadRequestErr(errors.Join(err, attemptErr)), http.StatusBadRequest)
			return
		}

		userID, role, err := h.Svcs.User.GetUserSessionData(ctx, input.Email)
		if err != nil {
			h.Logger.ErrorContext(ctx, "get session related user data", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		session, err := h.Svcs.Session.CreateSession(ctx, userID, role)
		if err != nil {
			h.Logger.ErrorContext(ctx, "create session:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}

		if err := h.Svcs.Throttler.Reset(ctx, input.Email); err != nil {
			h.Logger.ErrorContext(ctx, "reset throttler", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionService.SessionName,
			Value:    session.ID,
			Expires:  time.Now().Add(sessionService.SessionDuration),
			HttpOnly: true,
		})
	})
}
