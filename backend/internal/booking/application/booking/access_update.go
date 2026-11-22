package booking

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
)

// accessUpdateTarget is an upcoming confirmed booking and its client's
// email, empty when the client has none.
type accessUpdateTarget struct {
	booking *domain.Booking
	email   string
}

// GetAccessUpdateRecipients counts the upcoming confirmed bookings, split
// into those with a client email and those without.
func (s *BookingService) GetAccessUpdateRecipients(ctx context.Context) (*domain.AccessUpdateRecipients, error) {
	targets, err := s.accessUpdateTargets(ctx)
	if err != nil {
		return nil, err
	}
	counts := &domain.AccessUpdateRecipients{}
	for _, t := range targets {
		if t.email == "" {
			counts.WithoutEmail++
		} else {
			counts.Recipients++
		}
	}
	return counts, nil
}

// SendAccessUpdate emails the current practice access to the client of every
// upcoming confirmed booking with an email. Sending is fire-and-forget per
// booking: a booking counts as sent once handed to the notifier. Refused
// with an invalid-value error when no door code is set. The door code and
// the instructions are read at send time by the notifier and never logged
// or audited here.
func (s *BookingService) SendAccessUpdate(ctx context.Context) (*domain.AccessUpdateRecipients, error) {
	if s.practiceAccess == nil {
		return nil, errs.NewInvalidValueErr("no door code is set")
	}
	// The door code is the practice's, the same for every room.
	access, err := s.practiceAccess.GetPracticeAccess(ctx, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("read practice access: %w", err)
	}
	if strings.TrimSpace(access.DoorCode) == "" {
		return nil, errs.NewInvalidValueErr("no door code is set")
	}

	targets, err := s.accessUpdateTargets(ctx)
	if err != nil {
		return nil, err
	}

	result := &domain.AccessUpdateRecipients{}
	for _, t := range targets {
		if t.email == "" {
			result.WithoutEmail++
			continue
		}
		if s.accessNotifier == nil {
			continue
		}
		data := s.buildNotificationData(t.booking, "")
		data.ClientEmail = t.email
		if err := s.accessNotifier.SendBookingAccessUpdate(ctx, data); err != nil {
			slog.ErrorContext(ctx, "failed to send access update notification",
				"booking_id", t.booking.ID, "error", err)
			continue
		}
		result.Recipients++
	}

	actorID, actorRole := audit.ActorFromContext(ctx)
	audit.WriteOrLog(ctx, s.audit, audit.Entry{
		ActorID:    actorID,
		ActorRole:  actorRole,
		Action:     "booking.access_update_sent",
		TargetType: "practice_access",
		Payload: map[string]any{
			"recipients":    result.Recipients,
			"without_email": result.WithoutEmail,
		},
	})

	return result, nil
}

// accessUpdateTargets lists the upcoming confirmed bookings (confirmed, slot
// not started yet) with their client's email. Slot times are encrypted, so
// all confirmed bookings are listed and filtered after decryption, as the
// reminder scheduler does. A guest's email is on the booking; a registered
// client's is resolved through the auth-user client, and a client whose email
// cannot be resolved counts as without email.
func (s *BookingService) accessUpdateTargets(ctx context.Context) ([]accessUpdateTarget, error) {
	bookingsEncx, err := s.bookingRepo.GetUpcoming(ctx, ports.BookingFilter{})
	if err != nil {
		return nil, fmt.Errorf("list upcoming bookings: %w", err)
	}

	now := s.now()
	targets := make([]accessUpdateTarget, 0, len(bookingsEncx))
	for _, bx := range bookingsEncx {
		booking, err := domain.DecryptBookingEncx(ctx, s.crypto, bx)
		if err != nil {
			slog.ErrorContext(ctx, "access update: failed to decrypt booking",
				"booking_id", bx.ID, "error", err)
			continue
		}
		if booking.Status != domain.BookingStatusConfirmed || !booking.SlotStartTime.After(now) {
			continue
		}
		targets = append(targets, accessUpdateTarget{booking: booking, email: s.clientEmail(ctx, booking)})
	}
	return targets, nil
}

// clientEmail is the email of the booking's client: the guest email on a
// guest booking, the account email of a registered client, or "" when none
// is known.
func (s *BookingService) clientEmail(ctx context.Context, booking *domain.Booking) string {
	if booking.IsGuestBooking() {
		return strings.TrimSpace(booking.GuestEmail)
	}
	if s.authUserClient == nil {
		return ""
	}
	email, err := s.authUserClient.GetUserEmail(ctx, *booking.ClientID)
	if err != nil {
		slog.WarnContext(ctx, "access update: failed to resolve the client email",
			"booking_id", booking.ID, "client_id", *booking.ClientID, "error", err)
		return ""
	}
	return strings.TrimSpace(email)
}
