package userHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, ok := ctx.Value(contextutil.UserIDKey).(string)
	if !ok {
		logger.ErrorContext(ctx, "user ID not found in context")
		serverutil.WriteResponse(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
		return
	}

	// get the sessionID from the cookie
	sessionID := r.Cookies()[0].Value
	// - call the user repo to delete the user account from the user repository
	err = a.Svcs.User.DeleteUser(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "delete user:", "error", err)
		serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
	// - call the service repo to delete all user's session
	err = a.Svcs.Session.RemoveSession(ctx, sessionID)
	if err != nil {
		logger.ErrorContext(ctx, "delete user session:", "error", err)
		serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
