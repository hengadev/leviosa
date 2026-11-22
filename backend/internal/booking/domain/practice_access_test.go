package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShowsPracticeAccess(t *testing.T) {
	now := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

	upcoming := func(t *testing.T, status BookingStatus) *Booking {
		b := bookingIn(t, status)
		b.SlotStartTime = now.Add(24 * time.Hour)
		b.SlotEndTime = now.Add(25 * time.Hour)
		return b
	}

	t.Run("shows for a confirmed booking whose slot has not started", func(t *testing.T) {
		assert.True(t, ShowsPracticeAccess(upcoming(t, BookingStatusConfirmed), now))
	})

	t.Run("shows for a confirmed booking whose slot is under way", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		b.SlotStartTime = now.Add(-30 * time.Minute)
		b.SlotEndTime = now.Add(30 * time.Minute)
		assert.True(t, ShowsPracticeAccess(b, now))
	})

	t.Run("hidden for a confirmed booking whose slot has ended", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		b.SlotStartTime = now.Add(-2 * time.Hour)
		b.SlotEndTime = now.Add(-time.Hour)
		assert.False(t, ShowsPracticeAccess(b, now))
	})

	t.Run("hidden at the very end of the slot", func(t *testing.T) {
		b := bookingIn(t, BookingStatusConfirmed)
		b.SlotStartTime = now.Add(-time.Hour)
		b.SlotEndTime = now
		assert.False(t, ShowsPracticeAccess(b, now))
	})

	for _, status := range []BookingStatus{
		BookingStatusAwaitingPayment,
		BookingStatusExpired,
		BookingStatusCancelled,
		BookingStatusCompleted,
		BookingStatusNoShow,
	} {
		t.Run("hidden for a "+string(status)+" booking", func(t *testing.T) {
			assert.False(t, ShowsPracticeAccess(upcoming(t, status), now))
		})
	}

	t.Run("hidden for no booking", func(t *testing.T) {
		assert.False(t, ShowsPracticeAccess(nil, now))
	})
}

func TestPracticeAccess_IsSet(t *testing.T) {
	assert.False(t, PracticeAccess{Address: "1 rue de la Paix, 75002 Paris"}.IsSet())
	assert.False(t, PracticeAccess{DoorCode: "  ", Instructions: "\n"}.IsSet())
	assert.True(t, PracticeAccess{DoorCode: "4821B"}.IsSet())
	assert.True(t, PracticeAccess{Instructions: "2e étage, porte gauche"}.IsSet())
}
