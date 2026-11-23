package notification

import (
	"context"
	"testing"
	"time"

	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
)

// A practitioner booking's client gets nothing from the app (issue 041,
// ADR-0021): no confirmation, reminder, cancellation or access update.
func TestPractitionerBooking_NoClientEmail(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.Origin = bookingPorts.BookingOriginPractitioner
	data.ClientEmail = "marie.dupont@example.com"
	spy := adapter.emailService.(*spyEmailService)

	ctx := context.Background()
	sends := map[string]func(context.Context, bookingPorts.BookingNotificationData) error{
		"confirmation":  adapter.SendBookingConfirmation,
		"reminder":      adapter.SendBookingReminder,
		"cancellation":  adapter.SendBookingCancellation,
		"access update": adapter.SendBookingAccessUpdate,
	}
	for name, send := range sends {
		if err := send(ctx, data); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	time.Sleep(100 * time.Millisecond) // the messages are fire-and-forget

	spy.mu.Lock()
	defer spy.mu.Unlock()
	if n := len(spy.bookingConfirmations) + len(spy.bookingReminders) + len(spy.clientCancellations) +
		len(spy.partnerCancellations) + len(spy.accessUpdates) + len(spy.partnerNotices); n != 0 {
		t.Fatalf("a practitioner booking sent %d emails, want none", n)
	}
	if calls := push.snapshot(); len(calls) != 0 {
		t.Fatalf("a practitioner booking sent %d notifications, want none", len(calls))
	}
}

// The same messages still go out for an online booking.
func TestOnlineBooking_ReminderStillSent(t *testing.T) {
	adapter, data := newPushAdapter(newFakePushNotifier(), time.Now())
	data.Origin = bookingPorts.BookingOriginOnline
	spy := adapter.emailService.(*spyEmailService)

	if err := adapter.SendBookingReminder(context.Background(), data); err != nil {
		t.Fatalf("SendBookingReminder: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		spy.mu.Lock()
		sent := len(spy.bookingReminders)
		spy.mu.Unlock()
		if sent == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no reminder was sent for an online booking")
}
