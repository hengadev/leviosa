package carePlanService

import (
	"time"

	"github.com/google/uuid"
)

type CarePlan struct {
	ID             string    `json:"id"`
	RegistrationID string    `json:"registration_id"`
	UserID         string    `json:"user_id"`
	Feedback       string    `json:"feedback"`  // Session feedback
	Exercices      []string  `json:"exercices"` // Recommended exercices
	Videos         []string  `json:"videos"`    // optional - stores links to the exercies recommended
	CreatedAt      time.Time `json:"created_at"`
}

func NewCarePlan(
	registrationID string,
	userID string,
	content string,
	videos []string,
) *CarePlan {
	return &CarePlan{
		ID:             uuid.NewString(),
		UserID:         userID,
		RegistrationID: registrationID,
		Videos:         videos,
	}
}
