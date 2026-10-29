package notification

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	bookingPorts "github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"

	"github.com/google/uuid"
)

// ---- Spy SMS service ----

type spySMSService struct {
	mu       sync.Mutex
	messages []notificationDomain.GenericSMSRequest
}

func (s *spySMSService) SendOTP(ctx context.Context, req notificationDomain.OTPSMSRequest) error {
	return nil
}
func (s *spySMSService) SendSMS(ctx context.Context, req notificationDomain.GenericSMSRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, req)
	return nil
}

// ---- Spy email service ----

type spyEmailService struct {
	mu                   sync.Mutex
	bookingConfirmations []notificationDomain.BookingConfirmationRequest
	clientCancellations  []notificationDomain.BookingClientCancellationRequest
	partnerCancellations []notificationDomain.BookingPartnerCancellationRequest
	bookingReminders     []notificationDomain.BookingReminderRequest
	paymentNotifications []notificationDomain.PaymentNotificationRequest
	disputeAlerts        []notificationDomain.PaymentDisputeAlertRequest
	partnerNotices       []notificationDomain.BookingPartnerNoticeRequest
	latePaymentRefunds   []notificationDomain.LatePaymentRefundRequest
}

func (s *spyEmailService) SendLatePaymentRefundEmail(ctx context.Context, req notificationDomain.LatePaymentRefundRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latePaymentRefunds = append(s.latePaymentRefunds, req)
	return nil
}

func (s *spyEmailService) SendOTPEmail(ctx context.Context, req notificationDomain.OTPEmailRequest) error {
	return nil
}
func (s *spyEmailService) SendWelcomeEmail(ctx context.Context, req notificationDomain.WelcomeEmailRequest) error {
	return nil
}
func (s *spyEmailService) SendVerifyEmailEmail(ctx context.Context, req notificationDomain.VerifyEmailRequest) error {
	return nil
}
func (s *spyEmailService) SendEventNotificationEmail(ctx context.Context, req notificationDomain.EventNotificationRequest) error {
	return nil
}
func (s *spyEmailService) SendPaymentNotificationEmail(ctx context.Context, req notificationDomain.PaymentNotificationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paymentNotifications = append(s.paymentNotifications, req)
	return nil
}
func (s *spyEmailService) SendPaymentFailedEmail(ctx context.Context, req notificationDomain.PaymentNotificationRequest) error {
	return nil
}
func (s *spyEmailService) SendBookingConfirmationEmail(ctx context.Context, req notificationDomain.BookingConfirmationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookingConfirmations = append(s.bookingConfirmations, req)
	return nil
}
func (s *spyEmailService) SendBookingPartnerNoticeEmail(ctx context.Context, req notificationDomain.BookingPartnerNoticeRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.partnerNotices = append(s.partnerNotices, req)
	return nil
}
func (s *spyEmailService) SendBookingClientCancellationEmail(ctx context.Context, req notificationDomain.BookingClientCancellationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientCancellations = append(s.clientCancellations, req)
	return nil
}
func (s *spyEmailService) SendBookingPartnerCancellationEmail(ctx context.Context, req notificationDomain.BookingPartnerCancellationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.partnerCancellations = append(s.partnerCancellations, req)
	return nil
}
func (s *spyEmailService) SendBookingReminderEmail(ctx context.Context, req notificationDomain.BookingReminderRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookingReminders = append(s.bookingReminders, req)
	return nil
}
func (s *spyEmailService) SendContactFormEmail(ctx context.Context, req notificationDomain.ContactFormRequest) error {
	return nil
}
func (s *spyEmailService) SendPaymentDisputeAlertEmail(ctx context.Context, req notificationDomain.PaymentDisputeAlertRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disputeAlerts = append(s.disputeAlerts, req)
	return nil
}

// ---- Stub fetchers ----

type stubUserFetcher struct {
	users map[uuid.UUID]*UserInfo
}

func (f *stubUserFetcher) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	u, ok := f.users[userID]
	if !ok {
		return nil, fmt.Errorf("user %s not found", userID)
	}
	return u, nil
}

type stubRoomFetcher struct {
	rooms map[uuid.UUID]*RoomInfo
}

func (f *stubRoomFetcher) GetRoom(ctx context.Context, roomID uuid.UUID) (*RoomInfo, error) {
	r, ok := f.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room %s not found", roomID)
	}
	return r, nil
}

