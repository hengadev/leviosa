package api

import (
	"github.com/GaryHY/event-reservation-app/internal/types"
	"net/http"
)

const (
	YYYYMMDD = "2006-01-02"
)

func NewServer(store Store) *Server {
	server := new(Server)
	server.Store = store

	router := http.NewServeMux()
	router.Handle("/event", http.HandlerFunc(server.eventHandler))
	router.Handle("/event/{id}", http.HandlerFunc(server.eventByIdHandler))
	// You can use the r.PathValue to get the value of id from that URL

	router.Handle("/signup", http.HandlerFunc(server.signUpHandler))
	router.Handle("/signin", http.HandlerFunc(server.signInHandler))
	router.Handle("/signout", http.HandlerFunc(server.signOutHandler))

	// router.Handle("/votes", http.HandlerFunc(server.votesHandler))

	server.Handler = router

	return server
}

type Server struct {
	Store Store
	http.Handler
}

type Store interface {
	// NOTE: pas besoin de cela
	GetEventByID(id string) types.Event
	PostEvent(event *types.Event)
	GetAllEvents() []types.Event
	CreateUser(newUser *types.User) error
	CheckUser(email string) bool
	GetHashPassword(user *types.User) (hashpassword string)
	CreateSession(newSession *types.Session) error
	HasSession(email string) bool
	DeleteSession(session *types.Session) error
	DeleteSessionByID(id string) error
	// NOTE: The next one is for auth, do I need that in the store
	// VerifyUser(user types.User) bool
}
