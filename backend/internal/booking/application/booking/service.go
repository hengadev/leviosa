package booking

import (
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"
	catalogPorts "github.com/Leviosa-care/leviosa/backend/internal/catalog/ports"
	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
	"github.com/hengadev/encx"
)

type BookingService struct {
	bookingRepo         ports.BookingRepository
	availabilityRepo    ports.AvailabilityRepository
	paymentService      ports.PaymentService
	productService      catalogPorts.PublicProductService
	priceService        catalogPorts.PublicPriceService
	notificationService ports.BookingNotificationService
	roomService         ports.RoomService
	authUserClient      ports.AuthUserClient
	crypto              encx.CryptoService
	tokenSecret         []byte
	audit               audit.Logger
}

// New creates a new instance of the booking service
func New(
	bookingRepo ports.BookingRepository,
	availabilityRepo ports.AvailabilityRepository,
	paymentService ports.PaymentService,
	productService catalogPorts.PublicProductService,
	priceService catalogPorts.PublicPriceService,
	notificationService ports.BookingNotificationService,
	crypto encx.CryptoService,
	opts ...Option,
) ports.BookingService {
	svc := &BookingService{
		bookingRepo:         bookingRepo,
		availabilityRepo:    availabilityRepo,
		paymentService:      paymentService,
		productService:      productService,
		priceService:        priceService,
		notificationService: notificationService,
		crypto:              crypto,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// Option configures optional BookingService dependencies.
type Option func(*BookingService)

// WithRoomService injects the room service dependency.
func WithRoomService(rs ports.RoomService) Option {
	return func(s *BookingService) { s.roomService = rs }
}

// WithAuthUserClient injects the auth-user client dependency.
func WithAuthUserClient(c ports.AuthUserClient) Option {
	return func(s *BookingService) { s.authUserClient = c }
}

// WithTokenSecret sets the HMAC secret used to sign booking tokens.
func WithTokenSecret(secret []byte) Option {
	return func(s *BookingService) { s.tokenSecret = secret }
}

// WithAuditLogger injects the audit-log port used to record sensitive
// booking actions (refunds, admin overrides). If not provided, those use
// cases simply skip writing an audit entry.
func WithAuditLogger(l audit.Logger) Option {
	return func(s *BookingService) { s.audit = l }
}
