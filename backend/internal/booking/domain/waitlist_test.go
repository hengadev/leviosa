package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validWaitlistRange() (time.Time, time.Time) {
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(7 * 24 * time.Hour)
	return start, end
}

func TestNewWaitlistEntry(t *testing.T) {
	productID := uuid.New()
	clientID := uuid.New()

	t.Run("creates a registered-client entry in the waiting state", func(t *testing.T) {
		start, end := validWaitlistRange()
		entry, err := NewWaitlistEntry(productID, &clientID, start, end, "any morning works", "", "", "", "")
		require.NoError(t, err)
		require.NotNil(t, entry)

		assert.NotEqual(t, uuid.Nil, entry.ID)
		assert.Equal(t, productID, entry.ProductID)
		assert.Equal(t, &clientID, entry.ClientID)
		assert.Equal(t, WaitlistEntryStatusWaiting, entry.Status)
		assert.False(t, entry.IsGuestEntry())
		assert.True(t, entry.IsActive())
	})

	t.Run("creates a guest entry when client ID is nil and guest fields are provided", func(t *testing.T) {
		start, end := validWaitlistRange()
		entry, err := NewWaitlistEntry(productID, nil, start, end, "", "Jane", "Doe", "jane@example.com", "")
		require.NoError(t, err)
		require.NotNil(t, entry)

		assert.Nil(t, entry.ClientID)
		assert.True(t, entry.IsGuestEntry())
		assert.Equal(t, "Jane Doe", entry.GuestDisplayName())
		assert.Equal(t, WaitlistEntryStatusWaiting, entry.Status)
	})

	t.Run("guest entry accepts phone only, without email", func(t *testing.T) {
		start, end := validWaitlistRange()
		entry, err := NewWaitlistEntry(productID, nil, start, end, "", "Jane", "Doe", "", "+15551234567")
		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.True(t, entry.IsGuestEntry())
	})

	t.Run("rejects a nil product ID", func(t *testing.T) {
		start, end := validWaitlistRange()
		_, err := NewWaitlistEntry(uuid.Nil, &clientID, start, end, "", "", "", "", "")
		assert.ErrorIs(t, err, ErrInvalidWaitlistProductID)
	})

	t.Run("rejects a range where end is not after start", func(t *testing.T) {
		start := time.Now().Add(24 * time.Hour)
		_, err := NewWaitlistEntry(productID, &clientID, start, start, "", "", "", "", "")
		assert.ErrorIs(t, err, ErrInvalidWaitlistDateRange)

		_, err = NewWaitlistEntry(productID, &clientID, start, start.Add(-time.Hour), "", "", "", "", "")
		assert.ErrorIs(t, err, ErrInvalidWaitlistDateRange)
	})

	t.Run("rejects a range that has already fully elapsed", func(t *testing.T) {
		start := time.Now().Add(-48 * time.Hour)
		end := time.Now().Add(-24 * time.Hour)
		_, err := NewWaitlistEntry(productID, &clientID, start, end, "", "", "", "", "")
		assert.ErrorIs(t, err, ErrCannotJoinWaitlistInPast)
	})

	t.Run("rejects when neither client ID nor guest fields are provided", func(t *testing.T) {
		start, end := validWaitlistRange()
		_, err := NewWaitlistEntry(productID, nil, start, end, "", "", "", "", "")
		assert.ErrorIs(t, err, ErrInvalidWaitlistIdentity)
	})

	t.Run("rejects incomplete guest fields (missing contact info)", func(t *testing.T) {
		start, end := validWaitlistRange()
		_, err := NewWaitlistEntry(productID, nil, start, end, "", "Jane", "Doe", "", "")
		assert.ErrorIs(t, err, ErrInvalidWaitlistIdentity)
	})

	t.Run("rejects both a client ID and guest fields at once", func(t *testing.T) {
		start, end := validWaitlistRange()
		_, err := NewWaitlistEntry(productID, &clientID, start, end, "", "Jane", "Doe", "jane@example.com", "")
		assert.ErrorIs(t, err, ErrAmbiguousWaitlistIdentity)
	})
}

