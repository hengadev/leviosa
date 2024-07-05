package userhandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"github.com/GaryHY/event-reservation-app/internal/http/handler"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func CreateAccount(usr *user.Service, ssn *session.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// TODO: Need to see if that works especially the part with the birth date.
		input, err := serverutil.Decode[user.User](r)
		user, err := usr.CreateAccount(ctx, &input)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create account", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		sessionID, err := ssn.CreateSession(ctx, user.ID, user.Role)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create session", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		serverutil.Encode(w, http.StatusCreated, struct {
			SessionID string `json:"sessionid"`
		}{
			SessionID: sessionID,
		})
	})
}

func Signin(usr user.Reader, ssn *session.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// parse the request body
		input, pbms, err := serverutil.DecodeValid[user.Credentials](r)
		if len(pbms) > 0 {
			slog.ErrorContext(ctx, "failed to authenticate the user, bad request", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to authenticate the user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// validate credentials
		user, err := usr.ValidateCredentials(ctx, &input)
		if user != nil {
			slog.ErrorContext(ctx, "failed to validate user credentials", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		if err != nil {
			slog.ErrorContext(ctx, "failed to validate user credentials", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// create session
		sessionID, err := ssn.CreateSession(ctx, user.ID, user.Role)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create session", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		// send session ID to user
		serverutil.Encode(w, http.StatusCreated, struct {
			SessionID string `json:"sessionid"`
		}{
			SessionID: sessionID,
		})
	})
}

func UpdateUser(svc *user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// get the user from this
		user, err := serverutil.Decode[user.User](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the user ID", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		// modify the user
		if err = svc.UpdateAccount(ctx, &user); err != nil {
			slog.ErrorContext(ctx, "failed to modify the user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func GetUser(repo user.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		// I need to get the userID from the sessionID that I have in the header
		userID, err := serverutil.Decode[struct {
			ID string `json:"userid"`
		}](r)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get the user ID", "error", err)
			http.Error(w, handler.NewBadRequestErr(err), http.StatusBadRequest)
			return
		}
		user, err := repo.FindAccountByID(ctx, userID.ID)
		if err := serverutil.Encode(w, http.StatusFound, user); err != nil {
			slog.ErrorContext(ctx, "failed to send the user", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
	})
}

func DeleteUser(svc *user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("delete the user handler")
	})
}
