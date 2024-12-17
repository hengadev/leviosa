package server

import (
	"fmt"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/server/handler/checkout"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/event"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/payment"
	// "github.com/GaryHY/event-reservation-app/internal/server/handler/register"
	"github.com/GaryHY/event-reservation-app/internal/server/app"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/user"
	"github.com/GaryHY/event-reservation-app/internal/server/handler/vote"
	mw "github.com/GaryHY/event-reservation-app/internal/server/middleware"
	"github.com/GaryHY/event-reservation-app/pkg/contextutil"
	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

// make all the routes start with "/api/v1"
func (s *Server) addRoutes(h *app.App) {
	mux := http.NewServeMux()
	// basic route to test things out
	mux.Handle("GET /api/v1/hello", sayHello(h))

	// TODO: placed here just to test things out brother
	// use some sort of middleware to handle that brother

	// handler declaration
	userHandler := userHandler.New(h)
	voteHandler := vote.NewHandler(h)
	paymentHandler := payment.NewHandler(h)
	checkoutHandler := checkout.NewHandler(h)
	// registerHandler := register.NewHandler(h)
	eventHandler := event.NewHandler(h)

	// TODO: now use the auth in the different handler here
	mux.Handle("POST /api/v1/register", userHandler.Register())

	// assign to multiplexer
	// user
	mux.Handle("GET /api/v1/me", userHandler.GetUser())
	mux.Handle("PUT /api/v1/me", userHandler.UpdateUser())
	mux.Handle("DELETE /api/v1/me", userHandler.DeleteUser())
	mux.Handle(fmt.Sprintf("POST /api/v1/%s", serverutil.SIGNUPENDPOINT), userHandler.CreateAccount())
	rateLimit := mw.PerIPRateLimit(1, 1)
	mux.Handle(fmt.Sprintf("POST /api/v1/%s", serverutil.SIGNINENDPOINT), rateLimit(userHandler.Signin()))
	mux.Handle("POST /api/v1/signout", userHandler.Signout())

	mux.Handle("POST /api/v1/oauth/{provider}/user", userHandler.HandleOAuth())

	// payment
	mux.Handle("POST /api/v1/admin/payment", paymentHandler.CreateEventProduct())
	mux.Handle("DELETE /api/v1/admin/payment", paymentHandler.DeleteEventProduct())

	// vote
	mux.Handle("GET /api/v1/vote/{month}/{year}", voteHandler.GetVotesByUserID())
	// checkout
	mux.Handle("POST /api/v1/checkout/{id}/{spot}", checkoutHandler.CreateCheckoutSession())
	// register
	// mux.Handle("POST /api/v1/register", registerHandler.MakeRegistration())
	// event
	mux.Handle("GET /api/v1/event/{id}", eventHandler.FindEventByID())

	s.srv.Handler = mux
}

func sayHello(h *app.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintln(w, "here I am in the basic thing brother")
		// here use the logger to do  the things that you want to do brother
		ctx := r.Context()
		logger, err := contextutil.GetLoggerFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(ctx, "a debug message for the first use")
		logger.InfoContext(ctx, "here is the first message from the logger inside the handler ")
		fmt.Fprintln(w, "hello world!")
	})
}
