package http

import (
	"fmt"
	"net/http"

	si "github.com/GaryHY/event-reservation-app/internal/http/handler/signin"
	// su "github.com/GaryHY/event-reservation-app/internal/http/handler/signup"
	uh "github.com/GaryHY/event-reservation-app/internal/http/handler/user"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
)

// how to get the handler out of this

// func addRoutes(
// 	mux *http.ServeMux,
// 	svcs *handler.Services,
// 	// logger  *logging.Logger
// ) http.Handler {
// 	// middleware for all the routes : add the dependency for the auth mw (session repository etc...)
// 	var handler http.Handler = mux
// 	handler = mw.Use(mw.RequestID, mw.RequestID)
// 	// handler declaration
// 	// user
// 	handleGetUser := uh.GetUser(svcs.Repo.User)
// 	handleUpdateUser := uh.UpdateUser(svcs.Svc.User)
// 	handleDeleteUser := uh.DeleteUser(svcs.Svc.User)
// 	mux.Handle("GET /me", handleGetUser)
// 	mux.Handle("PUT /me", handleUpdateUser)
// 	mux.Handle("DELETE /me", handleDeleteUser)
// 	// user:signup
// 	handleSignup := uh.CreateAccount(svcs.Svc.User, svcs.Svc.Session)
// 	mux.Handle("POST /signup", handleSignup)
// 	//signin
// 	handleSignin := si.SigninHandler(svcs)
// 	mux.Handle("POST /signin", handleSignin)
// 	return handler
// }

func (s *Server) addRoutes(svcs *handler.Services) {
	mux := http.NewServeMux()
	// basic route
	mux.Handle("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))

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
	// user:signin
	handleSignin := si.SigninHandler(svcs)
	mux.Handle("POST /signin", handleSignin)

	s.srv.Handler = mux
}
