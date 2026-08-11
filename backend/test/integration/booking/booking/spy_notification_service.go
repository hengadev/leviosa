package booking

import (
	"context"
	"sync"

	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// SpyNotificationService records all notification calls for test assertions.
// It implements bookingPorts.BookingNotificationService.
type SpyNotificationService struct {
	mu                    sync.Mutex
	BookingConfirmations  []bookingPorts.BookingNotificationData
	PaymentConfirmations  []bookingPorts.BookingNotificationData
	PaymentFailed         []bookingPorts.BookingNotificationData
	BookingCancellations  []bookingPorts.BookingNotificationData
	BookingReminders      []bookingPorts.BookingNotificationData
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PaymentConfirmations = append(s.PaymentConfirmations, data)
	return nil
}

func (s *SpyNotificationService) SendPaymentFailed(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PaymentFailed = append(s.PaymentFailed, data)
	return nil
}

func (s *SpyNotificationService) SendBookingCancellation(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingCancellations = append(s.BookingCancellations, data)
	return nil
}

func (s *SpyNotificationService) SendBookingReminder(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingReminders = append(s.BookingReminders, data)
	return nil
}

// IntegrationSMSSpy records SMS send attempts made through the real
// notification.BookingNotificationAdapter (as opposed to SpyNotificationService,
// which substitutes for the whole adapter at the port level). It implements
// notificationPorts.SMSService.
type IntegrationSMSSpy struct {
	mu       sync.Mutex
	Messages []notificationDomain.GenericSMSRequest
}

func (s *IntegrationSMSSpy) SendOTP(ctx context.Context, req notificationDomain.OTPSMSRequest) error {
	return nil
}

func (s *IntegrationSMSSpy) SendSMS(ctx context.Context, req notificationDomain.GenericSMSRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, req)
	return nil
}

// IntegrationEmailSpy records booking-reminder (and other booking) email
// attempts made through the real notification.BookingNotificationAdapter. It
// implements notificationPorts.EmailService.
type IntegrationEmailSpy struct {
	mu                   sync.Mutex
	BookingConfirmations []notificationDomain.BookingConfirmationRequest
	BookingCancellations []notificationDomain.BookingCancellationRequest
	BookingReminders     []notificationDomain.BookingReminderRequest
}

func (s *IntegrationEmailSpy) SendOTPEmail(ctx context.Context, req notificationDomain.OTPEmailRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendWelcomeEmail(ctx context.Context, req notificationDomain.WelcomeEmailRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendVerifyEmailEmail(ctx context.Context, req notificationDomain.VerifyEmailRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendEventNotificationEmail(ctx context.Context, req notificationDomain.EventNotificationRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendPaymentNotificationEmail(ctx context.Context, req notificationDomain.PaymentNotificationRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendPaymentFailedEmail(ctx context.Context, req notificationDomain.PaymentNotificationRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendBookingConfirmationEmail(ctx context.Context, req notificationDomain.BookingConfirmationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingConfirmations = append(s.BookingConfirmations, req)
	return nil
}
func (s *IntegrationEmailSpy) SendBookingCancellationEmail(ctx context.Context, req notificationDomain.BookingCancellationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingCancellations = append(s.BookingCancellations, req)
	return nil
}
func (s *IntegrationEmailSpy) SendBookingReminderEmail(ctx context.Context, req notificationDomain.BookingReminderRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingReminders = append(s.BookingReminders, req)
	return nil
}
