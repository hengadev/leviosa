package domain

import (
	"time"

	"github.com/google/uuid"
)

// JoinWaitlistRequest is the request body for POST /waitlist.
type JoinWaitlistRequest struct {
	ProductID         uuid.UUID  `json:"product_id" validate:"required"`
	ClientID          *uuid.UUID `json:"client_id,omitempty"`
	DesiredRangeStart time.Time  `json:"desired_range_start" validate:"required"`
	DesiredRangeEnd   time.Time  `json:"desired_range_end" validate:"required"`
	Notes             string     `json:"notes,omitempty" validate:"max=1000"`

	// Guest fields (required when client_id is omitted)
	GuestFirstName string `json:"guest_first_name,omitempty"`
	GuestLastName  string `json:"guest_last_name,omitempty"`
	GuestEmail     string `json:"guest_email,omitempty"`
	GuestPhone     string `json:"guest_phone,omitempty"`
}

// WaitlistEntryResponse is the DTO returned after joining a waitlist.
type WaitlistEntryResponse struct {
	ID                uuid.UUID           `json:"id"`
	ProductID         uuid.UUID           `json:"product_id"`
	ClientID          *uuid.UUID          `json:"client_id,omitempty"`
	DesiredRangeStart time.Time           `json:"desired_range_start"`
	DesiredRangeEnd   time.Time           `json:"desired_range_end"`
	Notes             string              `json:"notes,omitempty"`
	Status            WaitlistEntryStatus `json:"status"`
	CreatedAt         time.Time           `json:"created_at"`

	// Guest contact fields (populated for guest entries)
	GuestFirstName string `json:"guest_first_name,omitempty"`
	GuestLastName  string `json:"guest_last_name,omitempty"`
	GuestEmail     string `json:"guest_email,omitempty"`
	GuestPhone     string `json:"guest_phone,omitempty"`
}
