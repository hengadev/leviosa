package handler

import (
	// "net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

// in the grpc thing, it is our baseHandler in the grpc server struct

// TODO: add all the services.
type Service struct {
	User    *user.Service
	Session *session.Service
}

// TODO: add all the repos readers.
type Repo struct {
	User    *user.Reader
	Session *session.Reader
}

type Services struct {
	Svc  *Service
	Repo *Repo
}
