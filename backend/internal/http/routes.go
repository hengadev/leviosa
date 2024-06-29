package http

import (
	"net/http"

	si "github.com/GaryHY/event-reservation-app/internal/http/handler/signin"
	su "github.com/GaryHY/event-reservation-app/internal/http/handler/signup"
	uh "github.com/GaryHY/event-reservation-app/internal/http/handler/user"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
)

// how to get the handler out of this

func addRoutes(
	mux *http.ServeMux,
	svcs *handler.Services,
	// logger  *logging.Logger
) http.Handler {

	var handler http.Handler = mux

	// middleware for all the routes : add the dependency for the auth mw (session repository etc...)
	// handler = mw.NewMiddleware(mw.Auth(), mw.Auth())
	// handler declaration
	// user
	handleGetUser := uh.GetUser(svcs.Repo.User)
	handleUpdateUser := uh.UpdateUser(svcs.Svc.User)
	handleDeleteUser := uh.DeleteUser(svcs.Svc.User)
	mux.Handle("GET /me", handleGetUser)
	mux.Handle("PUT /me", handleUpdateUser)
	mux.Handle("DELETE /me", handleDeleteUser)
	// user:signup
	handleSignup := uh.CreateAccount(svcs.Svc.User, svcs.Svc.Session)
	mux.Handle("POST /signup", handleSignup)
	//signin
	handleSignin := si.SigninHandler(svcs)
	mux.Handle("POST /signin", handleSignin)

	return handler
}
