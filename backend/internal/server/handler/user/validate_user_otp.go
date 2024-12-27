package userHandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/internal/domain/user/security"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (h *AppInstance) ValidateUserOTP() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO:
		// - get otp and email from userOTP
		userOTP, err := serverutil.DecodeValid[models.UserOTP](r.Context(), r.Body)
		if err != nil {
			switch {
			case errors.Is(err, serverutil.ErrDecodeJSON):
				logger.WarnContext(ctx, "decode user", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
				return
			default:
				logger.WarnContext(ctx, "invalid sign up user", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
				return
			}
		}
		// hash email
		emailHash := security.HashEmail(userOTP.Email)

		// validate otp
		if err = h.Svcs.OTP.ValidateOTP(ctx, emailHash, userOTP.OTP); err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				logger.WarnContext(ctx, "provided OTP not found in database")
				http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, rp.ErrContext):
				logger.WarnContext(ctx, "context error, deadline or timeout while adding unverified user")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database compare OTPs failed")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrInvalidValue):
				logger.WarnContext(ctx, "invalid OTP provided")
				http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpected error while validation OTP sent")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
		}

		// add to pending user (using the hash email)
		if err = h.Svcs.User.CreatePendingUser(ctx, userOTP.Email); err != nil {
			switch {
			case errors.Is(err, rp.ErrContext):
				logger.WarnContext(ctx, "context error, deadline or timeout while adding pending user")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrNotFound):
				logger.WarnContext(ctx, "user not found in unverified_users table")
				http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpected error while adding user to pending_user table")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrNotCreated):
				logger.WarnContext(ctx, "database creating user to pending_user table failed")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database adding user to pending_user table failed")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
		}
		w.WriteHeader(http.StatusCreated)
	})
}