// TestWaitlistEntryStateTransitions exercises the full state machine described
// in the waitlist PRD: waiting -> notified -> claimed | expired, and
// waiting|notified -> cancelled. Only the `waiting` state is reachable through
// the join-waitlist flow shipped in this issue; the rest are exercised here at
// the domain level so the entity's invariants are proven even though nothing
// drives these transitions yet (that's issue 015).
func TestWaitlistEntryStateTransitions(t *testing.T) {
	newEntry := func(t *testing.T) *WaitlistEntry {
		t.Helper()
		start, end := validWaitlistRange()
		clientID := uuid.New()
		entry, err := NewWaitlistEntry(uuid.New(), &clientID, start, end, "", "", "", "", "")
		require.NoError(t, err)
		return entry
	}

	t.Run("waiting -> notified succeeds and stamps NotifiedAt", func(t *testing.T) {
		entry := newEntry(t)
		err := entry.MarkNotified()
		require.NoError(t, err)
		assert.Equal(t, WaitlistEntryStatusNotified, entry.Status)
		require.NotNil(t, entry.NotifiedAt)
	})

	t.Run("notified -> claimed succeeds and stamps ClaimedAt", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		err := entry.Claim()
		require.NoError(t, err)
		assert.Equal(t, WaitlistEntryStatusClaimed, entry.Status)
		require.NotNil(t, entry.ClaimedAt)
		assert.False(t, entry.IsActive())
	})

	t.Run("notified -> expired succeeds and stamps ExpiredAt", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		err := entry.Expire()
		require.NoError(t, err)
		assert.Equal(t, WaitlistEntryStatusExpired, entry.Status)
		require.NotNil(t, entry.ExpiredAt)
		assert.False(t, entry.IsActive())
	})

	t.Run("waiting -> cancelled succeeds and stamps CancelledAt", func(t *testing.T) {
		entry := newEntry(t)
		err := entry.Cancel()
		require.NoError(t, err)
		assert.Equal(t, WaitlistEntryStatusCancelled, entry.Status)
		require.NotNil(t, entry.CancelledAt)
		assert.False(t, entry.IsActive())
	})

	t.Run("notified -> cancelled succeeds", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		err := entry.Cancel()
		require.NoError(t, err)
		assert.Equal(t, WaitlistEntryStatusCancelled, entry.Status)
	})

	t.Run("cannot mark an already-notified entry as notified again", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		err := entry.MarkNotified()
		assert.ErrorIs(t, err, ErrWaitlistEntryNotWaiting)
	})

	t.Run("cannot claim directly from waiting (must be notified first)", func(t *testing.T) {
		entry := newEntry(t)
		err := entry.Claim()
		assert.ErrorIs(t, err, ErrWaitlistEntryNotNotified)
	})

	t.Run("cannot expire directly from waiting", func(t *testing.T) {
		entry := newEntry(t)
		err := entry.Expire()
		assert.ErrorIs(t, err, ErrWaitlistEntryNotNotified)
	})

	t.Run("cannot claim an already-claimed entry", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		require.NoError(t, entry.Claim())
		err := entry.Claim()
		assert.ErrorIs(t, err, ErrWaitlistEntryNotNotified)
	})

	t.Run("cannot cancel an already-cancelled entry", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.Cancel())
		err := entry.Cancel()
		assert.ErrorIs(t, err, ErrWaitlistEntryAlreadyCancelled)
	})

	t.Run("cannot cancel a claimed entry", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		require.NoError(t, entry.Claim())
		err := entry.Cancel()
		assert.ErrorIs(t, err, ErrCannotCancelWaitlistEntry)
	})

	t.Run("cannot cancel an expired entry", func(t *testing.T) {
		entry := newEntry(t)
		require.NoError(t, entry.MarkNotified())
		require.NoError(t, entry.Expire())
		err := entry.Cancel()
		assert.ErrorIs(t, err, ErrCannotCancelWaitlistEntry)
	})
}

func TestWaitlistEntry_GuestDisplayName(t *testing.T) {
	t.Run("returns empty string for a registered-client entry", func(t *testing.T) {
		clientID := uuid.New()
		start, end := validWaitlistRange()
		entry, err := NewWaitlistEntry(uuid.New(), &clientID, start, end, "", "", "", "", "")
		require.NoError(t, err)
		assert.Equal(t, "", entry.GuestDisplayName())
	})
}
