package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// CreateBooking creates a new booking for a specific product slot within an
// availability. The booking starts awaiting payment and holds its slot for
// the payment hold (ADR-0018); the payment-succeeded webhook confirms it.
//
// Flow:
//  1. Validate and fetch availability
//  2. Validate and fetch product
//  3. Calculate slot end time based on product duration
//  4. Validate the start with the slot engine (grid, fit, overlap, notice,
//     horizon), exactly as the public slot endpoint computes slots
//  5. Get product price; the amount due online is the product's deposit when
//     one is set, otherwise the full price
//  6. Create the booking, awaiting payment, held until now + payment hold
//  7. Create the PaymentIntent for the amount due online; its client secret
//     is returned on the booking (never persisted). With nothing due online
//     the booking is confirmed at once
//  8. Persist under a lock on the partner's agenda, re-checking the slot
//     against all the partner's bookings, so two clients submitting
//     overlapping slots cannot both get them
//
// No notification is sent while the booking awaits payment: the
// booking-confirmed event is emitted when the payment webhook confirms it.
// No client notes are accepted: Leviosa stores no health data (ADR-0019).
func (s *BookingService) CreateBooking(
	ctx context.Context,
	availabilityID uuid.UUID,
	clientID *uuid.UUID,
	productID uuid.UUID,
	slotStartTime time.Time,
	guestFirstName, guestLastName, guestEmail, guestPhone string,
	locale domain.BookingLocale,
) (*domain.Booking, error) {
	if locale == "" {
		locale = domain.BookingLocaleFR
	}
	if _, err := domain.ParseBookingLocale(string(locale)); err != nil {
		return nil, errs.NewInvalidValueErr(err.Error())
	}

	// 1. Fetch and decrypt availability
	availabilityEncx, err := s.availabilityRepo.GetByID(ctx, availabilityID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("availability not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get availability: %w", err)
	}

	availability, err := domain.DecryptAvailabilityEncx(ctx, s.crypto, availabilityEncx)
	if err != nil {
		return nil, errs.NewNotDecryptedErr("availability", err)
	}

	// Verify availability is still open for bookings: a one-off one that has
	// not started, a recurring one whose series has not ended (the engine
	// below checks the requested occurrence).
	if !availability.IsAvailable() {
		return nil, errs.NewConflictErr(errAvailabilityClosed)
	}

	// 2. Fetch product from catalog service
	product, err := s.productService.GetProductByID(ctx, productID.String())
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("product not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	// 3. Calculate slot end time based on product duration
	slotEndTime := slotStartTime.Add(time.Duration(product.Duration) * time.Minute)

	// 4. Validate the start with the slot engine, as the public slot endpoint
	// computes it: a start the practice page would not offer is refused. The
	// engine runs twice so a start that is on the grid but already taken
	// answers a conflict rather than an invalid value.
	now := time.Now()
	policy, err := s.bookingPolicy(ctx)
	if err != nil {
		return nil, err
	}
	// The windows are the availability's occurrences containing the start:
	// for a recurring availability, any occurrence of its series.
	windows := availability.Occurrences(slotStartTime, slotStartTime.Add(time.Nanosecond))
	query, err := s.slotQuery(ctx, availability.UserID, windows, product, policy, now)
	if err != nil {
		return nil, err
	}
	occupancies := query.Occupancies
	query.Occupancies = nil
	if !domain.OffersSlot(query, slotStartTime) {
		return nil, errs.NewInvalidValueErr("slot_start_time: not a slot offered for this availability and product")
	}
	query.Occupancies = occupancies
	if !domain.OffersSlot(query, slotStartTime) {
		return nil, errs.NewConflictErr(errSlotTaken)
	}

	// 5. Get product price from pricing service
	// Default currency for Europe-based business
	const defaultCurrency = "EUR"

	price, err := s.priceService.GetActiveOneTimePriceByProductID(ctx, productID.String(), defaultCurrency)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("no active price found for product: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get product price: %w", err)
	}

	totalPriceCents := price.Amount
	amountDueOnline, err := catalogDomain.AmountDueOnline(product.Deposit, totalPriceCents)
	if err != nil {
		return nil, fmt.Errorf("amount due online: %w", err)
	}

	// 6. Create booking entity with slot information, awaiting payment
	booking, err := domain.NewBooking(
		availabilityID,
		clientID,
		availability.UserID, // partner ID
		availability.RoomID,
		totalPriceCents,
		price.Currency,
		guestFirstName,
		guestLastName,
		guestEmail,
		guestPhone,
	)
	if err != nil {
		return nil, fmt.Errorf("create booking entity: %w", err)
	}

	booking.ProductID = productID
	booking.SlotStartTime = slotStartTime
	booking.SlotEndTime = slotEndTime
	booking.Locale = locale
	booking.AmountDueOnlineCents = amountDueOnline

	hold := policy.PaymentHold
	if hold <= 0 {
		hold = domain.DefaultPaymentHold
	}
	if err := booking.HoldUntil(now.Add(hold)); err != nil {
		return nil, fmt.Errorf("start payment hold: %w", err)
	}

	// Generate signed booking token for public guest access
	if len(s.tokenSecret) > 0 {
		token := domain.GenerateBookingToken(booking.ID, booking.SlotEndTime, s.tokenSecret)
		booking.SetToken(token)
	}

	// 7. Create the PaymentIntent for the amount due online. With nothing due
	// online there is no payment to wait for: the booking is confirmed now.
	if amountDueOnline > 0 {
		paymentIntentID, clientSecret, err := s.createBookingPaymentIntent(ctx, booking, product.Name, amountDueOnline)
		if err != nil {
			return nil, err
		}
		booking.SetPaymentIntentID(paymentIntentID)
		booking.PaymentClientSecret = clientSecret
	} else if err := booking.Confirm(); err != nil {
		return nil, fmt.Errorf("confirm booking with nothing due online: %w", err)
	}

	// 8. Encrypt and persist under the partner's agenda lock, re-checking the
	// slot against the partner's bookings committed meanwhile, in any of
	// their availabilities.
	bookingEncx, err := domain.ProcessBookingEncx(ctx, s.crypto, booking)
	if err != nil {
		return nil, fmt.Errorf("encrypt booking: %w", err)
	}

	err = s.bookingRepo.CreateGuarded(ctx, bookingEncx, func(existing []*domain.BookingEncx) error {
		buffers := map[uuid.UUID]time.Duration{product.ID: query.BufferTime}
		occupancies, err := s.occupancies(ctx, existing, buffers, query.BufferTime, now)
		if err != nil {
			return err
		}
		query.Occupancies = occupancies
		if !domain.OffersSlot(query, slotStartTime) {
			return errs.NewConflictErr(errSlotTaken)
		}
		return nil
	})
	if err != nil {
		if booking.PaymentIntentID != nil {
			// Nobody will pay this intent: cancel it (best effort).
			if cancelErr := s.paymentService.CancelPaymentIntent(ctx, *booking.PaymentIntentID); cancelErr != nil {
				slog.WarnContext(ctx, "failed to cancel the payment intent of a booking that was not created",
					"booking_id", booking.ID, "err", cancelErr)
			}
		}
		if errors.Is(err, errs.ErrConflict) {
			return nil, err
		}
		return nil, fmt.Errorf("create booking: %w", err)
	}

	// Nothing was due online: the booking is confirmed now, so the
	// booking-confirmed event goes out now.
	if booking.Status == domain.BookingStatusConfirmed {
		s.sendBookingConfirmed(ctx, booking, product.Name)
	}

	return booking, nil
}