type stubBuildingFetcher struct {
	buildings map[uuid.UUID]*BuildingInfo
}

func (f *stubBuildingFetcher) GetBuilding(ctx context.Context, buildingID uuid.UUID) (*BuildingInfo, error) {
	b, ok := f.buildings[buildingID]
	if !ok {
		return nil, fmt.Errorf("building %s not found", buildingID)
	}
	return b, nil
}

type stubProductFetcher struct {
	products map[string]*catalogDomain.ProductRes
}

func (f *stubProductFetcher) GetProductByID(ctx context.Context, id string) (*catalogDomain.ProductRes, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, fmt.Errorf("product %s not found", id)
	}
	return p, nil
}
func (f *stubProductFetcher) GetAllPublishedProducts(ctx context.Context) ([]*catalogDomain.ProductRes, error) {
	return nil, nil
}
func (f *stubProductFetcher) GetAllProducts(ctx context.Context) ([]*catalogDomain.ProductRes, error) {
	return nil, nil
}

// ---- Stub settings provider ----

type stubSettingsProvider struct {
	companyEmail   string
	companyAddress string
}

func (s *stubSettingsProvider) GetCompanyEmail(ctx context.Context) (string, error) {
	return s.companyEmail, nil
}
func (s *stubSettingsProvider) GetCompanyLegalAddress(ctx context.Context) (string, error) {
	return s.companyAddress, nil
}
func (s *stubSettingsProvider) GetCompanyInstagram(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyLogo(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyTelephone(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyName(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) InvalidateCache(key string) {}
func (s *stubSettingsProvider) InvalidateAllCache()        {}

// ---- Helpers ----

func newTestAdapter(
	spy *spyEmailService,
	smsSpy *spySMSService,
	users map[uuid.UUID]*UserInfo,
	rooms map[uuid.UUID]*RoomInfo,
	buildings map[uuid.UUID]*BuildingInfo,
	products map[string]*catalogDomain.ProductRes,
) *BookingNotificationAdapter {
	return NewBookingNotificationAdapter(
		spy,
		smsSpy,
		"https://app.leviosa.com",
		&stubUserFetcher{users: users},
		&stubRoomFetcher{rooms: rooms},
		&stubBuildingFetcher{buildings: buildings},
		&stubProductFetcher{products: products},
		&stubSettingsProvider{companyEmail: "admin@leviosa.test"},
	)
}

func makeBookingData() bookingPorts.BookingNotificationData {
	return bookingPorts.BookingNotificationData{
		BookingID:       uuid.New(),
		ClientID:        uuid.New(),
		PartnerID:       uuid.New(),
		RoomID:          uuid.New(),
		ProductID:       uuid.New(),
		SlotStartTime:   time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		SlotEndTime:     time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC),
		TotalPriceCents: 5000,
		Currency:        "EUR",
	}
}

// ---- Tests ----

func TestSendBookingConfirmation_MapsFieldsCorrectly(t *testing.T) {
	data := makeBookingData()
	data.TotalPriceCents = 6500
	data.AmountPaidOnlineCents = 2000
	data.Locale = "en"

	users := map[uuid.UUID]*UserInfo{
		data.ClientID:  {Email: "client@example.com", FirstName: "Alice", LastName: "Smith", Phone: "+33612345678"},
		data.PartnerID: {Email: "partner@example.com", FirstName: "Bob", LastName: "Jones"},
	}
	rooms := map[uuid.UUID]*RoomInfo{
		data.RoomID: {Name: "Room A", BuildingID: uuid.New()},
	}
	buildingID := rooms[data.RoomID].BuildingID
	buildings := map[uuid.UUID]*BuildingInfo{
		buildingID: {Name: "Main Building", Address: "123 Rue de Paris, 75001 Paris"},
	}
	products := map[string]*catalogDomain.ProductRes{
		data.ProductID.String(): {Name: "Massage Therapy"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, rooms, buildings, products)

	// Call the internal method directly to verify field mapping without goroutine races.
	if err := adapter.sendBookingConfirmationEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.bookingConfirmations) != 1 {
		t.Fatalf("expected 1 booking confirmation, got %d", len(spy.bookingConfirmations))
	}
	got := spy.bookingConfirmations[0]
	assertEqual(t, "ToEmail", "client@example.com", got.ToEmail)
	assertEqual(t, "ToFirstName", "Alice", got.ToFirstName)
	assertEqual(t, "ToLastName", "Smith", got.ToLastName)
	assertEqual(t, "Locale", "en", string(got.Locale))
	assertEqual(t, "ProductName", "Massage Therapy", got.ProductName)
	assertEqual(t, "Address (building, no company address)", "123 Rue de Paris, 75001 Paris", got.Address)
	assertEqual(t, "PartnerName", "Bob Jones", got.PartnerName)
	assertEqual(t, "BookingID", data.BookingID.String(), got.BookingID)
	assertEqual(t, "Currency", "EUR", got.Currency)
	if !got.StartTime.Equal(data.SlotStartTime) || !got.EndTime.Equal(data.SlotEndTime) {
		t.Errorf("times = %v–%v, want %v–%v", got.StartTime, got.EndTime, data.SlotStartTime, data.SlotEndTime)
	}
	if got.AmountPaidOnlineCents != 2000 || got.BalanceDueCents != 4500 {
		t.Errorf("paid online %d / balance %d, want 2000 / 4500", got.AmountPaidOnlineCents, got.BalanceDueCents)
	}

	if len(spy.partnerNotices) != 1 {
		t.Fatalf("expected 1 partner notice, got %d", len(spy.partnerNotices))
	}
	notice := spy.partnerNotices[0]
	assertEqual(t, "partner ToEmail", "partner@example.com", notice.ToEmail)
	assertEqual(t, "partner ToFirstName", "Bob", notice.ToFirstName)
	assertEqual(t, "partner ClientName", "Alice Smith", notice.ClientName)
	assertEqual(t, "partner ClientEmail", "client@example.com", notice.ClientEmail)
	assertEqual(t, "partner ClientPhone", "+33612345678", notice.ClientPhone)
	assertEqual(t, "partner ProductName", "Massage Therapy", notice.ProductName)

	if len(smsSpy.messages) != 0 {
		t.Errorf("booking notifications send no SMS, got %d", len(smsSpy.messages))
	}
}

func TestSendBookingConfirmation_PracticeAddressFromSettings(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.ClientName = "Alice Smith"
	data.PartnerEmail = "partner@example.com"
	data.PartnerName = "Bob Jones"
	data.Address = "Building address"

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	adapter.settingsProvider = &stubSettingsProvider{companyAddress: "12 rue de la Paix, 75002 Paris"}

	if err := adapter.sendBookingConfirmationEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Address", "12 rue de la Paix, 75002 Paris", spy.bookingConfirmations[0].Address)
}

func TestSendBookingConfirmation_PaidInFull_NoBalance(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.AmountPaidOnlineCents = data.TotalPriceCents

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendBookingConfirmationEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := spy.bookingConfirmations[0].BalanceDueCents; got != 0 {
		t.Errorf("balance = %d, want 0", got)
	}
	assertEqual(t, "Locale (missing reads as French)", "fr", string(spy.bookingConfirmations[0].Locale))
}

func TestSendPaymentConfirmation_MapsFieldsCorrectly(t *testing.T) {
	data := makeBookingData()
	data.ProductName = "Pre-populated Product" // already known

	users := map[uuid.UUID]*UserInfo{
		data.ClientID:  {Email: "client@example.com", FirstName: "Carol", LastName: "White"},
		data.PartnerID: {Email: "partner@example.com", FirstName: "Dave", LastName: "Brown"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, nil, nil, nil)

	err := adapter.sendPaymentConfirmationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.paymentNotifications) != 1 {
		t.Fatalf("expected 1 payment notification, got %d", len(spy.paymentNotifications))
	}

	got := spy.paymentNotifications[0]

	assertEqual(t, "ToEmail", "client@example.com", got.ToEmail)
	assertEqual(t, "ToFirstName", "Carol", got.ToFirstName)
	assertEqual(t, "ToLastName", "White", got.ToLastName)
	assertEqual(t, "Amount", "€50.00", got.Amount)
	assertEqual(t, "Product", "Pre-populated Product", got.Product)
}

func TestSendBookingConfirmation_GuestBooking(t *testing.T) {
	data := makeBookingData()
	data.IsGuestBooking = true
	data.GuestEmail = "guest@example.com"
	data.ClientEmail = "guest@example.com"
	data.ClientName = "Jane Marie Doe"
	data.ClientFirstName = "Jane Marie"
	data.ClientLastName = "Doe"
	data.ClientID = uuid.Nil

	// Only partner needs to be fetched
	users := map[uuid.UUID]*UserInfo{
		data.PartnerID: {Email: "partner@example.com", FirstName: "Bob", LastName: "Jones"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, nil, nil, nil)

	err := adapter.sendBookingConfirmationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.bookingConfirmations) != 1 {
		t.Fatalf("expected 1 booking confirmation, got %d", len(spy.bookingConfirmations))
	}

	got := spy.bookingConfirmations[0]
	assertEqual(t, "ToEmail", "guest@example.com", got.ToEmail)
	assertEqual(t, "ToFirstName", "Jane Marie", got.ToFirstName)
	assertEqual(t, "ToLastName", "Doe", got.ToLastName)

	if len(spy.partnerNotices) != 1 {
		t.Fatalf("expected 1 partner notice, got %d", len(spy.partnerNotices))
	}
	assertEqual(t, "partner ClientName", "Jane Marie Doe", spy.partnerNotices[0].ClientName)
	assertEqual(t, "partner ClientEmail", "guest@example.com", spy.partnerNotices[0].ClientEmail)
}

func TestSendBookingConfirmation_MissingClientEmail_NoEmailSent(t *testing.T) {
	data := makeBookingData()
	// No users, no guest email
	emailSpy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(emailSpy, smsSpy, nil, nil, nil, nil)

	err := adapter.sendBookingConfirmationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error (fire-and-forget): %v", err)
	}

	if len(emailSpy.bookingConfirmations) != 0 {
		t.Fatalf("expected 0 confirmation emails when client email is missing, got %d", len(emailSpy.bookingConfirmations))
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		cents    int
		currency string
		want     string
	}{
		{5000, "EUR", "€50.00"},
		{1000, "USD", "$10.00"},
		{2500, "GBP", "£25.00"},
		{999, "JPY", "9.99 JPY"},
	}
	for _, tt := range tests {
		got := formatAmount(tt.cents, tt.currency)
		if got != tt.want {
			t.Errorf("formatAmount(%d, %q) = %q, want %q", tt.cents, tt.currency, got, tt.want)
		}
	}
}

func TestFirstNameLastName(t *testing.T) {
	assertEqual(t, "firstName", "Alice", firstName("Alice Smith"))
	assertEqual(t, "firstName", "Alice", firstName("Alice"))
	assertEqual(t, "lastName", "Smith", lastName("Alice Smith"))
	assertEqual(t, "lastName", "", lastName("Alice"))
}

func TestSendBookingCancellation_SendsToClientAndPartner(t *testing.T) {
	data := makeBookingData()
	data.Locale = "en"
	data.CancelledBy = "client"
	data.AmountPaidOnlineCents = 3000
	data.RefundedAmountCents = 1500
	cancelledAt := time.Date(2026, 6, 14, 9, 0, 0, 0, time.UTC)
	data.CancelledAt = &cancelledAt

	users := map[uuid.UUID]*UserInfo{
		data.ClientID:  {Email: "client@example.com", FirstName: "Alice", LastName: "Smith"},
		data.PartnerID: {Email: "partner@example.com", FirstName: "Bob", LastName: "Jones"},
	}
	products := map[string]*catalogDomain.ProductRes{
		data.ProductID.String(): {Name: "Massage Therapy"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, nil, nil, products)

	err := adapter.sendBookingCancellationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.clientCancellations) != 1 || len(spy.partnerCancellations) != 1 {
		t.Fatalf("expected 1 client and 1 partner cancellation email, got %d and %d", len(spy.clientCancellations), len(spy.partnerCancellations))
	}

	clientEmail := spy.clientCancellations[0]
	assertEqual(t, "client ToEmail", "client@example.com", clientEmail.ToEmail)
	assertEqual(t, "client ToFirstName", "Alice", clientEmail.ToFirstName)
	assertEqual(t, "client ToLastName", "Smith", clientEmail.ToLastName)
	assertEqual(t, "client Locale", string(notificationDomain.LocaleEN), string(clientEmail.Locale))
	assertEqual(t, "client BookingID", data.BookingID.String(), clientEmail.BookingID)
	assertEqual(t, "client ProductName", "Massage Therapy", clientEmail.ProductName)
	assertEqual(t, "client CancelledBy", "client", clientEmail.CancelledBy)
	if clientEmail.AmountPaidOnlineCents != 3000 || clientEmail.RefundedAmountCents != 1500 {
		t.Errorf("client amounts = paid %d, refunded %d; want 3000, 1500", clientEmail.AmountPaidOnlineCents, clientEmail.RefundedAmountCents)
	}
	if !clientEmail.StartTime.Equal(data.SlotStartTime) {
		t.Errorf("client StartTime = %v, want %v", clientEmail.StartTime, data.SlotStartTime)
	}

	partnerEmail := spy.partnerCancellations[0]
	assertEqual(t, "partner ToEmail", "partner@example.com", partnerEmail.ToEmail)
	assertEqual(t, "partner ToFirstName", "Bob", partnerEmail.ToFirstName)
	assertEqual(t, "partner BookingID", data.BookingID.String(), partnerEmail.BookingID)
	assertEqual(t, "partner ProductName", "Massage Therapy", partnerEmail.ProductName)
	assertEqual(t, "partner ClientName", "Alice Smith", partnerEmail.ClientName)
	if partnerEmail.RefundedAmountCents != 1500 {
		t.Errorf("partner refunded = %d, want 1500", partnerEmail.RefundedAmountCents)
	}

	if len(smsSpy.messages) != 0 {
		t.Errorf("expected no SMS for a booking cancellation, got %d", len(smsSpy.messages))
	}
}

func TestSendBookingCancellation_GuestBooking(t *testing.T) {
	data := makeBookingData()
	data.IsGuestBooking = true
	data.GuestEmail = "guest@example.com"
	data.ClientEmail = "guest@example.com"
	data.ClientName = "Jane Doe"
	data.ClientFirstName = "Jane"
	data.ClientLastName = "Doe"
	data.ClientID = uuid.Nil

	users := map[uuid.UUID]*UserInfo{
		data.PartnerID: {Email: "partner@example.com", FirstName: "Bob", LastName: "Jones"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, nil, nil, nil)

	err := adapter.sendBookingCancellationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.clientCancellations) != 1 || len(spy.partnerCancellations) != 1 {
		t.Fatalf("expected 1 client and 1 partner cancellation email, got %d and %d", len(spy.clientCancellations), len(spy.partnerCancellations))
	}

	assertEqual(t, "guest client ToEmail", "guest@example.com", spy.clientCancellations[0].ToEmail)
	assertEqual(t, "guest client ToFirstName", "Jane", spy.clientCancellations[0].ToFirstName)
	assertEqual(t, "guest client Locale", string(notificationDomain.LocaleFR), string(spy.clientCancellations[0].Locale))
	assertEqual(t, "partner ToEmail", "partner@example.com", spy.partnerCancellations[0].ToEmail)
}

func TestSendBookingCancellation_MissingClientEmail_StillSendsToPartner(t *testing.T) {
	data := makeBookingData()

	users := map[uuid.UUID]*UserInfo{
		data.PartnerID: {Email: "partner@example.com", FirstName: "Bob", LastName: "Jones"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, nil, nil, nil)

	err := adapter.sendBookingCancellationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Partner should still receive the cancellation email even if client email is missing
	if len(spy.clientCancellations) != 0 || len(spy.partnerCancellations) != 1 {
		t.Fatalf("expected the partner email only, got %d client and %d partner", len(spy.clientCancellations), len(spy.partnerCancellations))
	}
	assertEqual(t, "partner ToEmail", "partner@example.com", spy.partnerCancellations[0].ToEmail)
}

func TestSendBookingCancellation_MissingPartnerEmail_StillSendsToClient(t *testing.T) {
	data := makeBookingData()

	users := map[uuid.UUID]*UserInfo{
		data.ClientID: {Email: "client@example.com", FirstName: "Alice", LastName: "Smith"},
	}

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, users, nil, nil, nil)

	err := adapter.sendBookingCancellationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Client should still receive the cancellation email even if partner email is missing
	if len(spy.clientCancellations) != 1 || len(spy.partnerCancellations) != 0 {
		t.Fatalf("expected the client email only, got %d client and %d partner", len(spy.clientCancellations), len(spy.partnerCancellations))
	}
	assertEqual(t, "client ToEmail", "client@example.com", spy.clientCancellations[0].ToEmail)
}

func TestSendBookingCancellation_BothEmailsMissing_LogsAndReturnsNil(t *testing.T) {
	data := makeBookingData()

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, nil, nil, nil, nil)

	err := adapter.sendBookingCancellationEmail(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.clientCancellations)+len(spy.partnerCancellations) != 0 {
		t.Fatalf("expected 0 cancellation emails, got %d", len(spy.clientCancellations)+len(spy.partnerCancellations))
	}
}

