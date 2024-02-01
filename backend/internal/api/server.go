package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func NewServer(store Store) *Server {
	server := new(Server)
	server.Store = store

	router := http.NewServeMux()
	router.Handle("/event", http.HandlerFunc(server.eventHandler))
	// NOTE: Maybe change the auth name to signin or signup:
	// TODO: Add the routes for the sign-in and the sign-up
	router.Handle("/signup", http.HandlerFunc(server.signUpHandler))
	router.Handle("/signin", http.HandlerFunc(server.signInHandler))
	server.Handler = router

	return server
}

type Server struct {
	Store Store
	http.Handler
}

type Store interface {
	GetEventNameByID(id string) types.Event
	PostEvent(name string)
	CreateUser(newUser types.User) error
	CheckUser(user types.User) bool
	// GetUser(user types.User)
	// VerifyUser(user types.User) bool
}

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. get email and password from the request that sends in a json format in the body
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// TODO: Find the right to status to send back, the idea is that the data sent is malformatted
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 1.5. check if the user mail is not already used in the database
	// 2. validate email and password
	if !s.Store.CheckUser(user) && user.ValidateEmail() && user.ValidatePassword() {
		// 3. hash with bcrypt the password and register both email and hashpassword to the database
		bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
		if err != nil {
			log.Fatal("Failed hashing the password - ", err)
		}
		hashpassword := string(bytes)
		if err := s.Store.CreateUser(types.User{Email: user.Email, Password: hashpassword}); err != nil {
			w.WriteHeader(http.StatusConflict)
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) signInHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.showEventBydID(w, r)
	case http.MethodPost:
		s.makeEvent(r)
	}
}

func (s *Server) showEventBydID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	event := s.Store.GetEventNameByID(id)
	if event.Name == "" {
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(&event)
}

func (s *Server) makeEvent(r *http.Request) {
	name := r.URL.Query().Get("name")
	s.Store.PostEvent(name)
}

func (s *Server) authorize() bool {
	// TODO: Fonction a implementer pour acceder a certains services
	return true
}
