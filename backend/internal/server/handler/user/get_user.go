package userHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/leviosa/internal/domain"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) GetUser(w http.ResponseWriter, r *http.Request) {
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
	user, err := a.Svcs.User.FindUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			logger.WarnContext(ctx, handler.NewNotFoundErr(err))
			serverutil.WriteResponse(w, handler.NewNotFoundErr(err), http.StatusNotFound)
		case errors.Is(err, rp.ErrContext), errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, handler.NewInternalErr(err))
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrInvalidValue):
			logger.WarnContext(ctx, err.Error())
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		default:
			logger.WarnContext(ctx, "find user by ID:", "error", err)
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err := serverutil.Encode(w, http.StatusOK, user); err != nil {
		logger.WarnContext(ctx, "get found user:", "error", err)
		serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		return
	}
}
