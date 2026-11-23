package booking

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
)

// Practitioner booking (issue 041, ADR-0021): the practitioner books a
// client who called. Beside the online CreateBooking, not a flag on it:
//
//   - the slot rules are the slot engine's (inside an open availability
//     occurrence, exceptions excluded, on the grid, fitting, not overlapping
//     another booking with its buffer) without the minimum notice and the
//     booking horizon (BookingPolicy.ForPractitioner);
//   - the booking is confirmed at once: no payment hold, no PaymentIntent,
//     nothing paid online, so the full price is due at the practice;
//   - its origin is practitioner, and the app sends its client nothing: no
//     confirmation (sendBookingConfirmed is never called here), no reminder,
//     no cancellation email, no practice access, and no booking token;
//   - it is persisted through the same guarded insert, under the partner's
//     agenda lock, as an online booking, so the two can't take one slot.

// priceCurrency is the currency bookings are priced in.
const priceCurrency = "EUR"

// CreatePractitionerBooking books a slot of a practitioner's agenda for a
// client who called. The caller must be the availability's partner
// (callerID) or an administrator (callerIsAdmin); anyone else gets
// errs.ErrForbidden.
func (s *BookingService) CreatePractitionerBooking(
	ctx context.Context,
	input domain.PractitionerBookingInput,
	callerID uuid.UUID,
	callerIsAdmin bool,
) (*domain.Booking, error) {
	input.Normalize()
	if problem := input.ClientProblem(); problem != "" {
		return nil, errs.NewInvalidValueErr(problem)
	}

	availability, err := s.loadAvailability(ctx, input.AvailabilityID)
	if err != nil {
		return nil, err
	}
	if availability.UserID != callerID && !callerIsAdmin {
		return nil, errs.NewForbiddenErr("availability belongs to another partner")
	}
	now := s.now()
	if !availability.IsOpenForPractitioner(now) {
		return nil, errs.NewConflictErr(errAvailabilityClosed)
	}

	product, err := s.practitionerProduct(ctx, availability.UserID, input.ProductID)
	if err != nil {
		return nil, err
	}

	policy, err := s.bookingPolicy(ctx)
	if err != nil {
		return nil, err
	}
	query, err := s.checkSlot(ctx, availability, product, policy.ForPractitioner(), now, input.SlotStartTime)
	if err != nil {
		return nil, err
	}

	price, err := s.priceService.GetActiveOneTimePriceByProductID(ctx, product.ID.String(), priceCurrency)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("no active price found for product: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get product price: %w", err)
	}

	booking, err := domain.NewBooking(
		availability.ID,
		input.ClientID,
		availability.UserID,
		availability.RoomID,
		price.Amount,
		price.Currency,
		input.GuestFirstName,
		input.GuestLastName,
		input.GuestEmail,
		input.GuestPhone,
	)
	if err != nil {
		return nil, errs.NewInvalidValueErr(fmt.Sprintf("booking: %v", err))
	}
	booking.ProductID = product.ID
	booking.SlotStartTime = input.SlotStartTime
	booking.SlotEndTime = input.SlotStartTime.Add(time.Duration(product.Duration) * time.Minute)
	booking.Locale = domain.BookingLocaleFR
	booking.Origin = domain.BookingOriginPractitioner

	// Nothing is due online: confirmed now, the full price due at the
	// practice (Booking.AmountDueOnSiteCents).
	if err := booking.Confirm(); err != nil {
		return nil, fmt.Errorf("confirm practitioner booking: %w", err)
	}
	nothingPaid := 0
	booking.AmountPaidOnlineCents = &nothingPaid

	bookingEncx, err := domain.ProcessBookingEncx(ctx, s.crypto, booking)
	if err != nil {
		return nil, fmt.Errorf("encrypt booking: %w", err)
	}
	if err := s.bookingRepo.CreateGuarded(ctx, bookingEncx, s.slotStillFree(ctx, product, query, input.SlotStartTime, now)); err != nil {
		if errors.Is(err, errs.ErrConflict) {
			return nil, err
		}
		return nil, fmt.Errorf("create booking: %w", err)
	}

	actorID, actorRole := audit.ActorFromContext(ctx)
	audit.WriteOrLog(ctx, s.audit, audit.Entry{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     "booking.created_by_practitioner",
		TargetType: "booking",
		TargetID:   booking.ID.String(),
		Payload: map[string]any{
			"partner_id": booking.PartnerID.String(),
			"product_id": booking.ProductID.String(),
			"is_guest":   booking.IsGuestBooking(),
		},
	})

	return booking, nil
}

