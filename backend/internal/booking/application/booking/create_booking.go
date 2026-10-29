package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// CreateBooking creates a new booking for a specific product slot within an availability.
//
// Flow:
//  1. Validate and fetch availability
//  2. Validate and fetch product
//  3. Calculate slot end time based on product duration
//  4. Validate the start with the slot engine (grid, fit, overlap, notice,
//     horizon), exactly as the public slot endpoint computes slots
//  5. Get product price from Stripe
//  6. Create payment intent if price > 0
//  7. Create and persist booking entity
//
// Parameters:
//   - availabilityID: The availability block to book within
//   - clientID: The user making the booking
//   - productID: The product/service being booked
//   - slotStartTime: Desired start time (must be a slot the engine offers)
//
// No client notes are accepted: Leviosa stores no health data (ADR-0019).
//
// Returns:
//   - Created booking with encrypted fields
//   - Error if validation fails or slot is unavailable
func (s *BookingService) CreateBooking(
	ctx context.Context,
	availabilityID uuid.UUID,
	clientID *uuid.UUID,
	productID uuid.UUID,
	slotStartTime time.Time,
	guestFirstName, guestLastName, guestEmail, guestPhone string,
) (*domain.Booking, error) {
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

	// Verify availability is still open for bookings
	if !availability.IsAvailable() {
		return nil, fmt.Errorf("availability is no longer available for booking")
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
	policy, err := s.bookingPolicy(ctx)
	if err != nil {
		return nil, err
	}
	query, err := s.slotQuery(ctx, []*domain.Availability{availability}, product, policy, time.Now())
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
		return nil, errs.NewConflictErr(fmt.Errorf("time slot is already booked"))
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

	// 6. Create booking entity with slot information
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

	// Set slot information (new fields)
	booking.ProductID = productID
	booking.SlotStartTime = slotStartTime
	booking.SlotEndTime = slotEndTime

	// Generate signed booking token for public guest access
	if len(s.tokenSecret) > 0 {
		token := domain.GenerateBookingToken(booking.ID, booking.SlotEndTime, s.tokenSecret)
		booking.SetToken(token)
	}

	// 7. Create payment intent if booking has a price
	if totalPriceCents > 0 {
		description := fmt.Sprintf("Booking for %s - %s to %s",
			product.Name,
			slotStartTime.Format("Jan 02, 15:04"),
			slotEndTime.Format("15:04"))

		metadata := map[string]string{
			"booking_id":      booking.ID.String(),
			"availability_id": availabilityID.String(),
			"partner_id":      availability.UserID.String(),
			"product_id":      productID.String(),
			"slot_start":      slotStartTime.Format(time.RFC3339),
			"slot_end":        slotEndTime.Format(time.RFC3339),
		}
		if clientID != nil {
			metadata["client_id"] = clientID.String()
		}
		if booking.IsGuestBooking() {
			metadata["guest_name"] = booking.GuestDisplayName()
			metadata["guest_email"] = guestEmail
			metadata["guest_phone"] = guestPhone
		}

		// Resolve the email Stripe should send the hosted receipt to: the guest's
		// email for guest bookings, or the authenticated client's email otherwise.
		// Best-effort — a lookup failure should not block booking creation.
		receiptEmail := guestEmail
		if receiptEmail == "" && clientID != nil && s.authUserClient != nil {
			email, err := s.authUserClient.GetUserEmail(ctx, *clientID)
			if err != nil {
				slog.WarnContext(ctx, "failed to resolve client email for payment receipt",
					"client_id", *clientID, "err", err)
			} else {
				receiptEmail = email
			}
		}

		paymentIntentID, _, err := s.paymentService.CreatePaymentIntent(
			ctx,
			totalPriceCents,
			booking.Currency,
			description,
			receiptEmail,
			metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("create payment intent: %w", err)
		}

		booking.SetPaymentIntentID(paymentIntentID)
	}

	// 8. Encrypt and persist booking to repository
	bookingEncx, err := domain.ProcessBookingEncx(ctx, s.crypto, booking)
	if err != nil {
		return nil, fmt.Errorf("encrypt booking: %w", err)
	}

	if err := s.bookingRepo.Create(ctx, bookingEncx); err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}

	// 9. Send booking confirmation notification (best effort - don't fail booking on notification error)
	if s.notificationService != nil {
		notificationData := s.buildNotificationData(booking, product.Name)
		if err := s.notificationService.SendBookingConfirmation(ctx, notificationData); err != nil {
			slog.WarnContext(ctx, "failed to send booking confirmation notification",
				"booking_id", booking.ID,
				"is_guest", booking.IsGuestBooking(),
				"err", err)
		}
	}

	return booking, nil
}
