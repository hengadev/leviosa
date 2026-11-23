package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clientBooking(id uuid.UUID, clientID *uuid.UUID, name, phone, email string, start time.Time, status BookingStatus) PartnerBookingResponse {
	return PartnerBookingResponse{
		ID:            id,
		ClientID:      clientID,
		ClientName:    name,
		ClientPhone:   phone,
		ClientEmail:   email,
		ProductName:   "Consultation",
		SlotStartTime: start,
		SlotEndTime:   start.Add(time.Hour),
		Status:        status,
	}
}

func TestClientGroupKey(t *testing.T) {
	clientID := uuid.New()

	t.Run("registered client groups by account", func(t *testing.T) {
		a := clientBooking(uuid.New(), &clientID, "A", "+33612345678", "a@example.com", time.Now(), BookingStatusConfirmed)
		b := clientBooking(uuid.New(), &clientID, "A", "", "other@example.com", time.Now(), BookingStatusCompleted)
		assert.Equal(t, ClientGroupKey(a), ClientGroupKey(b))
		assert.Equal(t, "client:"+clientID.String(), ClientGroupKey(a))
	})

	t.Run("guest phones in different formats group together", func(t *testing.T) {
		a := clientBooking(uuid.New(), nil, "G", "06 12 34 56 78", "", time.Now(), BookingStatusConfirmed)
		b := clientBooking(uuid.New(), nil, "G", "+33612345678", "ignored@example.com", time.Now(), BookingStatusCompleted)
		c := clientBooking(uuid.New(), nil, "G", "06.12-34.56.78", "", time.Now(), BookingStatusNoShow)
		assert.Equal(t, ClientGroupKey(a), ClientGroupKey(b))
		assert.Equal(t, ClientGroupKey(a), ClientGroupKey(c))
		assert.Equal(t, "phone:+33612345678", ClientGroupKey(a))
	})

	t.Run("email when no phone, lower-cased", func(t *testing.T) {
		a := clientBooking(uuid.New(), nil, "G", "", "Guest@Example.com", time.Now(), BookingStatusConfirmed)
		assert.Equal(t, "email:guest@example.com", ClientGroupKey(a))
	})

	t.Run("a guest with the same phone as a client stays separate", func(t *testing.T) {
		guest := clientBooking(uuid.New(), nil, "G", "+33612345678", "", time.Now(), BookingStatusConfirmed)
		assert.Equal(t, "phone:+33612345678", ClientGroupKey(guest))
	})
}

