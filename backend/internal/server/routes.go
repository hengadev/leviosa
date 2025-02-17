package server

import (
	"fmt"
	"net/http"

	"github.com/GaryHY/leviosa/internal/server/app"
	"github.com/GaryHY/leviosa/internal/server/handler/event"
	"github.com/GaryHY/leviosa/internal/server/handler/health"
	"github.com/GaryHY/leviosa/internal/server/handler/product"
	"github.com/GaryHY/leviosa/internal/server/handler/user"
	"github.com/GaryHY/leviosa/internal/server/handler/vote"
	mw "github.com/GaryHY/leviosa/internal/server/middleware"
)

func (s *Server) addRoutes(h *app.App) {
	router := http.NewServeMux()

	// basic route to test things out
	router.HandleFunc("GET /api/v1/hello", sayHello)

	// handlers declaration
	healthHandler := healthHandler.New(h)
	userHandler := userHandler.New(h)
	voteHandler := vote.NewHandler(h)
	eventHandler := eventHandler.New(h)
	productHandler := productHandler.New(h)

	// middlewares declaration
	rateLimit := mw.PerIPRateLimit(1, 1)

	router.HandleFunc("GET /health", healthHandler.CheckHealth)

	// user
	router.HandleFunc("GET /api/v1/user/me", userHandler.GetUser)
	router.HandleFunc("PUT /api/v1/user/me", userHandler.UpdateUser)
	router.HandleFunc("DELETE /api/v1/user/me", userHandler.DeleteUser)

	router.HandleFunc("POST /api/v1/user/signin", rateLimit(userHandler.Signin))

	router.HandleFunc("POST /api/v1/user/register", rateLimit(userHandler.RegisterUserOTP))
	router.HandleFunc("POST /api/v1/user/validate-otp", rateLimit(userHandler.ValidateUserOTP))
	router.HandleFunc("GET /api/v1/user/approve-user", rateLimit(userHandler.GetUsersToApprove))
	router.HandleFunc("POST /api/v1/user/approve-user", rateLimit(userHandler.ApproveUserRegistration))

	router.HandleFunc("POST /api/v1/user/signout", userHandler.Signout)

	router.HandleFunc("POST /api/v1/oauth/{provider}", userHandler.HandleOAuth)

	// vote
	router.HandleFunc("GET /api/v1/vote/{month}/{year}", voteHandler.GetVotesByUserID)

	// register
	// NOTE: the old way to do the reservation thing
	// mux.Handle("POST /api/v1/register", registerHandler.MakeRegistration())
	// TODO: the better way to do the reservation thing
	// mux.Handle("POST /api/v1/register/event", registerHandler.MakeRegistration())
	// mux.Handle("POST /api/v1/register/consultation", registerHandler.MakeRegistration())

	// products
	router.HandleFunc("POST /api/v1/products", productHandler.CreateProduct)
	router.HandleFunc("GET /api/v1/products/{id}", productHandler.GetProduct)
	router.HandleFunc("DELETE /api/v1/products/{id}", productHandler.DeleteProduct)
	router.HandleFunc("PUT /api/v1/products/{id}", productHandler.UpdateProduct)

	// product types
	router.HandleFunc("POST /api/v1/product-types", productHandler.CreateOffer)
	router.HandleFunc("DELETE /api/v1/product-types/{id}", productHandler.DeleteOffer)

	// event
	router.HandleFunc("GET /api/v1/events/{id}", eventHandler.FindEventByID)
	router.HandleFunc("POST /api/v1/events", eventHandler.CreateEvent)
	router.HandleFunc("PUT /api/v1/events/{id}", eventHandler.ModifyEvent)
	router.HandleFunc("DELETE /api/v1/events/{id}", eventHandler.FindEventByID)
	router.HandleFunc("GET /api/v1/events/users", eventHandler.FindEventsForUser)

	router.HandleFunc("POST /api/v1/upload-image", handleImage)

	s.srv.Handler = router
}

// TODO: how can I make groups for that thing and make sure that I can add as much middleware to a group as I want ?

func handleImage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here we are in the handle image handler")
	err := r.ParseMultipartForm(10 << 20) // Limit upload size to 10MB
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to retrieve image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Println("the filename that I uploaded is:", handler.Filename)

	w.WriteHeader(http.StatusOK)
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("here on the server, someone hit the hello endpoint\n\n")
	fmt.Fprintln(w, "hello world!")
}
