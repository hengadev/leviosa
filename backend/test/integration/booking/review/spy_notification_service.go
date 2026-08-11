package review

import (
	"context"
	"sync"

	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
)

// SpyNotificationService is a no-op implementation of bookingPorts.BookingNotificationService,
// sufficient to satisfy the booking service's dependency for these review tests (which don't
// assert on notification content).
type SpyNotificationService struct {
	mu                   sync.Mutex
	BookingConfirmations []bookingPorts.BookingNotificationData
}

func NewSpyNotificationService() *SpyNotificationService {
	return &SpyNotificationService{}
}

func (s *SpyNotificationService) SendBookingConfirmation(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingConfirmations = append(s.BookingConfirmations, data)
	return nil
}

func (s *SpyNotificationService) SendPaymentConfirmation(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	return nil
}

func (s *SpyNotificationService) SendPaymentFailed(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	return nil
}

func (s *SpyNotificationService) SendBookingCancellation(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	return nil
}

func (s *SpyNotificationService) SendBookingReminder(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	return nil
}
