package booking

import (
	"context"
	"sync"

	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/google/uuid"
)

// SpyNotificationService records all notification calls for test assertions.
// It implements bookingPorts.BookingNotificationService.
type SpyNotificationService struct {
	mu                   sync.Mutex
	BookingConfirmations []bookingPorts.BookingNotificationData
	PaymentConfirmations []bookingPorts.BookingNotificationData
	PaymentFailed        []bookingPorts.BookingNotificationData
	BookingCancellations []bookingPorts.BookingNotificationData
	BookingReminders     []bookingPorts.BookingNotificationData
	DisputeAlerts        []bookingPorts.BookingNotificationData
	LatePaymentRefunds   []bookingPorts.BookingNotificationData

	// forward, when set, also hands every booking confirmation (and
	// cancellation) to a real notification service (the
	// BookingNotificationAdapter), so a test can follow the event down to the
	// emails it sends.
	forward bookingPorts.BookingNotificationService
}

// ForwardConfirmationsTo makes the spy hand booking confirmations to next as
// well; the returned function stops it.
func (s *SpyNotificationService) ForwardConfirmationsTo(next bookingPorts.BookingNotificationService) (stop func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forward = next
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.forward = nil
	}
}

func NewSpyNotificationService() *SpyNotificationService {
	return &SpyNotificationService{}
}

func (s *SpyNotificationService) SendBookingConfirmation(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	s.BookingConfirmations = append(s.BookingConfirmations, data)
	next := s.forward
	s.mu.Unlock()
	if next != nil {
		return next.SendBookingConfirmation(ctx, data)
	}
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
	s.BookingCancellations = append(s.BookingCancellations, data)
	next := s.forward
	s.mu.Unlock()
	if next != nil {
		return next.SendBookingCancellation(ctx, data)
	}
	return nil
}

func (s *SpyNotificationService) SendBookingReminder(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingReminders = append(s.BookingReminders, data)
	return nil
}

func (s *SpyNotificationService) SendPaymentDisputeAlert(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DisputeAlerts = append(s.DisputeAlerts, data)
	return nil
}

// SendLatePaymentRefund records the late-payment refund notification (the
// spy also stands for bookingPorts.LatePaymentNotificationService) and hands
// it to the forward target when that one sends it too.
func (s *SpyNotificationService) SendLatePaymentRefund(ctx context.Context, data bookingPorts.BookingNotificationData) error {
	s.mu.Lock()
	s.LatePaymentRefunds = append(s.LatePaymentRefunds, data)
	next := s.forward
	s.mu.Unlock()
	if late, ok := next.(bookingPorts.LatePaymentNotificationService); ok {
		return late.SendLatePaymentRefund(ctx, data)
	}
	return nil
}

// CountFor counts the notifications of a kind recorded for a booking.
func (s *SpyNotificationService) CountFor(kind func(*SpyNotificationService) []bookingPorts.BookingNotificationData, bookingID uuid.UUID) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, data := range kind(s) {
		if data.BookingID == bookingID {
			count++
		}
	}
	return count
}

// IntegrationSMSSpy records SMS send attempts made through the real
// notification.BookingNotificationAdapter (as opposed to SpyNotificationService,
// which substitutes for the whole adapter at the port level). It implements
// notificationPorts.SMSService.
type IntegrationSMSSpy struct {
	mu       sync.Mutex
	Messages []notificationDomain.GenericSMSRequest
}

// Count returns how many SMS were attempted.
func (s *IntegrationSMSSpy) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Messages)
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
	ClientCancellations  []notificationDomain.BookingClientCancellationRequest
	PartnerCancellations []notificationDomain.BookingPartnerCancellationRequest
	BookingReminders     []notificationDomain.BookingReminderRequest
	PartnerNotices       []notificationDomain.BookingPartnerNoticeRequest
	LatePaymentRefunds   []notificationDomain.LatePaymentRefundRequest
}

// LatePaymentRefundSnapshot returns copies of the late-payment refund emails
// recorded so far.
func (s *IntegrationEmailSpy) LatePaymentRefundSnapshot() []notificationDomain.LatePaymentRefundRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]notificationDomain.LatePaymentRefundRequest(nil), s.LatePaymentRefunds...)
}

func (s *IntegrationEmailSpy) SendLatePaymentRefundEmail(ctx context.Context, req notificationDomain.LatePaymentRefundRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LatePaymentRefunds = append(s.LatePaymentRefunds, req)
	return nil
}

// Snapshot returns copies of the booking-confirmed emails recorded so far.
func (s *IntegrationEmailSpy) Snapshot() ([]notificationDomain.BookingConfirmationRequest, []notificationDomain.BookingPartnerNoticeRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]notificationDomain.BookingConfirmationRequest(nil), s.BookingConfirmations...),
		append([]notificationDomain.BookingPartnerNoticeRequest(nil), s.PartnerNotices...)
}

// Reminders returns a copy of the booking reminders recorded so far.
func (s *IntegrationEmailSpy) Reminders() []notificationDomain.BookingReminderRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]notificationDomain.BookingReminderRequest(nil), s.BookingReminders...)
}

// Reset forgets every recorded email.
func (s *IntegrationEmailSpy) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingConfirmations, s.BookingReminders, s.PartnerNotices = nil, nil, nil
	s.ClientCancellations, s.PartnerCancellations = nil, nil
	s.LatePaymentRefunds = nil
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
func (s *IntegrationEmailSpy) SendBookingPartnerNoticeEmail(ctx context.Context, req notificationDomain.BookingPartnerNoticeRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PartnerNotices = append(s.PartnerNotices, req)
	return nil
}
func (s *IntegrationEmailSpy) SendBookingClientCancellationEmail(ctx context.Context, req notificationDomain.BookingClientCancellationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ClientCancellations = append(s.ClientCancellations, req)
	return nil
}
func (s *IntegrationEmailSpy) SendBookingPartnerCancellationEmail(ctx context.Context, req notificationDomain.BookingPartnerCancellationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PartnerCancellations = append(s.PartnerCancellations, req)
	return nil
}

// CancellationSnapshot returns copies of the cancellation emails recorded so
// far: the clients' and the partners'.
func (s *IntegrationEmailSpy) CancellationSnapshot() ([]notificationDomain.BookingClientCancellationRequest, []notificationDomain.BookingPartnerCancellationRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]notificationDomain.BookingClientCancellationRequest(nil), s.ClientCancellations...),
		append([]notificationDomain.BookingPartnerCancellationRequest(nil), s.PartnerCancellations...)
}
func (s *IntegrationEmailSpy) SendBookingReminderEmail(ctx context.Context, req notificationDomain.BookingReminderRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BookingReminders = append(s.BookingReminders, req)
	return nil
}
func (s *IntegrationEmailSpy) SendContactFormEmail(ctx context.Context, req notificationDomain.ContactFormRequest) error {
	return nil
}
func (s *IntegrationEmailSpy) SendPaymentDisputeAlertEmail(ctx context.Context, req notificationDomain.PaymentDisputeAlertRequest) error {
	return nil
}
