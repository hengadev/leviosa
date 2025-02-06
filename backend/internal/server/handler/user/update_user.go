package userHandler

import (
	"errors"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, ok := ctx.Value(contextutil.UserIDKey).(string)
	if !ok {
		logger.ErrorContext(ctx, "user ID not found in context")
		serverutil.WriteResponse(w, errors.New("failed to get user ID from context").Error(), http.StatusInternalServerError)
		return
	}

	// use a custom valid for the updtate thing
	user, err := serverutil.Decode[models.User](r.Body)
	if err != nil {
		logger.ErrorContext(ctx, "failed to decode user", "error", err)
		serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
	// modify user
	if err = a.Svcs.User.UpdateAccount(ctx, &user, userID); err != nil {
		switch err {
		// TODO: handle the validation error to just send back the fields that are not updated because prohibited from updates
		default:
			logger.ErrorContext(ctx, "failed to modify the user", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	}
}
