package userHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (h *AppInstance) GetUsersToApprove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := contextutil.ValidateRoleInContext(ctx, models.ADMINISTRATOR); err != nil {
		logger.ErrorContext(ctx, "get role from request", "error", err)
		serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}

	users, err := h.App.Svcs.User.GetAllPendingUsers(ctx)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database getting pending users list failed")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, "unexpected error while getting pending users list")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error, deadline or timeout while getting pending users list")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		}
		return
	}
	if err := serverutil.Encode(w, int(http.StatusOK), users); err != nil {
		logger.ErrorContext(ctx, "failed to encode pendings users list", "error", err)
		serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
	}
}