func TestSendBookingCancellation_FireAndForget(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.PartnerEmail = "partner@example.com"
	data.ClientName = "Test Client"
	data.PartnerName = "Test Partner"
	data.CancellationReason = "Test"

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, nil, nil, nil, nil)

	// SendBookingCancellation returns nil immediately (fire-and-forget)
	err := adapter.SendBookingCancellation(context.Background(), data)
	if err != nil {
		t.Fatalf("expected nil error from fire-and-forget, got: %v", err)
	}

	// Wait for goroutine to execute
	time.Sleep(100 * time.Millisecond)
}

func TestSendBookingConfirmation_FireAndForget(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.ClientName = "Test Client"

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, nil, nil, nil, nil)

	// SendBookingConfirmation returns nil immediately (fire-and-forget)
	err := adapter.SendBookingConfirmation(context.Background(), data)
	if err != nil {
		t.Fatalf("expected nil error from fire-and-forget, got: %v", err)
	}

	// Wait for goroutine to execute
	time.Sleep(100 * time.Millisecond)

	// The goroutine will fail to send (no client in fetcher) but that's OK,
	// the important thing is SendBookingConfirmation itself returned nil.
}

func TestSendBookingConfirmation_ManageURLWithToken(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.Token = "abc+123/token="

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendBookingConfirmationEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "ManageURL", "https://app.leviosa.com/bookings?token=abc%2B123%2Ftoken%3D", spy.bookingConfirmations[0].ManageURL)
}

