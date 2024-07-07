package http

import (
	"fmt"
	"net/http"

	py "github.com/GaryHY/event-reservation-app/internal/http/handler/payment"
	reg "github.com/GaryHY/event-reservation-app/internal/http/handler/register"
	uh "github.com/GaryHY/event-reservation-app/internal/http/handler/user"
	vh "github.com/GaryHY/event-reservation-app/internal/http/handler/vote"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func (s *Server) addRoutes(svcs *handler.Handler) {
	mux := http.NewServeMux()
	// basic route
	mux.Handle("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))
	// TODO: Do I need that ? Plus I have the same thing with EnableHeaders
	// something := mw.EnableMethods(
	// 	uh.GetUser(svcs.Repos.User),
	// 	http.MethodGet,
	// 	http.MethodPost,
	// )

	// handler declaration
	// user
	handleGetUser := uh.GetUser(svcs.Repos.User)
	handleUpdateUser := uh.UpdateUser(svcs.Svcs.User)
	handleDeleteUser := uh.DeleteUser(svcs.Svcs.User)
	handleSignup := uh.CreateAccount(svcs.Svcs.User, svcs.Svcs.Session)
	handleSignin := uh.Signin(svcs.Repos.User, svcs.Svcs.Session)
	// payment
	handlePostPayment := py.CreateEventProduct(svcs.Svcs.Payment, svcs.Svcs.Event)
	handleDeletePayment := py.DeleteEventProduct(svcs.Svcs.Payment, svcs.Svcs.Event)
	// vote
	handleGetVotesByUserID := vh.GetVotesByUserID(svcs.Svcs.Vote)
	// register
	handlePostRegistration := reg.MakeRegistration(svcs.Svcs.Register, svcs.Svcs.Event, *svcs.Svcs.Checkout)

	// assign to multiplexer
	// user
	mux.Handle("GET /me", handleGetUser)
	mux.Handle("PUT /me", handleUpdateUser)
	mux.Handle("DELETE /me", handleDeleteUser)
	mux.Handle(fmt.Sprintf("POST %s", serverutil.SIGNUPENDPOINT), handleSignup)
	mux.Handle(fmt.Sprintf("POST %s", serverutil.SIGNINENDPOINT), handleSignin)
	// payment
	mux.Handle("POST /admin/payment", handlePostPayment)
	mux.Handle("DELETE /admin/payment", handleDeletePayment)
	// vote
	mux.Handle("GET /vote/{month}/{year}", handleGetVotesByUserID)
	// register
	mux.Handle("GET /register/{id}", handlePostRegistration)

	s.srv.Handler = mux
}
