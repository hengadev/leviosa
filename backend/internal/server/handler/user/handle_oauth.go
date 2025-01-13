package userHandler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/leviosa/internal/domain"
	"github.com/GaryHY/leviosa/internal/domain/session"
	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/internal/domain/user/security"
	rp "github.com/GaryHY/leviosa/internal/repository"
	"github.com/GaryHY/leviosa/internal/server/handler"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

func (a *AppInstance) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger, err := contextutil.GetLoggerFromContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "logger not found in context", "error", err)
		serverutil.WriteResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var provider models.ProviderType
	inputProvider := r.PathValue("provider")
	if err = provider.Set(inputProvider); err != nil {
		logger.ErrorContext(ctx, "invalid provider", "error", err)
		serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		return
	}

	user, err := a.decodeAndValidUser(ctx, w, r.Body, logger, provider)
	if err != nil {
		return
	}
	if err := a.handleUser(ctx, w, user, logger, provider); err != nil {
		return
	}
	if err := a.handleSession(ctx, w, user.Email, logger); err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *AppInstance) decodeAndValidUser(ctx context.Context, w http.ResponseWriter, body io.ReadCloser, logger *slog.Logger, provider models.ProviderType) (*models.User, error) {
	// put that into a decodeAndValid function
	var oauthUser models.OAuthUser
	var user *models.User
	var err error
	switch provider {
	case models.Google:
		oauthUser, err = serverutil.DecodeValid[models.GoogleUser](ctx, body)
		user = oauthUser.ToUser()
	case models.Apple:
		oauthUser, err = serverutil.DecodeValid[models.AppleUser](ctx, body)
		user = oauthUser.ToUser()
	}
	if err != nil {
		switch {
		case errors.Is(err, serverutil.ErrDecodeJSON):
			logger.WarnContext(ctx, err.Error())
			serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		case errors.Is(err, serverutil.ErrValidStruct):
			logger.WarnContext(ctx, "invalid struct")
			serverutil.WriteResponse(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
		default:
			logger.WarnContext(ctx, "invalid decode valid")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusBadRequest)
		}
		return nil, err
	}
	return user, nil
}

func (a *AppInstance) handleUser(ctx context.Context, w http.ResponseWriter, user *models.User, logger *slog.Logger, provider models.ProviderType) error {
	if err := a.Repos.User.HasOAuthUser(ctx, security.HashEmail(user.Email), provider); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			if err := a.Svcs.User.CreateOAuthPendingUser(ctx, user, provider); err != nil {
				switch {
				case errors.Is(err, domain.ErrInvalidValue):
					logger.WarnContext(ctx, "invalid value")
					serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
				case errors.Is(err, rp.ErrContext):
					logger.WarnContext(ctx, "context error, deadline or timeout while checking for user existence")
					serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
				case errors.Is(err, domain.ErrNotCreated):
					logger.WarnContext(ctx, "failed to create oauth user")
					serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
				case errors.Is(err, domain.ErrUnexpectedType):
					logger.WarnContext(ctx, "unexpected errror adding oauth user")
					serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
				}
				return err
			}
			if errs := a.Svcs.Mail.PendingUser(ctx, user); len(errs) > 0 {
				logger.WarnContext(ctx, "sending mail to welcome new oauth pending user")
				serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
				return errs
			}
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error, deadline or timeout while checking for user existence")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database checking for oauth user existence query failed")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, "unexpected errror checking for oauth user existence in pending_users table")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
		return err
	}
	return nil
}

func (a *AppInstance) handleSession(ctx context.Context, w http.ResponseWriter, email string, logger *slog.Logger) error {
	userID, role, err := a.Svcs.User.GetUserSessionData(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidValue):
			logger.WarnContext(ctx, "invalid value in getting oauth user session data")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrNotFound):
			logger.WarnContext(ctx, "user session data (userID, role) not found in database")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, rp.ErrContext):
			logger.WarnContext(ctx, "context error, deadline or timeout while checking for user existence")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database getting user session data query failed")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, "unexpected error getting oauth user session data")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
	}

	sessionID, err := a.Svcs.Session.CreateSession(ctx, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidValue):
			logger.WarnContext(ctx, "invalid value in session validation")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrMarshalJSON):
			logger.WarnContext(ctx, "marshal session data for oauth user")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrQueryFailed):
			logger.WarnContext(ctx, "database creating session for oauth user query failed")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		case errors.Is(err, domain.ErrUnexpectedType):
			logger.WarnContext(ctx, "unexpected error creating session for oauth user")
			serverutil.WriteResponse(w, handler.NewInternalErr(err), http.StatusInternalServerError)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionService.SessionName,
		Value:    sessionID,
		Expires:  time.Now().Add(sessionService.SessionDuration),
		HttpOnly: true,
	})
	return nil
}
