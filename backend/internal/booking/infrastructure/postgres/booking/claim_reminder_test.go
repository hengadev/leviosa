package bookingRepository_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tb "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaimReminder(t *testing.T) {
	ctx := context.Background()

	t.Run("claims a confirmed unreminded booking once", func(t *testing.T) {
		tb.ClearBookingsTable(t, ctx, testPool)

		booking := tb.NewTestBookingEncx(t)
		booking.Status = domain.BookingStatusConfirmed
		require.NoError(t, tb.InsertBookingEncx(t, ctx, testPool, booking))

		claimed, err := repo.ClaimReminder(ctx, booking.ID)
		require.NoError(t, err)
		assert.True(t, claimed)

		claimed, err = repo.ClaimReminder(ctx, booking.ID)
		require.NoError(t, err)
		assert.False(t, claimed, "a booking is claimed for a reminder only once")

		updated, err := tb.GetBookingEncxByID(t, ctx, testPool, booking.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.RemindedAt)
	})

	t.Run("concurrent claims: exactly one wins", func(t *testing.T) {
		tb.ClearBookingsTable(t, ctx, testPool)

		booking := tb.NewTestBookingEncx(t)
		booking.Status = domain.BookingStatusConfirmed
		require.NoError(t, tb.InsertBookingEncx(t, ctx, testPool, booking))

		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			wins int
		)
		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				claimed, err := repo.ClaimReminder(ctx, booking.ID)
				assert.NoError(t, err)
				if claimed {
					mu.Lock()
					wins++
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		assert.Equal(t, 1, wins)
	})

	for _, status := range []domain.BookingStatus{
		domain.BookingStatusCancelled,
		domain.BookingStatusExpired,
		domain.BookingStatusAwaitingPayment,
	} {
		t.Run("does not claim a "+string(status)+" booking", func(t *testing.T) {
			tb.ClearBookingsTable(t, ctx, testPool)

			booking := tb.NewTestBookingEncx(t)
			booking.Status = status
			require.NoError(t, tb.InsertBookingEncx(t, ctx, testPool, booking))

			claimed, err := repo.ClaimReminder(ctx, booking.ID)
			require.NoError(t, err)
			assert.False(t, claimed)

			updated, err := tb.GetBookingEncxByID(t, ctx, testPool, booking.ID)
			require.NoError(t, err)
			assert.Nil(t, updated.RemindedAt)
		})
	}

	t.Run("unknown booking is not claimed", func(t *testing.T) {
		claimed, err := repo.ClaimReminder(ctx, uuid.New())
		require.NoError(t, err)
		assert.False(t, claimed)
	})
}
