package domain

import (
	"time"

	"github.com/google/uuid"
)

// FreeSlots is the answer to a slot query: the engine's slots plus the
// bounds the policy puts on any query, so the calendar knows where to stop.
type FreeSlots struct {
	Slots []Slot
	// EarliestStart is now + minimum notice: no slot starts before it.
	EarliestStart time.Time
	// HorizonEnd is now + booking horizon: no slot starts after it.
	HorizonEnd time.Time
}

// FreeSlotsResponse is the public slot endpoint's body.
type FreeSlotsResponse struct {
	TimeZone      string         `json:"time_zone"`
	EarliestStart time.Time      `json:"earliest_start"`
	HorizonEnd    time.Time      `json:"horizon_end"`
	Slots         []SlotResponse `json:"slots"`
}

type SlotResponse struct {
	AvailabilityID uuid.UUID `json:"availability_id"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
}

// ToResponse renders the slots for the public endpoint, times in the
// practice's time zone (the offsets show Europe/Paris).
func (f *FreeSlots) ToResponse() FreeSlotsResponse {
	slots := make([]SlotResponse, 0, len(f.Slots))
	for _, slot := range f.Slots {
		slots = append(slots, SlotResponse{
			AvailabilityID: slot.AvailabilityID,
			Start:          slot.Start.In(PracticeLocation),
			End:            slot.End.In(PracticeLocation),
		})
	}
	return FreeSlotsResponse{
		TimeZone:      PracticeLocation.String(),
		EarliestStart: f.EarliestStart.In(PracticeLocation),
		HorizonEnd:    f.HorizonEnd.In(PracticeLocation),
		Slots:         slots,
	}
}
