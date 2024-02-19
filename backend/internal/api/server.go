package api

import (
	"encoding/json"
	"github.com/GaryHY/event-reservation-app/internal/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
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
	// router.Handle("/signout", http.HandlerFunc(server.signOutHandler))
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
	GetHashPassword(user types.User) (hashpassword string)
	CreateSession(session_id string, newSession *types.Session) error
	HasSession(user types.User) bool
	PostEvent(event *types.Event)
	GetAllEvents() []types.Event
	DeleteSession(session *types.Session) error
	// NOTE: The next one is for auth
	// VerifyUser(user types.User) bool
}

func (s *Server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(w, r)
	if !user.ValidateEmail() || !user.ValidatePassword() {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if s.Store.CheckUser(user) {
		w.WriteHeader(http.StatusConflict)
		// TODO: Send in response body some message about the user already registered
		return
	}
	if err := s.Store.CreateUser(user); err == nil {
		w.WriteHeader(http.StatusCreated)
		return
	}
	return
}

func (s *Server) signInHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(w, r)
	if !s.Store.CheckUser(user) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// TODO: Add the check is the user already has a session
	if s.Store.HasSession(user) {
		w.WriteHeader(http.StatusOK)
		return
	}

	hashpassword := s.Store.GetHashPassword(user)
	if err := bcrypt.CompareHashAndPassword([]byte(hashpassword), []byte(user.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	session := types.NewSession(user)
	session_id := uuid.NewString()
	if err := s.Store.CreateSession(session_id, session); err != nil {
		log.Fatal("Failed to create session in the database for the user")
	}

	loc, _ := time.LoadLocation("Europe/Paris")
	createdTime, err := time.ParseInLocation(time.RFC822, session.Created_at, loc)
	expired_at := createdTime.Add(session.Expiry)
	if err != nil {
		log.Fatal("Cannot parse the time - ", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    types.SessionCookieName,
		Value:   session_id,
		Expires: expired_at,
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) signOutHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromRequest(w, r)
	_ = user
	// 1. get the cookie from the request to find the session id
	// 2. remove the line in the sessions table
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

func getUserFromRequest(w http.ResponseWriter, r *http.Request) types.User {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return types.User{}
	}
	return user
}

func (s *Server) AuthorizeUser() bool {
	// TODO: Fonction a implementer pour acceder a certains services
	return true
}
