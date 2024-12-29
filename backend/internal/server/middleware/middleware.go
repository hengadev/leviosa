package middleware

import (
	"context"
	"net/http"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
)

type Handlerfunc func(http.ResponseWriter, *http.Request)

type Middleware func(Handlerfunc) Handlerfunc

type sessionGetterFunc func(ctx context.Context, sessionID string) (*sessionService.Session, error)
