package userHandler

import (
	"errors"
	"fmt"
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
			serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("get otp and email from payload")
		userOTP, err := serverutil.DecodeValid[models.UserOTP](r.Context(), r.Body)
		if err != nil {
			switch {
			case errors.Is(err, serverutil.ErrDecodeJSON):
				logger.WarnContext(ctx, "decode user", "error", err)
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			default:
				logger.WarnContext(ctx, "invalid sign up user", "error", err)
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}
		fmt.Printf("the userOTP that I get: %#+v\n", userOTP)
		// hash email
		emailHash := security.HashEmail(userOTP.Email)

		// validate otp
		if err = h.Svcs.OTP.ValidateOTP(ctx, emailHash, userOTP.OTP); err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				logger.WarnContext(ctx, "OTP validation query failed due to database error")
				serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, rp.ErrContext):
				logger.WarnContext(ctx, "context error, deadline or timeout while adding pending user")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database compare OTPs failed")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrInvalidValue):
				logger.WarnContext(ctx, "invalid OTP provided")
				serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpected error while validation OTP sent")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}

		// add to pending user (using the hash email)
		if err = h.Svcs.User.CreatePendingUser(ctx, userOTP.Email); err != nil {
			switch {
			case errors.Is(err, rp.ErrContext):
				logger.WarnContext(ctx, "context error, deadline or timeout while adding pending user")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrNotFound):
				logger.WarnContext(ctx, "user not found in unverified_users table")
				serverutil.WriteResponse(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpected error while adding user to pending_users table")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrNotCreated):
				logger.WarnContext(ctx, "database creating user to pending_users table failed")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database adding user to pending_users table failed")
				serverutil.WriteResponse(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}

		serverutil.WriteResponse(w, "pending user successfully created", http.StatusCreated)
	})
}
