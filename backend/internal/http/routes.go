package http

import (
	"fmt"
	"net/http"

	si "github.com/GaryHY/event-reservation-app/internal/http/handler/signin"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
	// su "github.com/GaryHY/event-reservation-app/internal/http/handler/signup"
	uh "github.com/GaryHY/event-reservation-app/internal/http/handler/user"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
)

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
	mux.Handle(fmt.Sprintf("POST ", serverutil.SIGNUPENDPOINT), handleSignup)
	// user:signin
	handleSignin := si.SigninHandler(svcs)
	mux.Handle(fmt.Sprintf("POST ", serverutil.SIGNINENDPOINT), handleSignin)

	s.srv.Handler = mux
}
