package booking

import (
	"context"
	"testing"
	"time"

	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

// stubProductService is a minimal fake of catalogPorts.PublicProductService for
// unit-testing GetBookingFinancialStatus without a catalog dependency.
type stubProductService struct {
	product *catalogDomain.ProductRes
	err     error
}

func (s stubProductService) GetProductByID(_ context.Context, _ string) (*catalogDomain.ProductRes, error) {
	return s.product, s.err
}
func (s stubProductService) GetAllPublishedProducts(_ context.Context) ([]*catalogDomain.ProductRes, error) {
	return nil, nil
}
func (s stubProductService) GetAllProducts(_ context.Context) ([]*catalogDomain.ProductRes, error) {
	return nil, nil
}

func newTestBooking() *domain.Booking {
	return &domain.Booking{
		ID:              uuid.New(),
		ProductID:       uuid.New(),
		SlotStartTime:   time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
		Status:          domain.BookingStatusConfirmed,
		PaymentStatus:   domain.PaymentStatusPaid,
		TotalPriceCents: 5000,
		UpdatedAt:       time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC),
	}
}

func TestGetBookingFinancialStatus_CancellationDeadline(t *testing.T) {
	svc := &BookingService{
		productService: stubProductService{
			product: &catalogDomain.ProductRes{CancellationHours: 24},
		},
	}

	booking := newTestBooking()
	status, err := svc.GetBookingFinancialStatus(context.Background(), booking)
	require.NoError(t, err)

	require.NotNil(t, status.CancellationDeadline)
	assert.Equal(t, booking.SlotStartTime.Add(-24*time.Hour), *status.CancellationDeadline)
	require.NotNil(t, status.CancellationHours)
	assert.Equal(t, 24, *status.CancellationHours)
	assert.Nil(t, status.RefundAmountCents)
	assert.Nil(t, status.RefundedAt)
}

func TestGetBookingFinancialStatus_NoDeadlineWhenNotCancellable(t *testing.T) {
	svc := &BookingService{
		productService: stubProductService{
			product: &catalogDomain.ProductRes{CancellationHours: 24},
		},
	}

	booking := newTestBooking()
	booking.Status = domain.BookingStatusCancelled

	status, err := svc.GetBookingFinancialStatus(context.Background(), booking)
	require.NoError(t, err)
	assert.Nil(t, status.CancellationDeadline)
	assert.Nil(t, status.CancellationHours)
}

func TestGetBookingFinancialStatus_NoDeadlineWhenPolicyIsZero(t *testing.T) {
	svc := &BookingService{
		productService: stubProductService{
			product: &catalogDomain.ProductRes{CancellationHours: 0},
		},
	}

	booking := newTestBooking()
	status, err := svc.GetBookingFinancialStatus(context.Background(), booking)
	require.NoError(t, err)
	assert.Nil(t, status.CancellationDeadline)
}

func TestGetBookingFinancialStatus_RefundedFullAmount(t *testing.T) {
	svc := &BookingService{
		productService: stubProductService{
			product: &catalogDomain.ProductRes{CancellationHours: 24},
		},
	}

	booking := newTestBooking()
	booking.Status = domain.BookingStatusCancelled
	booking.PaymentStatus = domain.PaymentStatusRefunded

	status, err := svc.GetBookingFinancialStatus(context.Background(), booking)
	require.NoError(t, err)

	require.NotNil(t, status.RefundAmountCents)
	assert.Equal(t, booking.TotalPriceCents, *status.RefundAmountCents)
	require.NotNil(t, status.RefundedAt)
	assert.Equal(t, booking.UpdatedAt, *status.RefundedAt)
	// Cancelled bookings are no longer cancellable, so no deadline is surfaced.
	assert.Nil(t, status.CancellationDeadline)
}

func TestGetBookingFinancialStatus_ProductLookupNotFoundFailsOpen(t *testing.T) {
	svc := &BookingService{
		productService: stubProductService{
			err: errs.ErrRepositoryNotFound,
		},
	}

	booking := newTestBooking()
	status, err := svc.GetBookingFinancialStatus(context.Background(), booking)
	require.NoError(t, err)
	assert.Nil(t, status.CancellationDeadline)
}
