package notification

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"

	"github.com/google/uuid"
)

// Practitioner notifications on a new booking (issue 044, ADR-0020), with a
// fake push notifier in place of Web Push.

type pushCall struct {
	userID uuid.UUID
	msg    notificationDomain.PushMessage
}

type fakePushNotifier struct {
	mu    sync.Mutex
	calls []pushCall
	done  chan struct{}
	err   error
}

func newFakePushNotifier() *fakePushNotifier {
	return &fakePushNotifier{done: make(chan struct{}, 8)}
}

func (f *fakePushNotifier) NotifyUser(_ context.Context, userID uuid.UUID, msg notificationDomain.PushMessage) error {
	f.mu.Lock()
	f.calls = append(f.calls, pushCall{userID: userID, msg: msg})
	f.mu.Unlock()
	f.done <- struct{}{}
	return f.err
}

func (f *fakePushNotifier) snapshot() []pushCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]pushCall(nil), f.calls...)
}

// newPushAdapter is an adapter whose client is Marie Dupont, with her phone
// and email known, a product, and a practice access holding a door code.
func newPushAdapter(push *fakePushNotifier, now time.Time) (*BookingNotificationAdapter, bookingPorts.BookingNotificationData) {
	data := makeBookingData()
	data.SlotStartTime = time.Date(2026, 11, 28, 9, 0, 0, 0, time.UTC) // sam. 28 nov. 10:00 in Paris
	data.SlotEndTime = data.SlotStartTime.Add(time.Hour)
	users := map[uuid.UUID]*UserInfo{
		data.ClientID:  {Email: "marie.dupont@example.com", FirstName: "Marie", LastName: "Dupont", Phone: "+33612345678"},
		data.PartnerID: {Email: "praticien@example.com", FirstName: "Paul", LastName: "Martin"},
	}
	products := map[string]*catalogDomain.ProductRes{data.ProductID.String(): {Name: "Ostéopathie"}}

	adapter := newTestAdapter(&spyEmailService{}, &spySMSService{}, users, nil, nil, products)
	access := &fakePracticeAccess{}
	access.set("4821B", "Sonnez deux fois.")
	WithPracticeAccessProvider(access)(adapter)
	WithPractitionerPush(push)(adapter)
	adapter.now = func() time.Time { return now }
	return adapter, data
}

func waitForPush(t *testing.T, push *fakePushNotifier) {
	t.Helper()
	select {
	case <-push.done:
	case <-time.After(2 * time.Second):
		t.Fatal("no practitioner notification was sent")
	}
}

func TestNewOnlineBooking_NotifiesItsPractitionerOnce(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC))
	data.Origin = bookingPorts.BookingOriginOnline

	if err := adapter.SendBookingConfirmation(context.Background(), data); err != nil {
		t.Fatalf("SendBookingConfirmation: %v", err)
	}
	waitForPush(t, push)
	time.Sleep(50 * time.Millisecond) // a second notification would show now

	calls := push.snapshot()
	if len(calls) != 1 {
		t.Fatalf("%d notifications, want 1", len(calls))
	}
	if calls[0].userID != data.PartnerID {
		t.Errorf("notified %s, want the booking's practitioner %s", calls[0].userID, data.PartnerID)
	}
	msg := calls[0].msg
	if msg.Title != "Nouvelle réservation" || msg.Body != "Marie D. · Ostéopathie · sam. 28 nov. 10:00" {
		t.Errorf("message = %+v", msg)
	}
	if msg.URL != notificationDomain.PractitionerBookingsURL {
		t.Errorf("url = %q, want Réservations for a booking on another day", msg.URL)
	}
}

func TestNewBooking_EmptyOriginIsOnline(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Date(2026, 11, 28, 7, 0, 0, 0, time.UTC))

	if err := adapter.SendBookingConfirmation(context.Background(), data); err != nil {
		t.Fatalf("SendBookingConfirmation: %v", err)
	}
	waitForPush(t, push)
	if got := push.snapshot()[0].msg.URL; got != notificationDomain.PractitionerTodayURL {
		t.Errorf("url = %q, want Aujourd'hui for a booking today", got)
	}
}

