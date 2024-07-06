package http

import (
	"fmt"
	"net/http"

	ch "github.com/GaryHY/event-reservation-app/internal/http/handler/checkout"
	py "github.com/GaryHY/event-reservation-app/internal/http/handler/payment"
	uh "github.com/GaryHY/event-reservation-app/internal/http/handler/user"
	vh "github.com/GaryHY/event-reservation-app/internal/http/handler/vote"
	mw "github.com/GaryHY/event-reservation-app/internal/http/middleware"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
)

func (s *Server) addRoutes(svcs *handler.Handler) {
	mux := http.NewServeMux()
	// basic route
	mux.Handle("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world")
	}))

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
	// checkout
	handleCheckout := ch.CreateCheckoutSession(*svcs.Svcs.Checkout, svcs.Repos.Event)
	// vote
	handleGetVotesByUserID := vh.GetVotesByUserID(svcs.Svcs.Vote)

	// assign to multiplexer
	// user
	mux.Handle("GET /me", handleGetUser)
	mux.Handle("PUT /me", handleUpdateUser)
	mux.Handle("DELETE /me", handleDeleteUser)
	mux.Handle(fmt.Sprintf("POST ", serverutil.SIGNUPENDPOINT), handleSignup)
	mux.Handle(fmt.Sprintf("POST ", serverutil.SIGNINENDPOINT), handleSignin)
	// payment
	mux.Handle("POST /admin/payment", handlePostPayment)
	mux.Handle("DELETE /admin/payment", handleDeletePayment)
	// checkout
	mux.Handle("POST /checkout", handleCheckout)
	// vote
	mux.Handle("GET /vote/{month}/{year}", handleCheckout)

	s.srv.Handler = mux
}
