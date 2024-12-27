package userHandler

import (
	"errors"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (a *AppInstance) UpdateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userID, ok := ctx.Value(contextutil.UserIDKey).(string)
		if !ok {
			logger.ErrorContext(ctx, "user ID not found in context")
			http.Error(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
			return
		}

		// use a custom valid for the updtate thing
		user, err := serverutil.Decode[models.User](r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// modify user
		if err = a.Svcs.User.UpdateAccount(ctx, &user, userID); err != nil {
			logger.ErrorContext(ctx, "failed to modify the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}
