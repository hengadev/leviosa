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
			logger.ErrorContext(ctx, "user locked, too much attempts")
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}

		input, err := a.decodeAndValidateCredentials(ctx, w, r, logger)
		if err != nil {
			logger.ErrorContext(ctx, "check if user is locked")
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}

		if err = a.checkThrottling(ctx, w, input.Email, logger); err != nil {
			return
		}

		if err = a.validateUserCredentials(ctx, w, input, logger); err != nil {
			return
		}

		if err = a.createAndSetSession(ctx, w, input.Email, logger); err != nil {
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
