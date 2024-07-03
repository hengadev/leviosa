package handler

import (
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/internal/domain/user"
)

type Service struct {
	User    *user.Service
	Session *session.Service
}

// TODO: add all the repos readers.
type Repo struct {
	User    user.Reader
	Session session.Reader
}

type Services struct {
	Svc  *Service
	Repo *Repo
}
