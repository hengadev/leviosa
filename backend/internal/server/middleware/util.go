package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hengadev/leviosa/internal/domain"
)

func getSessionIDFromRequest(r *http.Request) (string, error) {
	sessionID := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if strings.TrimSpace(sessionID) == "" {
		return "", domain.NewInvalidValueErr("missing or empty session ID in Authorization header")
	}
	if err := uuid.Validate(sessionID); err != nil {
		return "", domain.NewInvalidValueErr("must be a valid UUID")
	}
	return sessionID, nil
}
