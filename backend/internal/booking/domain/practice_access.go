package domain

import (
	"strings"
	"time"
)

// PracticeAccess is how a client gets into the practice: its address, the
// door code and the access instructions. The door code and the instructions
// are private (settings, encrypted): they reach only the client of a booking
// that may see them (ShowsPracticeAccess) and never appear in a log line.
type PracticeAccess struct {
	Address      string
	DoorCode     string
	Instructions string
}

// IsSet reports whether there is anything to show beyond the address: a door
// code or access instructions.
func (a PracticeAccess) IsSet() bool {
	return strings.TrimSpace(a.DoorCode) != "" || strings.TrimSpace(a.Instructions) != ""
}

// ShowsPracticeAccess is the access visibility rule: the practice access is
// shown for a booking only while it is confirmed and its slot has not ended.
func ShowsPracticeAccess(b *Booking, now time.Time) bool {
	if b == nil || b.Status != BookingStatusConfirmed {
		return false
	}
	return now.Before(b.SlotEndTime)
}

// PracticeAccessView is the practice access in the public booking lookup
// response.
type PracticeAccessView struct {
	Address      string `json:"address"`
	DoorCode     string `json:"door_code"`
	Instructions string `json:"instructions"`
}
