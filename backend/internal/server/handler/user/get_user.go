package userHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) GetUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		user, err := a.Repos.User.FindAccountByID(ctx, userID)
		if err != nil {
			logger.WarnContext(ctx, "find user:", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err := serverutil.Encode(w, http.StatusFound, user); err != nil {
			logger.WarnContext(ctx, "send back user:", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
