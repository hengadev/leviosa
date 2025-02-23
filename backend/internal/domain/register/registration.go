package registerService

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Registration struct {
	ID                 string           `json:"id"`
	UserID             string           `json:"user_id"`
	ProductID          string           `json:"product_id,omitempty"`
	Type               RegistrationType `json:"type"`
	StartTime          time.Time        `json:"start_time"`
	StartTimeFormatted string           `json:"-"`
	EndTime            time.Time        `json:"end_time"`
	EndTimeFormatted   string           `json:"-"`
	Duration           time.Duration    `json:"duration"`
	Location           *string          `json:"location,omitempty"` // Address for at home consultation
	Notes              *string          `json:"notes,omitempty"`    // Add notes from the user
	CreatedAt          time.Time        `json:"created_at"`
	CreatedAtFormatted string           `json:"-"`
	UpdatedAt          time.Time        `json:"updated_at_formart"`
	UpdatedAtFormatted string           `json:"-"`
}

func NewRegistration(
	userID string,
	productID string,
	regType RegistrationType,
	startTime time.Time,
	endTime time.Time,
	location *string,
	notes *string,
) *Registration {
	return &Registration{
		ID:        uuid.NewString(),
		UserID:    userID,
		ProductID: productID,
		Type:      regType,
		StartTime: startTime,
		EndTime:   endTime,
		Location:  location,
		Notes:     notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (r *Registration) Update(ctx context.Context) error {
	r.UpdatedAt = time.Now()
	return nil
}