// errAvailabilityClosed is the conflict a booking gets on an availability
// that was deleted (cancelled), blocked or has ended.
var errAvailabilityClosed = errors.New("availability is no longer available for booking")

// errSlotTaken is the conflict a booking gets when another one holds its slot.
var errSlotTaken = errors.New("time slot is already booked")

// createBookingPaymentIntent creates the PaymentIntent for the amount due
// online, with the booking in its metadata so the webhook finds it.
func (s *BookingService) createBookingPaymentIntent(ctx context.Context, booking *domain.Booking, productName string, amount int) (string, string, error) {
	description := fmt.Sprintf("Booking for %s - %s to %s",
		productName,
		booking.SlotStartTime.Format("Jan 02, 15:04"),
		booking.SlotEndTime.Format("15:04"))

	metadata := map[string]string{
		"booking_id":      booking.ID.String(),
		"availability_id": booking.AvailabilityID.String(),
		"partner_id":      booking.PartnerID.String(),
		"product_id":      booking.ProductID.String(),
		"slot_start":      booking.SlotStartTime.Format(time.RFC3339),
		"slot_end":        booking.SlotEndTime.Format(time.RFC3339),
	}
	if booking.ClientID != nil {
		metadata["client_id"] = booking.ClientID.String()
	}
	if booking.IsGuestBooking() {
		metadata["guest_name"] = booking.GuestDisplayName()
		metadata["guest_email"] = booking.GuestEmail
		metadata["guest_phone"] = booking.GuestPhone
	}

	// Resolve the email Stripe should send the hosted receipt to: the guest's
	// email for guest bookings, or the authenticated client's email otherwise.
	// Best-effort — a lookup failure should not block booking creation.
	receiptEmail := booking.GuestEmail
	if receiptEmail == "" && booking.ClientID != nil && s.authUserClient != nil {
		email, err := s.authUserClient.GetUserEmail(ctx, *booking.ClientID)
		if err != nil {
			slog.WarnContext(ctx, "failed to resolve client email for payment receipt",
				"client_id", *booking.ClientID, "err", err)
		} else {
			receiptEmail = email
		}
	}

	paymentIntentID, clientSecret, err := s.paymentService.CreatePaymentIntent(
		ctx,
		amount,
		booking.Currency,
		description,
		receiptEmail,
		metadata,
	)
	if err != nil {
		return "", "", fmt.Errorf("create payment intent: %w", err)
	}
	return paymentIntentID, clientSecret, nil
}

// sendBookingConfirmed emits the booking-confirmed event (best effort: a
// notification failure never undoes the confirmation).
func (s *BookingService) sendBookingConfirmed(ctx context.Context, booking *domain.Booking, productName string) {
	if s.notificationService == nil {
		return
	}
	notificationData := s.buildNotificationData(booking, productName)
	if err := s.notificationService.SendBookingConfirmation(ctx, notificationData); err != nil {
		slog.WarnContext(ctx, "failed to send booking confirmation notification",
			"booking_id", booking.ID,
			"is_guest", booking.IsGuestBooking(),
			"err", err)
	}
}
