package userHandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	// "time"

	// "github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
)

func (a *AppInstance) HandleOAuth() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// get provider to have right OAuthUser type
		provider := r.PathValue("provider")
		if provider == "" {
			var err error = fmt.Errorf("missing provider in request")
			slog.ErrorContext(ctx, "oauth handling", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusBadRequest)
			return
		}
		// decode user sent and validate it.
		oauthUser := userService.DecodeOAuthUser(r, provider)
		user, err := a.Repos.User.GetOAuthUser(ctx, oauthUser.GetEmail(), provider)
		_ = user
		// TODO: find right status to send back with this because I do not know
		if errors.Is(err, rp.ErrNotFound) {
			fmt.Println("this is a not found error")
			user, err = a.Svcs.User.CreateOAuthAccount(ctx, oauthUser)
			if err != nil {
				slog.ErrorContext(ctx, "create oauth account", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
				return
			}
		}
		if err != nil {
			slog.ErrorContext(ctx, "create oauth account", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusBadRequest)
			return
		}
		// sessionID, err := h.Svcs.Session.CreateSession(ctx, user.ID, user.Role)
		// if err != nil {
		// 	slog.ErrorContext(ctx, "failed to create session", "error", err)
		// 	http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		// 	return
		// }
		// http.SetCookie(w, &http.Cookie{
		// 	Name:     sessionService.SessionName,
		// 	Value:    sessionID,
		// 	Expires:  time.Now().Add(sessionService.SessionDuration),
		// 	HttpOnly: true,
		// })
		w.WriteHeader(http.StatusCreated)
	})
}
