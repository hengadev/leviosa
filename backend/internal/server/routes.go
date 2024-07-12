package server

import (
	"fmt"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/http/handler/checkout"
	"github.com/GaryHY/event-reservation-app/internal/http/handler/payment"
	"github.com/GaryHY/event-reservation-app/internal/http/handler/register"
	"github.com/GaryHY/event-reservation-app/internal/http/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/http/handler/vote"
	"github.com/GaryHY/event-reservation-app/internal/http/service"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// make all the routes start with "/api/v1"
func (s *Server) addRoutes(h *handler.Handler) {
	mux := http.NewServeMux()
	// basic route to test things out
	mux.Handle("/hello", http.HandlerFunc(sayHello))
	// handler declaration
	userHandler := user.NewHandler(h)
	voteHandler := vote.NewHandler(h)
	paymentHandler := payment.NewHandler(h)
	checkoutHandler := checkout.NewHandler(h)
	registerHandler := register.NewHandler(h)

	// assign to multiplexer
	// user
	mux.Handle("GET /me", userHandler.GetUser())
	mux.Handle("PUT /me", userHandler.UpdateUser())
	mux.Handle("DELETE /me", userHandler.DeleteUser())
	mux.Handle(fmt.Sprintf("POST %s", serverutil.SIGNUPENDPOINT), userHandler.CreateAccount())
	mux.Handle(fmt.Sprintf("POST %s", serverutil.SIGNINENDPOINT), userHandler.Signin())
	mux.Handle("POST /signout", userHandler.Signout())
	// payment
	mux.Handle("POST /admin/payment", paymentHandler.CreateEventProduct())
	mux.Handle("DELETE /admin/payment", paymentHandler.DeleteEventProduct())
	// vote
	mux.Handle("GET /vote/{month}/{year}", voteHandler.GetVotesByUserID())
	// checkout
	mux.Handle("POST /checkout/{id}/{spot}", checkoutHandler.CreateCheckoutSession())
	// register
	mux.Handle("POST /register", registerHandler.MakeRegistration())

	s.srv.Handler = mux
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World!")
}
