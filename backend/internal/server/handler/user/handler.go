package user

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

type Handler struct {
	*handler.Handler
}

func NewHandler(handler *handler.Handler) *Handler {
	return &Handler{handler}
}

func (h *Handler) CreateAccount() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// decode user sent and validate it.
		input, pbms, err := serverutil.DecodeValid[user.User](r)
		if len(pbms) > 0 {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		user, err := h.Svcs.User.CreateAccount(ctx, &input)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create account", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// sessionID, err := ssn.CreateSession(ctx, user.ID, user.Role)
		sessionID, err := h.Svcs.Session.CreateSession(ctx, user.ID, user.Role)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create session", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err = serverutil.Encode(w, http.StatusCreated, struct {
			SessionID string `json:"sessionid"`
		}{
			SessionID: sessionID,
		}); err != nil {
			slog.ErrorContext(ctx, "failed to encode the votes from database", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) Signin() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// parse the request body
		input, pbms, err := serverutil.DecodeValid[user.Credentials](r)
		if len(pbms) > 0 {
			slog.ErrorContext(ctx, "failed to authenticate the user, bad request", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to authenticate the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// validate credentials
		user, err := h.Repos.User.ValidateCredentials(ctx, &input)
		if user == nil {
			slog.ErrorContext(ctx, "failed to validate user credentials", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to validate user credentials", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// create session
		sessionID, err := h.Svcs.Session.CreateSession(ctx, user.ID, user.Role)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create session", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// TODO: send the session information within a cookie ? as before ? Or remove that as a chore if the cookie is well sent.
		// type Response struct {
		// 	SessionID string `json:"sessionid"`
		// }
		// serverutil.Encode(w, http.StatusCreated, Response{sessionID})
		http.SetCookie(w, &http.Cookie{
			Name:     session.SessionName,
			Value:    sessionID,
			Expires:  time.Now().Add(session.SessionDuration),
			HttpOnly: true,
		})
	})
}

func (h *Handler) Signout() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		userID := ctx.Value(mw.SessionIDKey).(string)
		// remove the session that has userID
		if err := h.Svcs.Session.RemoveSession(ctx, userID); err != nil {
			slog.ErrorContext(ctx, "failed to remove user session", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// TODO: send some redirection or something ?
	})
}

func (h *Handler) UpdateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the user from this
		user, pbms, err := serverutil.DecodeValid[user.User](r)
		if len(pbms) > 0 {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to decode user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// modify the user
		if err = h.Svcs.User.UpdateAccount(ctx, &user); err != nil {
			slog.ErrorContext(ctx, "failed to modify the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) GetUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// I need to get the userID from the sessionID that I have in the header
		userID := ctx.Value(mw.SessionIDKey).(string)
		user, err := h.Repos.User.FindAccountByID(ctx, userID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		if err := serverutil.Encode(w, http.StatusFound, user); err != nil {
			slog.ErrorContext(ctx, "failed to send the user", "error", err)
			http.Error(w, errsrv.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) DeleteUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("delete the user handler")
	})
}
