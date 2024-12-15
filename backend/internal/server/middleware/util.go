package middleware

import (
	"net/http"
	"strings"

	app "github.com/GaryHY/event-reservation-app/internal/domain"
	"github.com/google/uuid"
)

func getSessionIDFromRequest(r *http.Request) (string, error) {
	sessionID := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if sessionID == "" {
		return "", app.NewSessionNotFoundErr(nil)
	}
	if err := uuid.Validate(sessionID); err != nil {
		return "", app.NewInvalidSessionErr(err)
	}
	return sessionID, nil
}
