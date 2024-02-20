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
	// for admins
	router.Handle("/admin/event", http.HandlerFunc(server.adminEventHandler))

	// for the users
	// router.Handle("/event", http.HandlerFunc(server.eventHandler)) // avec un get avec le query string et un get sans pour prendre tous les events d'un user

	// TODO: DO that one too
	// router.Handle("/event/{user_id}", http.HandlerFunc(server.eventByIdHandler))

	router.Handle("/signup", http.HandlerFunc(server.signUpHandler))
	router.Handle("/signin", http.HandlerFunc(server.signInHandler))
	router.Handle("/signout", http.HandlerFunc(server.signOutHandler))

	router.Handle("/votes", http.HandlerFunc(server.votesHandler))

	server.Handler = router

	return server
}

type Server struct {
	Store Store
	http.Handler
}

type Store interface {
	GetEventByID(id string) *types.Event
	GetAllEvents() []types.Event
	PostEvent(event *types.Event)
	DeleteEvent(event_id string) error
	GetUserId(user_email string) string
	GetUserIdBySessionId(session_id string) string
	CreateUser(newUser *types.UserStored) error
	CheckUser(email string) bool
	IsAdmin(session_id string) bool
	GetHashPassword(user *types.User) (hashpassword string)
	CreateSession(newSession *types.Session) error
	HasSession(email string) bool
	DeleteSession(session *types.Session) error
	DeleteSessionByID(id string) error
	CreateVote(newVote *types.Vote) error
	DecreaseEventPlacecount(event_id string) error
	CheckEvent(event_id string) bool
	CheckVote(userId, eventId string) bool
	// TODO: Function to implement
	Authorize(session_id string, role types.Role) bool
}