func TestSendBookingConfirmation_NoTokenNoManageURL(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.Token = ""

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendBookingConfirmationEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "ManageURL", "", spy.bookingConfirmations[0].ManageURL)
}

// SMS is disabled for booking notifications (PRD v1): with a phone number on
// file and an SMS client wired, no booking event sends an SMS.
func TestBookingNotifications_NeverSendSMS(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.ClientName = "Alice Smith"
	data.ClientPhone = "+33612345678"
	data.PartnerEmail = "partner@example.com"
	data.PartnerName = "Bob Jones"
	data.ProductName = "Massage Therapy"
	data.Token = "abc123token"
	data.CancellationReason = "Test"

	for name, send := range map[string]func(*BookingNotificationAdapter) error{
		"confirmation": func(a *BookingNotificationAdapter) error {
			return a.sendBookingConfirmationEmail(context.Background(), data)
		},
		"cancellation": func(a *BookingNotificationAdapter) error {
			return a.sendBookingCancellationEmail(context.Background(), data)
		},
		"reminder": func(a *BookingNotificationAdapter) error {
			return a.sendBookingReminderEmail(context.Background(), data)
		},
	} {
		t.Run(name, func(t *testing.T) {
			smsSpy := &spySMSService{}
			adapter := newTestAdapter(&spyEmailService{}, smsSpy, nil, nil, nil, nil)
			if err := send(adapter); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(smsSpy.messages) != 0 {
				t.Errorf("expected no SMS for a booking %s, got %d", name, len(smsSpy.messages))
			}
		})
	}
}