func TestGroupClientBookings(t *testing.T) {
	clientID := uuid.New()
	now := time.Date(2025, 6, 10, 12, 0, 0, 0, time.UTC)

	t.Run("excludes awaiting_payment and expired, keeps cancelled", func(t *testing.T) {
		bookings := []PartnerBookingResponse{
			clientBooking(uuid.New(), nil, "G", "0612345678", "", now, BookingStatusAwaitingPayment),
			clientBooking(uuid.New(), nil, "G", "0612345678", "", now, BookingStatusExpired),
			clientBooking(uuid.New(), nil, "G", "0612345678", "", now, BookingStatusCancelled),
		}
		groups := GroupClientBookings(bookings)
		require.Len(t, groups, 1)
		require.Len(t, groups[0].Bookings, 1)
		assert.Equal(t, BookingStatusCancelled, groups[0].Bookings[0].Status)
	})

	t.Run("distinct clients, contact filled from bookings", func(t *testing.T) {
		bookings := []PartnerBookingResponse{
			clientBooking(uuid.New(), &clientID, "Compte", "+33611111111", "compte@example.com", now, BookingStatusConfirmed),
			clientBooking(uuid.New(), nil, "Invité", "+33622222222", "invite@example.com", now, BookingStatusConfirmed),
			clientBooking(uuid.New(), nil, "Email", "", "EMAIL@example.com", now, BookingStatusConfirmed),
		}
		groups := GroupClientBookings(bookings)
		require.Len(t, groups, 3)
		assert.Equal(t, "Invité", groups[1].Name)
		assert.Equal(t, "+33622222222", groups[1].Phone)
	})

	t.Run("BuildClientFile splits and orders upcoming and past", func(t *testing.T) {
		g := ClientGroup{
			Key:   "phone:+33612345678",
			Name:  "G",
			Phone: "+33612345678",
			Bookings: []PartnerBookingResponse{
				clientBooking(uuid.New(), nil, "G", "+33612345678", "", now.Add(48*time.Hour), BookingStatusConfirmed),
				clientBooking(uuid.New(), nil, "G", "+33612345678", "", now.Add(-24*time.Hour), BookingStatusCompleted),
				clientBooking(uuid.New(), nil, "G", "+33612345678", "", now.Add(-48*time.Hour), BookingStatusCancelled),
				clientBooking(uuid.New(), nil, "G", "+33612345678", "", now.Add(24*time.Hour), BookingStatusConfirmed),
				// A session in progress counts as upcoming
				clientBooking(uuid.New(), nil, "G", "+33612345678", "", now.Add(-10*time.Minute), BookingStatusConfirmed),
			},
		}
		file := BuildClientFile(g, now)
		require.Len(t, file.Upcoming, 3)
		require.Len(t, file.Past, 2)
		// Soonest first
		assert.True(t, file.Upcoming[0].SlotStartTime.Before(file.Upcoming[1].SlotStartTime))
		assert.True(t, file.Upcoming[1].SlotStartTime.Before(file.Upcoming[2].SlotStartTime))
		// Latest first
		assert.True(t, file.Past[0].SlotStartTime.After(file.Past[1].SlotStartTime))
		assert.Equal(t, BookingStatusCancelled, file.Past[1].Status)
	})
}

func TestSearchClientSummaries(t *testing.T) {
	accountID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	groups := []ClientGroup{
		{Key: "client:a", ClientID: &accountID, Name: "Jean Dupont", Phone: "+33612345678", Email: "jean@example.com"},
		{Key: "client:b", Name: "Marie Martin", Phone: "+33798765432", Email: "marie@example.com"},
		{Key: "phone:+33699887766", FirstName: "Léa", LastName: "Bernard", Name: "Léa Bernard", Phone: "+33699887766"},
	}

	t.Run("empty query returns all clients", func(t *testing.T) {
		assert.Len(t, SearchClientSummaries(groups, ""), 3)
	})

	t.Run("name fragment, case-insensitive", func(t *testing.T) {
		clients := SearchClientSummaries(groups, "marie")
		require.Len(t, clients, 1)
		assert.Equal(t, "Marie Martin", clients[0].Name)
	})

	t.Run("an account carries its client id and a split name", func(t *testing.T) {
		clients := SearchClientSummaries(groups, "dupont")
		require.Len(t, clients, 1)
		assert.Equal(t, &accountID, clients[0].ClientID)
		assert.Equal(t, "Jean", clients[0].FirstName)
		assert.Equal(t, "Dupont", clients[0].LastName)
	})

	t.Run("a guest carries the bookings' guest name and no client id", func(t *testing.T) {
		clients := SearchClientSummaries(groups, "bernard")
		require.Len(t, clients, 1)
		assert.Nil(t, clients[0].ClientID)
		assert.Equal(t, "Léa", clients[0].FirstName)
		assert.Equal(t, "Bernard", clients[0].LastName)
		assert.Equal(t, "+33699887766", clients[0].Phone)
	})

	t.Run("phone fragment matches ignoring separators", func(t *testing.T) {
		assert.Len(t, SearchClientSummaries(groups, "6123"), 1)
		assert.Len(t, SearchClientSummaries(groups, "06 12 34"), 1)
		assert.Len(t, SearchClientSummaries(groups, "9999"), 0)
	})

	t.Run("email fragment does not match (name or phone only)", func(t *testing.T) {
		assert.Len(t, SearchClientSummaries(groups, "example.com"), 0)
	})
}
