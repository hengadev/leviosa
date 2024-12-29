package userHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (h *AppInstance) ApproveUserRegistration(w http.ResponseWriter, r *http.Request) {
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
	// TODO:
	// get the email_hash and the role for the user
	input, err := serverutil.DecodeValid[models.UserPendingResponse](ctx, r.Body)

	// TODO: make the error right here for the client and for the logs
	if err != nil {
		switch {
		case errors.Is(err, serverutil.ErrDecodeJSON):
			logger.WarnContext(ctx, "failed to decode user", "error", err)
			serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		default:
			logger.WarnContext(ctx, "validate user")
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	}
	user, err := h.Svcs.User.CreateUser(ctx, &input)
	if err != nil {
		switch {
		default:
			logger.WarnContext(ctx, "failed to create account", "error", err)
			serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	}
	// send email to user to tell them that their account have been approved
	if errs := h.Svcs.Mail.WelcomeUser(ctx, user); len(errs) > 0 {
		logger.WarnContext(ctx, "failed to send welcome email to new added user", "error", err)
		serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		return

	}

	serverutil.WriteResponse(w, "user approved", http.StatusCreated)
}