func TestSendBookingReminder_SendsEmail(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.ClientName = "Alice Smith"

	emailSpy := &spyEmailService{}
	adapter := newTestAdapter(emailSpy, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendBookingReminderEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(emailSpy.bookingReminders) != 1 {
		t.Fatalf("expected 1 reminder email, got %d", len(emailSpy.bookingReminders))
	}
	assertEqual(t, "ToEmail", "client@example.com", emailSpy.bookingReminders[0].ToEmail)
}

func TestSendBookingReminder_LocaleAddressAndManageLink(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.ClientName = "Alice Smith"
	data.PartnerName = "Bob Jones"
	data.ProductName = "Massage Therapy"
	data.Address = "Building address"
	data.Locale = "en"
	data.Token = "abc123token"

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	adapter.settingsProvider = &stubSettingsProvider{companyAddress: "12 rue de la Paix, 75002 Paris"}

	if err := adapter.sendBookingReminderEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.bookingReminders) != 1 {
		t.Fatalf("expected 1 reminder email, got %d", len(spy.bookingReminders))
	}
	req := spy.bookingReminders[0]
	assertEqual(t, "Locale", "en", string(req.Locale))
	assertEqual(t, "ToFirstName", "Alice", req.ToFirstName)
	assertEqual(t, "PartnerName", "Bob Jones", req.PartnerName)
	assertEqual(t, "ProductName", "Massage Therapy", req.ProductName)
	assertEqual(t, "Address", "12 rue de la Paix, 75002 Paris", req.Address)
	assertEqual(t, "ManageURL", "https://app.leviosa.com/bookings?token=abc123token", req.ManageURL)
	if !req.StartTime.Equal(data.SlotStartTime) || !req.EndTime.Equal(data.SlotEndTime) {
		t.Errorf("reminder times %v–%v, want %v–%v", req.StartTime, req.EndTime, data.SlotStartTime, data.SlotEndTime)
	}
}

