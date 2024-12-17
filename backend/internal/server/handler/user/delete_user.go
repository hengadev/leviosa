package userHandler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
)

func (a *AppInstance) DeleteUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// TODO:
		// - get the userID
		userIDstr := ctx.Value(mw.UserIDKey).(string)
		userID, err := strconv.Atoi(userIDstr)
		if err != nil {
			slog.ErrorContext(ctx, "userID string conversion to int:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// get the sessionID from the cookie
		sessionID := r.Cookies()[0].Value
		// - call the user repo to delete the user account from the user repository
		err = a.Svcs.User.DeleteUser(ctx, userID)
		if err != nil {
			fmt.Println("error in deleting the user")
			slog.ErrorContext(ctx, "delete user:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// - call the service repo to delete all user's session
		err = a.Svcs.Session.RemoveSession(ctx, sessionID)
		if err != nil {
			slog.ErrorContext(ctx, "delete user session:", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
