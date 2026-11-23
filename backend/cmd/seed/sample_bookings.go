package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5/pgxpool"

	bookingDomain "github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	bookingRepository "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/booking"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
)

// sampleBooking is one made-up appointment for the local agenda: every
// status and every payment state the Réservations page shows, so its cards
// can be checked without going through Stripe. Days are relative to the day
// the seed first runs; past ones land before the weekly availabilities
// `make dev` creates, upcoming ones on Mondays and Wednesdays, which those
// availabilities (Tue/Thu/Sat, and a one-off Friday) leave free.
type sampleBooking struct {
	firstName, lastName string
	// useClient books as the first seeded client account instead of a guest
	useClient     bool
	day           time.Time
	hour          int
	status        bookingDomain.BookingStatus
	paymentStatus bookingDomain.PaymentStatus
	cancelledBy   string
	// refundedPercent of what was paid online, for a cancellation
	refundedPercent int
}

// seedSampleBookings gives the practitioner a handful of bookings, once: it
// does nothing when the practitioner already has any. Local development only,
// enabled by "sample_bookings": true in the seed data. No payment intent is
// recorded, so cancelling one from the agenda never calls Stripe.
func seedSampleBookings(ctx context.Context, db *pgxpool.Pool, crypto encx.CryptoService, ids practiceIDs, product PracticeProductSeed, clientID *uuid.UUID) error {
	var existing int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM booking.bookings WHERE user_id = $1`, ids.UserID).Scan(&existing); err != nil {
		return fmt.Errorf("count practitioner bookings: %w", err)
	}
	if existing > 0 {
		return nil
	}

	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return fmt.Errorf("load Paris time zone: %w", err)
	}
	now := time.Now().In(paris)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, paris)
	next := func(weekday time.Weekday, weeksLater int) time.Time {
		days := (int(weekday) - int(today.Weekday()) + 7) % 7
		if days == 0 {
			days = 7
		}
		return today.AddDate(0, 0, days+7*weeksLater)
	}
	ago := func(days int) time.Time { return today.AddDate(0, 0, -days) }

	samples := []sampleBooking{
		{firstName: "Léa", lastName: "Bernard", day: next(time.Monday, 0), hour: 10, status: bookingDomain.BookingStatusConfirmed, paymentStatus: bookingDomain.PaymentStatusPaid},
		{useClient: true, day: next(time.Wednesday, 0), hour: 15, status: bookingDomain.BookingStatusConfirmed, paymentStatus: bookingDomain.PaymentStatusPaid},
		{firstName: "Thomas", lastName: "Petit", day: next(time.Monday, 1), hour: 14, status: bookingDomain.BookingStatusConfirmed, paymentStatus: bookingDomain.PaymentStatusDisputed},
		{firstName: "Sophie", lastName: "Laurent", day: ago(2), hour: 9, status: bookingDomain.BookingStatusConfirmed, paymentStatus: bookingDomain.PaymentStatusPaid},
		{firstName: "Hugo", lastName: "Moreau", day: ago(5), hour: 11, status: bookingDomain.BookingStatusCompleted, paymentStatus: bookingDomain.PaymentStatusPaid},
		{firstName: "Chloé", lastName: "Dubois", day: ago(7), hour: 16, status: bookingDomain.BookingStatusNoShow, paymentStatus: bookingDomain.PaymentStatusPaid},
		{firstName: "Julien", lastName: "Roux", day: ago(9), hour: 10, status: bookingDomain.BookingStatusCancelled, paymentStatus: bookingDomain.PaymentStatusRefunded, cancelledBy: "practitioner", refundedPercent: 100},
		{firstName: "Emma", lastName: "Garnier", day: ago(11), hour: 14, status: bookingDomain.BookingStatusCancelled, paymentStatus: bookingDomain.PaymentStatusPartiallyRefunded, cancelledBy: "client", refundedPercent: 50},
		{firstName: "Nathan", lastName: "Girard", day: ago(13), hour: 17, status: bookingDomain.BookingStatusCancelled, paymentStatus: bookingDomain.PaymentStatusFailed},
	}

	paidOnline, err := catalogDomain.AmountDueOnline(product.Deposit, product.PriceCents)
	if err != nil {
		return fmt.Errorf("sample bookings deposit: %w", err)
	}
	duration := time.Duration(product.DurationMinutes) * time.Minute

	// Each booking gets its own availability, encrypted up front so the
	// inserts below run back to back.
	availabilityRows := make([]*bookingDomain.AvailabilityEncx, len(samples))
	bookingRows := make([]*bookingDomain.BookingEncx, len(samples))
	for i, s := range samples {
		start := s.day.Add(time.Duration(s.hour) * time.Hour)
		end := start.Add(duration)

		availability := &bookingDomain.Availability{
			ID:          uuid.New(),
			UserID:      ids.UserID,
			RoomID:      ids.RoomID,
			StartTime:   start.UTC(),
			EndTime:     end.UTC(),
			MaxCapacity: 1,
			Status:      bookingDomain.AvailabilityStatusBooked,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if s.status == bookingDomain.BookingStatusCancelled {
			availability.Status = bookingDomain.AvailabilityStatusAvailable
		}
		if availabilityRows[i], err = bookingDomain.ProcessAvailabilityEncx(ctx, crypto, availability); err != nil {
			return fmt.Errorf("encrypt sample availability: %w", err)
		}

		amountPaid := paidOnline
		booking := &bookingDomain.Booking{
			ID:                    uuid.New(),
			AvailabilityID:        availability.ID,
			PartnerID:             ids.UserID,
			RoomID:                ids.RoomID,
			ProductID:             ids.ProductID,
			SlotStartTime:         start.UTC(),
			SlotEndTime:           end.UTC(),
			TotalPriceCents:       product.PriceCents,
			Currency:              product.Currency,
			PaymentStatus:         s.paymentStatus,
			Status:                s.status,
			Locale:                bookingDomain.BookingLocaleFR,
			AmountPaidOnlineCents: &amountPaid,
			CreatedAt:             start.AddDate(0, 0, -7),
			UpdatedAt:             now,
		}
		if s.useClient && clientID != nil {
			booking.ClientID = clientID
		} else {
			booking.GuestFirstName = s.firstName
			booking.GuestLastName = s.lastName
			booking.GuestEmail = "invite@example.com"
			booking.GuestPhone = "+33600000099"
		}
		switch s.status {
		case bookingDomain.BookingStatusCompleted, bookingDomain.BookingStatusNoShow:
			booking.CompletedAt = &end
		case bookingDomain.BookingStatusCancelled:
			cancelledAt := start.AddDate(0, 0, -2)
			booking.CancelledAt = &cancelledAt
		}
		if bookingRows[i], err = bookingDomain.ProcessBookingEncx(ctx, crypto, booking); err != nil {
			return fmt.Errorf("encrypt sample booking: %w", err)
		}
	}

	if err := insertSampleAvailabilities(ctx, db, availabilityRows); err != nil {
		return err
	}

	// The bookings go through the real repository. On a failure, everything
	// seeded here is removed, so the next run (which skips a practitioner who
	// has bookings) tries again from scratch.
	bookings := bookingRepository.New(ctx, db)
	err = func() error {
		for i, s := range samples {
			if err := bookings.Create(ctx, bookingRows[i]); err != nil {
				return fmt.Errorf("insert sample booking: %w", err)
			}
			// The insert leaves out who cancelled and what was refunded.
			if s.cancelledBy != "" {
				if _, err := db.Exec(ctx, `
					UPDATE booking.bookings SET cancelled_by = $2, refunded_amount_cents = $3 WHERE id = $1
				`, bookingRows[i].ID, s.cancelledBy, paidOnline*s.refundedPercent/100); err != nil {
					return fmt.Errorf("record sample cancellation: %w", err)
				}
			}
		}
		return nil
	}()
	if err != nil {
		bookingIDs := make([]uuid.UUID, len(bookingRows))
		availabilityIDs := make([]uuid.UUID, len(availabilityRows))
		for i := range bookingRows {
			bookingIDs[i], availabilityIDs[i] = bookingRows[i].ID, availabilityRows[i].ID
		}
		_, _ = db.Exec(ctx, `DELETE FROM booking.bookings WHERE id = ANY($1)`, bookingIDs)
		_, _ = db.Exec(ctx, `DELETE FROM booking.availabilities WHERE id = ANY($1)`, availabilityIDs)
		return err
	}

	log.Printf("Seeded %d sample bookings for the practitioner (local development only)", len(samples))
	return nil
}

// insertSampleAvailabilities writes the sample availabilities in one
// transaction. The schema refuses an availability that starts in the past
// (availabilities_start_time_check), yet every past booking was made on one:
// its availability simply was created before that day. To reproduce that
// state, the check is dropped for the inserts and added back NOT VALID —
// enforced on every later write, like the original, without rechecking rows
// that have since passed. Only the local database, only on the run that
// seeds the samples.
func insertSampleAvailabilities(ctx context.Context, db *pgxpool.Pool, rows []*bookingDomain.AvailabilityEncx) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin sample availabilities: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `ALTER TABLE booking.availabilities DROP CONSTRAINT availabilities_start_time_check`); err != nil {
		return fmt.Errorf("lift the past availability check: %w", err)
	}
	for _, a := range rows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO booking.availabilities (
				id, user_id, room_id, start_time, end_time,
				service_type_encrypted, price_cents, max_capacity,
				notes_encrypted, is_recurring, recurrence_pattern,
				status, created_at, updated_at, dek_encrypted, key_version, metadata
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		`, a.ID, a.UserID, a.RoomID, a.StartTime, a.EndTime,
			a.ServiceTypeEncrypted, a.PriceCents, a.MaxCapacity,
			a.NotesEncrypted, a.IsRecurring, a.RecurrencePattern,
			a.Status, a.CreatedAt, a.UpdatedAt, a.DEKEncrypted, a.KeyVersion, a.Metadata,
		); err != nil {
			return fmt.Errorf("insert sample availability: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
		ALTER TABLE booking.availabilities
		ADD CONSTRAINT availabilities_start_time_check CHECK (start_time >= NOW()) NOT VALID
	`); err != nil {
		return fmt.Errorf("restore the past availability check: %w", err)
	}
	return tx.Commit(ctx)
}