func TestSendBookingReminder_MissingLocaleIsFrench(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.Locale = ""

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendBookingReminderEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Locale", "fr", string(spy.bookingReminders[0].Locale))
}

func TestSendBookingReminder_FireAndForget(t *testing.T) {
	data := makeBookingData()
	data.ClientEmail = "client@example.com"
	data.ClientName = "Test Client"
	data.ClientPhone = "+33612345678"

	spy := &spyEmailService{}
	smsSpy := &spySMSService{}
	adapter := newTestAdapter(spy, smsSpy, nil, nil, nil, nil)

	// SendBookingReminder returns nil immediately (fire-and-forget)
	err := adapter.SendBookingReminder(context.Background(), data)
	if err != nil {
		t.Fatalf("expected nil error from fire-and-forget, got: %v", err)
	}

	// Wait for goroutine to execute
	time.Sleep(100 * time.Millisecond)
}

func assertEqual(t *testing.T, field, want, got string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

func TestSendLatePaymentRefund_GuestInEnglish(t *testing.T) {
	data := makeBookingData()
	data.IsGuestBooking = true
	data.ClientID = uuid.Nil
	data.ClientEmail = "guest@example.com"
	data.ClientFirstName = "Jane"
	data.ClientLastName = "Doe"
	data.Locale = "en"
	data.RefundedAmountCents = 2000
	data.ProductName = "Osteopathy"

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)

	if err := adapter.sendLatePaymentRefundEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(spy.latePaymentRefunds) != 1 {
		t.Fatalf("expected 1 refund email, got %d", len(spy.latePaymentRefunds))
	}
	got := spy.latePaymentRefunds[0]
	assertEqual(t, "ToEmail", "guest@example.com", got.ToEmail)
	assertEqual(t, "ToFirstName", "Jane", got.ToFirstName)
	assertEqual(t, "Locale", string(notificationDomain.LocaleEN), string(got.Locale))
	assertEqual(t, "BookAgainURL", "https://app.leviosa.com/en", got.BookAgainURL)
	assertEqual(t, "Currency", "EUR", got.Currency)
	if got.RefundedAmountCents != 2000 {
		t.Errorf("RefundedAmountCents = %d, want 2000", got.RefundedAmountCents)
	}
	if len(spy.partnerNotices) != 0 || len(spy.bookingConfirmations) != 0 {
		t.Error("a late-payment refund sends neither a confirmation nor a partner notice")
	}
}

func TestSendLatePaymentRefund_FrenchByDefault(t *testing.T) {
	data := makeBookingData()
	data.IsGuestBooking = true
	data.ClientEmail = "guest@example.com"
	data.RefundedAmountCents = 2000

	spy := &spyEmailService{}
	adapter := newTestAdapter(spy, &spySMSService{}, nil, nil, nil, nil)
	if err := adapter.sendLatePaymentRefundEmail(context.Background(), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.latePaymentRefunds) != 1 {
		t.Fatalf("expected 1 refund email, got %d", len(spy.latePaymentRefunds))
	}
	assertEqual(t, "Locale", string(notificationDomain.LocaleFR), string(spy.latePaymentRefunds[0].Locale))
	assertEqual(t, "BookAgainURL", "https://app.leviosa.com/", spy.latePaymentRefunds[0].BookAgainURL)
}