func TestPractitionerBooking_SendsNoNotification(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.Origin = bookingPorts.BookingOriginPractitioner

	if err := adapter.SendBookingConfirmation(context.Background(), data); err != nil {
		t.Fatalf("SendBookingConfirmation: %v", err)
	}
	if err := adapter.sendNewBookingPush(context.Background(), data); err != nil {
		t.Fatalf("sendNewBookingPush: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if calls := push.snapshot(); len(calls) != 0 {
		t.Fatalf("a practitioner booking sent %d notifications, want none", len(calls))
	}
}

func TestNewBookingPayload_NoPhoneEmailFullNameAmountOrAccess(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	// A guest booking carries every contact detail on the data itself.
	data.IsGuestBooking = true
	data.ClientFirstName = "Marie"
	data.ClientLastName = "Dupont"
	data.ClientName = "Marie Dupont"
	data.ClientEmail = "marie.dupont@example.com"
	data.ClientPhone = "+33612345678"
	data.GuestEmail = data.ClientEmail
	data.GuestPhone = data.ClientPhone
	data.TotalPriceCents = 6500
	data.AmountPaidOnlineCents = 2000

	if err := adapter.sendNewBookingPush(context.Background(), data); err != nil {
		t.Fatalf("sendNewBookingPush: %v", err)
	}
	calls := push.snapshot()
	if len(calls) != 1 {
		t.Fatalf("%d notifications, want 1", len(calls))
	}
	payload, _ := json.Marshal(calls[0].msg)
	if !strings.Contains(string(payload), "Marie D.") {
		t.Errorf("payload %s lacks the first name and initial", payload)
	}
	for _, secret := range []string{"Dupont", "+33612345678", "0612345678", "marie.dupont@example.com", "@", "4821B", "Sonnez", "65,00", "65 €", "6500", "20,00", "€"} {
		if strings.Contains(string(payload), secret) {
			t.Errorf("payload %s contains %q", payload, secret)
		}
	}
}

func TestNewBookingPush_FailureNeverFailsTheBooking(t *testing.T) {
	push := newFakePushNotifier()
	push.err = errors.New("push service down")
	adapter, data := newPushAdapter(push, time.Now())

	if err := adapter.SendBookingConfirmation(context.Background(), data); err != nil {
		t.Fatalf("a push failure must not reach the booking: %v", err)
	}
	waitForPush(t, push)
}

func TestNoPushNotifier_NothingSent(t *testing.T) {
	data := makeBookingData()
	adapter := newTestAdapter(&spyEmailService{}, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendNewBookingPush(context.Background(), data); err != nil {
		t.Fatalf("no notifier: %v", err)
	}
}

// --- Cancellations and payment problems (issue 045) ---

func TestClientCancellation_NotifiesItsPractitionerOnce(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC))
	data.CancelledBy = "client"

	if err := adapter.SendBookingCancellation(context.Background(), data); err != nil {
		t.Fatalf("SendBookingCancellation: %v", err)
	}
	waitForPush(t, push)
	time.Sleep(50 * time.Millisecond) // a second notification would show now

	calls := push.snapshot()
	if len(calls) != 1 {
		t.Fatalf("%d notifications, want 1", len(calls))
	}
	if calls[0].userID != data.PartnerID {
		t.Errorf("notified %s, want the booking's practitioner %s", calls[0].userID, data.PartnerID)
	}
	msg := calls[0].msg
	if msg.Title != "Annulation" || msg.Body != "Marie D. · Ostéopathie · sam. 28 nov. 10:00" {
		t.Errorf("message = %+v", msg)
	}
	if msg.URL != notificationDomain.PractitionerBookingsURL {
		t.Errorf("url = %q, want Réservations", msg.URL)
	}
}

