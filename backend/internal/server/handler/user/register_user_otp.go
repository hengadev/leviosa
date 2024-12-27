package userHandler

import (
	"context"
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

func (h *AppInstance) RegisterUserOTP() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "logger not found in context", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user, err := serverutil.DecodeValid[models.UserSignUp](r.Context(), r.Body)
		if err != nil {
			switch {
			case errors.Is(err, serverutil.ErrDecodeJSON):
				logger.WarnContext(ctx, "decode user", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			default:
				logger.WarnContext(ctx, "invalid sign up user", "error", err)
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}
		// TODO: I need to check the sent password against a list of leaked password
		if err := h.Svcs.User.CheckUser(ctx, user.Email); err != nil {
			switch {
			case errors.Is(err, domain.ErrNotFound):
				break
			case errors.Is(err, rp.ErrContext):
				logger.WarnContext(ctx, "context error, deadline or timeout while checking for user existence")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database checking for user existence query failed")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpted errror checking for user existence")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}
		emailHash, err := h.Svcs.User.CreateUnverifiedUser(ctx, &user)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrQueryFailed):
				logger.WarnContext(ctx, "database adding unverified user query failed")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrNotEncrypted):
				logger.WarnContext(ctx, "fail to encrypt unverified user")
				http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			case errors.Is(err, domain.ErrNotCreated):
				logger.WarnContext(ctx, "database adding unverified user query failed")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, rp.ErrContext):
				logger.WarnContext(ctx, "context error, deadline or timeout while adding unverified user")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			case errors.Is(err, domain.ErrUnexpectedType):
				logger.WarnContext(ctx, "unexpected errror adding unverified user")
				http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			}
			return
		}

		if err := h.generateAndSendOTP(ctx, w, logger, emailHash, user.Email); err != nil {
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

// TODO: do the error handling on that
func (h *AppInstance) generateAndSendOTP(
	ctx context.Context,
	w http.ResponseWriter,
	logger *slog.Logger,
	emailHash string,
	userEmail string,
) error {
	// generate OTP
	otp, err := h.Svcs.OTP.CreateOTP(ctx, emailHash)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database generate OTP failed")
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrMarshalJSON):
			logger.WarnContext(ctx, "marshal JSON OTP failed")
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error, deadline or timeout while adding unverified user")
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		default:
			logger.WarnContext(ctx, "generate OTP")
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
		}
		return err
	}
	// send email with OTP for user
	if errs := h.Svcs.Mail.SendOTP(ctx, userEmail, otp); len(errs) > 0 {
		logger.WarnContext(ctx, "send mail with OTP to specified user")
		http.Error(w, errsrv.NewInternalErr(errs), http.StatusInternalServerError)
		return err
	}
	return nil
}
