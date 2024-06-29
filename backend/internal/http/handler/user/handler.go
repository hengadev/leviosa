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
		input, err := serverutil.Decode[struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}](r)
		user, err := usr.CreateAccount(ctx, input.Email, input.Password)
		if err != nil {
			slog.ErrorContext(ctx, "failed to create account", "error", err)
			http.Error(w, handler.NewInternalErr(err), http.StatusInternalServerError)
			return
		}
		_ = user

		sessionID, err := ssn.CreateSession(ctx, user.ID)
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

func UpdateUser(svc *user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("updating the user handler")
	})
}

func GetUser(repo *user.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("get the user handler")
	})
}

func DeleteUser(svc *user.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("delete the user handler")
	})
}