func TestPractitionerCancellation_SendsNoNotification(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.CancelledBy = "practitioner"

	if err := adapter.SendBookingCancellation(context.Background(), data); err != nil {
		t.Fatalf("SendBookingCancellation: %v", err)
	}
	if err := adapter.sendCancellationPush(context.Background(), data); err != nil {
		t.Fatalf("sendCancellationPush: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if calls := push.snapshot(); len(calls) != 0 {
		t.Fatalf("a practitioner cancellation sent %d notifications, want none", len(calls))
	}
}

func TestUnattributedCancellation_SendsNoNotification(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.CancelledBy = "" // cancelled before the actor was recorded

	if err := adapter.sendCancellationPush(context.Background(), data); err != nil {
		t.Fatalf("sendCancellationPush: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if calls := push.snapshot(); len(calls) != 0 {
		t.Fatalf("%d notifications, want none", len(calls))
	}
}

func TestPaymentDispute_NotifiesPractitionerOnce(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.DisputeReason = "product_unreceived_service"

	if err := adapter.SendPaymentDisputeAlert(context.Background(), data); err != nil {
		t.Fatalf("SendPaymentDisputeAlert: %v", err)
	}
	waitForPush(t, push)
	time.Sleep(50 * time.Millisecond)

	calls := push.snapshot()
	if len(calls) != 1 {
		t.Fatalf("%d notifications, want 1", len(calls))
	}
	if calls[0].userID != data.PartnerID {
		t.Errorf("notified %s, want the practice's practitioner %s", calls[0].userID, data.PartnerID)
	}
	msg := calls[0].msg
	if msg.Title != "Litige de paiement" || msg.Body != "Marie D." {
		t.Errorf("message = %+v", msg)
	}
	if msg.URL != notificationDomain.PractitionerPaymentsURL {
		t.Errorf("url = %q, want Comptabilité", msg.URL)
	}
}

func TestLatePaymentRefund_NotifiesPractitionerOnce(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.RefundedAmountCents = 6500

	if err := adapter.SendLatePaymentRefund(context.Background(), data); err != nil {
		t.Fatalf("SendLatePaymentRefund: %v", err)
	}
	waitForPush(t, push)
	time.Sleep(50 * time.Millisecond)

	calls := push.snapshot()
	if len(calls) != 1 {
		t.Fatalf("%d notifications, want 1", len(calls))
	}
	msg := calls[0].msg
	if msg.Title != "Paiement tardif remboursé" || msg.Body != "Marie D." {
		t.Errorf("message = %+v", msg)
	}
	if msg.URL != notificationDomain.PractitionerPaymentsURL {
		t.Errorf("url = %q, want Comptabilité", msg.URL)
	}
}

func TestPaymentProblemPayload_NoPhoneEmailFullNameAmountOrAccess(t *testing.T) {
	push := newFakePushNotifier()
	adapter, data := newPushAdapter(push, time.Now())
	data.IsGuestBooking = true
	data.ClientFirstName = "Marie"
	data.ClientLastName = "Dupont"
	data.ClientName = "Marie Dupont"
	data.ClientEmail = "marie.dupont@example.com"
	data.ClientPhone = "+33612345678"
	data.TotalPriceCents = 6500
	data.RefundedAmountCents = 6500
	data.DisputeReason = "product_unreceived_service"
	data.CancelledBy = "client"

	for name, send := range map[string]func() error{
		"cancellation": func() error { return adapter.sendCancellationPush(context.Background(), data) },
		"dispute":      func() error { return adapter.sendDisputePush(context.Background(), data) },
		"late refund":  func() error { return adapter.sendLatePaymentRefundPush(context.Background(), data) },
	} {
		if err := send(); err != nil {
			t.Fatalf("%s push: %v", name, err)
		}
	}

	calls := push.snapshot()
	if len(calls) != 3 {
		t.Fatalf("%d notifications, want 3", len(calls))
	}
	for _, call := range calls {
		payload, _ := json.Marshal(call.msg)
		if !strings.Contains(string(payload), "Marie D.") {
			t.Errorf("%s: payload %s lacks the first name and initial", call.msg.Title, payload)
		}
		for _, secret := range []string{"Dupont", "+33612345678", "0612345678", "marie.dupont@example.com", "@", "4821B", "Sonnez", "6500", "65,00", "65 €", "€", "product_unreceived"} {
			if strings.Contains(string(payload), secret) {
				t.Errorf("%s: payload %s contains %q", call.msg.Title, payload, secret)
			}
		}
	}
}