// GetPractitionerFreeSlots returns the free slots of a practitioner's
// availabilities (partnerUserID: the partner's user ID) for one of their
// products, starting in [from, to): the slots the practitioner's form
// offers, without the minimum notice and the booking horizon. The public
// slot endpoint (GetFreeSlots) is unchanged.
func (s *BookingService) GetPractitionerFreeSlots(ctx context.Context, partnerUserID, productID uuid.UUID, from, to time.Time) (*domain.FreeSlots, error) {
	if !from.Before(to) {
		return nil, errs.NewInvalidValueErr("from must be before to")
	}
	product, err := s.practitionerProduct(ctx, partnerUserID, productID)
	if err != nil {
		return nil, err
	}
	policy, err := s.bookingPolicy(ctx)
	if err != nil {
		return nil, err
	}
	return s.freeSlots(ctx, partnerUserID, product, policy.ForPractitioner(), s.now(), from, to)
}

// GetPractitionerProducts lists the products a practitioner can book for a
// client who called: their published products with an active price, by
// name.
func (s *BookingService) GetPractitionerProducts(ctx context.Context, partnerUserID uuid.UUID) ([]domain.PractitionerProduct, error) {
	offered, err := s.offeredProductIDs(ctx, partnerUserID)
	if err != nil {
		return nil, err
	}

	var candidates []*catalogDomain.ProductRes
	if offered == nil {
		candidates, err = s.productService.GetAllPublishedProducts(ctx)
		if err != nil && !errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("list published products: %w", err)
		}
	} else {
		for _, id := range offered {
			product, err := s.productService.GetProductByID(ctx, id.String())
			if err != nil {
				if errors.Is(err, errs.ErrRepositoryNotFound) {
					continue
				}
				return nil, fmt.Errorf("get product %s: %w", id, err)
			}
			candidates = append(candidates, product)
		}
	}

	products := make([]domain.PractitionerProduct, 0, len(candidates))
	for _, product := range candidates {
		if product == nil || product.Status != catalogDomain.Published {
			continue
		}
		price, err := s.priceService.GetActiveOneTimePriceByProductID(ctx, product.ID.String(), priceCurrency)
		if err != nil {
			if errors.Is(err, errs.ErrRepositoryNotFound) {
				continue // no price: not bookable
			}
			return nil, fmt.Errorf("get price of product %s: %w", product.ID, err)
		}
		products = append(products, domain.PractitionerProduct{
			ID:              product.ID,
			Name:            product.Name,
			DurationMinutes: product.Duration,
			PriceCents:      price.Amount,
			Currency:        price.Currency,
		})
	}
	sort.SliceStable(products, func(i, j int) bool { return products[i].Name < products[j].Name })
	return products, nil
}

// practitionerProduct reads a product the practitioner may book: published,
// and one they offer.
func (s *BookingService) practitionerProduct(ctx context.Context, partnerUserID, productID uuid.UUID) (*catalogDomain.ProductRes, error) {
	product, err := s.productService.GetProductByID(ctx, productID.String())
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, fmt.Errorf("product not found: %w", errs.ErrRepositoryNotFound)
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	if product.Status != catalogDomain.Published {
		return nil, errs.NewConflictErr(errProductNotOffered)
	}
	offered, err := s.offeredProductIDs(ctx, partnerUserID)
	if err != nil {
		return nil, err
	}
	if offered != nil && !slices.Contains(offered, product.ID) {
		return nil, errs.NewConflictErr(errProductNotOffered)
	}
	return product, nil
}

// offeredProductIDs returns the products the partner's profile offers, or
// nil when it lists none (or no partner profile is reachable): the practice
// has one practitioner, whose bookable products are then every published
// one.
func (s *BookingService) offeredProductIDs(ctx context.Context, partnerUserID uuid.UUID) ([]uuid.UUID, error) {
	if s.authUserClient == nil {
		return nil, nil
	}
	partner, err := s.authUserClient.GetPartnerByUserID(ctx, partnerUserID)
	if err != nil {
		if errors.Is(err, errs.ErrRepositoryNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get partner of user %s: %w", partnerUserID, err)
	}
	if partner == nil || len(partner.ProductIDs) == 0 {
		return nil, nil
	}
	return partner.ProductIDs, nil
}

// loadAvailability reads and decrypts an availability.
func (s *BookingService) loadAvailability(ctx context.Context, id uuid.UUID) (*domain.Availability, error) {
	availabilityEncx, err := s.availabilityRepo.GetByID(ctx, id)
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
	return availability, nil
}
