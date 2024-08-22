package server

import (
	"fmt"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler/checkout"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/event"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/payment"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/register"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/vote"
	"github.com/GaryHY/event-reservation-app/internal/server/service"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// make all the routes start with "/api/v1"
func (s *Server) addRoutes(h *handler.Handler) {
	mux := http.NewServeMux()
	// basic route to test things out
	mux.Handle("GET /api/v1/hello", http.HandlerFunc(sayHello))
	// handler declaration
	userHandler := userHandler.New(h)
	voteHandler := vote.NewHandler(h)
	paymentHandler := payment.NewHandler(h)
	checkoutHandler := checkout.NewHandler(h)
	registerHandler := register.NewHandler(h)
	eventHandler := event.NewHandler(h)

	// assign to multiplexer
	// user
	mux.Handle("GET /api/v1/me", userHandler.GetUser())
	mux.Handle("PUT /api/v1/me", userHandler.UpdateUser())
	mux.Handle("DELETE /api/v1/me", userHandler.DeleteUser())
	mux.Handle(fmt.Sprintf("POST /api/v1/%s", serverutil.SIGNUPENDPOINT), userHandler.CreateAccount())
	mux.Handle(fmt.Sprintf("POST /api/v1/%s", serverutil.SIGNINENDPOINT), userHandler.Signin())
	mux.Handle("POST /api/v1/signout", userHandler.Signout())
	// payment
	mux.Handle("POST /api/v1/admin/payment", paymentHandler.CreateEventProduct())
	mux.Handle("DELETE /api/v1/admin/payment", paymentHandler.DeleteEventProduct())
	// vote
	mux.Handle("GET /api/v1/vote/{month}/{year}", voteHandler.GetVotesByUserID())
	// checkout
	mux.Handle("POST /api/v1/checkout/{id}/{spot}", checkoutHandler.CreateCheckoutSession())
	// register
	mux.Handle("POST /api/v1/register", registerHandler.MakeRegistration())
	// event
	mux.Handle("GET /api/v1/event/{id}", eventHandler.FindEventByID())

	s.srv.Handler = mux
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World!")
}
