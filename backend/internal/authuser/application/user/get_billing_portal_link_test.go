package user_test

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/application/user"
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
)

// This file uses its own dedicated mock implementations (billingPortal-prefixed) rather than
// the package's shared mockUserRepository/mockCryptoService/mockStripeService, whose
// DecryptData/DecryptDEKWithVersion stubs are hardcoded no-ops for other tests' purposes and
// aren't wired to per-test expectations.

type billingPortalMockUserRepository struct {
	mock.Mock
}

func (m *billingPortalMockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.UserEncx, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserEncx), args.Error(1)
}
func (m *billingPortalMockUserRepository) ExistsByEmailHash(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *billingPortalMockUserRepository) GetUserByEmailHash(_ context.Context, _ string) (*domain.UserEncx, error) {
	return nil, nil
}
func (m *billingPortalMockUserRepository) GetPendingUsers(_ context.Context) ([]*domain.UserEncx, error) {
	return nil, nil
}
func (m *billingPortalMockUserRepository) GetAllUsers(_ context.Context) ([]*domain.UserEncx, error) {
	return nil, nil
}
func (m *billingPortalMockUserRepository) CreateUser(_ context.Context, _ *domain.UserEncx) error {
	return nil
}
func (m *billingPortalMockUserRepository) UpdateUser(_ context.Context, _ *domain.UserEncx) error {
	return nil
}
func (m *billingPortalMockUserRepository) DeleteUser(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *billingPortalMockUserRepository) GetUserByGoogleID(_ context.Context, _ string) (*domain.UserEncx, error) {
	return nil, nil
}
func (m *billingPortalMockUserRepository) GetUserByAppleID(_ context.Context, _ string) (*domain.UserEncx, error) {
	return nil, nil
}
func (m *billingPortalMockUserRepository) ExistsByGoogleID(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *billingPortalMockUserRepository) ExistsByAppleID(_ context.Context, _ string) (bool, error) {
	return false, nil
}

type billingPortalMockCryptoService struct {
	mock.Mock
}

func (m *billingPortalMockCryptoService) GetPepper() []byte                  { return nil }
func (m *billingPortalMockCryptoService) GetArgon2Params() *encx.Argon2Params { return nil }
func (m *billingPortalMockCryptoService) GetAlias() string                   { return "" }
func (m *billingPortalMockCryptoService) GenerateDEK() ([]byte, error)       { return nil, nil }
func (m *billingPortalMockCryptoService) EncryptData(_ context.Context, _ []byte, _ []byte) ([]byte, error) {
	return nil, nil
}
func (m *billingPortalMockCryptoService) DecryptData(ctx context.Context, ciphertext []byte, dek []byte) ([]byte, error) {
	args := m.Called(ctx, ciphertext, dek)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
func (m *billingPortalMockCryptoService) EncryptDEK(_ context.Context, _ []byte) ([]byte, error) {
	return nil, nil
}
func (m *billingPortalMockCryptoService) DecryptDEKWithVersion(ctx context.Context, ciphertextDEK []byte, kekVersion int) ([]byte, error) {
	args := m.Called(ctx, ciphertextDEK, kekVersion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
func (m *billingPortalMockCryptoService) RotateKEK(_ context.Context) error { return nil }
func (m *billingPortalMockCryptoService) HashBasic(_ context.Context, _ []byte) string {
	return ""
}
func (m *billingPortalMockCryptoService) HashSecure(_ context.Context, _ []byte) (string, error) {
	return "", nil
}
func (m *billingPortalMockCryptoService) CompareSecureHashAndValue(_ context.Context, _ any, _ string) (bool, error) {
	return false, nil
}
func (m *billingPortalMockCryptoService) CompareBasicHashAndValue(_ context.Context, _ any, _ string) (bool, error) {
	return false, nil
}
func (m *billingPortalMockCryptoService) EncryptStream(_ context.Context, _ io.Reader, _ io.Writer, _ []byte) error {
	return nil
}
func (m *billingPortalMockCryptoService) DecryptStream(_ context.Context, _ io.Reader, _ io.Writer, _ []byte) error {
	return nil
}
func (m *billingPortalMockCryptoService) GetCurrentKEKVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *billingPortalMockCryptoService) GetKMSKeyIDForVersion(_ context.Context, _ string, _ int) (string, error) {
	return "", nil
}

type billingPortalMockStripeService struct {
	mock.Mock
}

func (m *billingPortalMockStripeService) CreateCustomer(_ context.Context, _ uuid.UUID, _, _, _ string) (*stripe.Customer, error) {
	return nil, nil
}
func (m *billingPortalMockStripeService) RetrieveCustomer(_ context.Context, _ string) (*stripe.Customer, error) {
	return nil, nil
}
func (m *billingPortalMockStripeService) UpdateCustomer(_ context.Context, _ string, _ *stripe.CustomerUpdateParams) (*stripe.Customer, error) {
	return nil, nil
}
func (m *billingPortalMockStripeService) DeleteCustomer(_ context.Context, _ string) (*stripe.Customer, error) {
	return nil, nil
}
func (m *billingPortalMockStripeService) FindCustomerByUserID(_ context.Context, _ uuid.UUID) (*stripe.Customer, error) {
	return nil, nil
}
func (m *billingPortalMockStripeService) CreateConnectedAccount(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}
func (m *billingPortalMockStripeService) CreateAccountLink(_ context.Context, _, _, _, _ string) (string, error) {
	return "", nil
}
func (m *billingPortalMockStripeService) VerifyConnectWebhookSignature(_ []byte, _ string) (string, bool, bool, error) {
	return "", false, false, nil
}
func (m *billingPortalMockStripeService) CreateBillingPortalSession(ctx context.Context, customerID, returnURL string) (string, error) {
	args := m.Called(ctx, customerID, returnURL)
	return args.String(0), args.Error(1)
}

// serializeString encodes a string in encx compact binary format: [4-byte LE length][UTF-8 data]
func serializeString(s string) []byte {
	data := []byte(s)
	result := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(result[:4], uint32(len(data)))
	copy(result[4:], data)
	return result
}

// buildBillingPortalUserEncx creates a UserEncx with only the fields needed by
// GetBillingPortalLink: the DEK (required to decrypt anything) and, optionally, an encrypted
// Stripe customer ID. All other encrypted fields are left empty so DecryptUserEncx skips them.
func buildBillingPortalUserEncx(userID uuid.UUID, stripeCustomerID string) *domain.UserEncx {
	encx := &domain.UserEncx{
		ID:           userID,
		DEKEncrypted: []byte("encrypted-dek"),
		KeyVersion:   1,
	}

	if stripeCustomerID != "" {
		encx.StripeCustomerIDEncrypted = []byte("encrypted-stripe-customer-id")
	}

	return encx
}

func TestGetBillingPortalLink(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	returnURL := "https://example.com/client/profile?stripe=return"

	t.Run("should return billing portal link when user has a Stripe customer ID", func(t *testing.T) {
		mockRepo := new(billingPortalMockUserRepository)
		mockCrypto := new(billingPortalMockCryptoService)
		mockStripe := new(billingPortalMockStripeService)

		stripeCustomerID := "cus_existing_123"
		expectedURL := "https://billing.stripe.com/session/test_123"

		userEncx := buildBillingPortalUserEncx(userID, stripeCustomerID)

		mockRepo.On("GetUserByID", ctx, userID).Return(userEncx, nil)

		mockCrypto.On("DecryptDEKWithVersion", ctx, mock.Anything, mock.Anything).Return([]byte("test-dek-32-bytes-long-enough!!!"), nil)
		mockCrypto.On("DecryptData", ctx, mock.Anything, mock.Anything).Return(serializeString(stripeCustomerID), nil)

		mockStripe.On("CreateBillingPortalSession", ctx, stripeCustomerID, returnURL).Return(expectedURL, nil)

		svc := user.New(mockRepo, mockCrypto, mockStripe)

		url, err := svc.GetBillingPortalLink(ctx, userID, returnURL)

		require.NoError(t, err)
		assert.Equal(t, expectedURL, url)

		mockRepo.AssertExpectations(t)
		mockCrypto.AssertExpectations(t)
		mockStripe.AssertExpectations(t)
	})

	t.Run("should return error when user has no Stripe customer ID", func(t *testing.T) {
		mockRepo := new(billingPortalMockUserRepository)
		mockCrypto := new(billingPortalMockCryptoService)
		mockStripe := new(billingPortalMockStripeService)

		userEncx := buildBillingPortalUserEncx(userID, "")

		mockRepo.On("GetUserByID", ctx, userID).Return(userEncx, nil)
		mockCrypto.On("DecryptDEKWithVersion", ctx, mock.Anything, mock.Anything).Return([]byte("test-dek-32-bytes-long-enough!!!"), nil)

		svc := user.New(mockRepo, mockCrypto, mockStripe)

		url, err := svc.GetBillingPortalLink(ctx, userID, returnURL)

		require.Error(t, err)
		assert.Empty(t, url)
		assert.True(t, errors.Is(err, errs.ErrDomainNotFound))

		mockStripe.AssertNotCalled(t, "CreateBillingPortalSession", mock.Anything, mock.Anything, mock.Anything)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockRepo := new(billingPortalMockUserRepository)
		mockCrypto := new(billingPortalMockCryptoService)
		mockStripe := new(billingPortalMockStripeService)

		mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("not found"))

		svc := user.New(mockRepo, mockCrypto, mockStripe)

		url, err := svc.GetBillingPortalLink(ctx, userID, returnURL)

		require.Error(t, err)
		assert.Empty(t, url)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when CreateBillingPortalSession fails", func(t *testing.T) {
		mockRepo := new(billingPortalMockUserRepository)
		mockCrypto := new(billingPortalMockCryptoService)
		mockStripe := new(billingPortalMockStripeService)

		stripeCustomerID := "cus_existing_456"
		userEncx := buildBillingPortalUserEncx(userID, stripeCustomerID)

		mockRepo.On("GetUserByID", ctx, userID).Return(userEncx, nil)
		mockCrypto.On("DecryptDEKWithVersion", ctx, mock.Anything, mock.Anything).Return([]byte("test-dek-32-bytes-long-enough!!!"), nil)
		mockCrypto.On("DecryptData", ctx, mock.Anything, mock.Anything).Return(serializeString(stripeCustomerID), nil)

		mockStripe.On("CreateBillingPortalSession", ctx, stripeCustomerID, returnURL).Return("", errors.New("stripe billing portal error"))

		svc := user.New(mockRepo, mockCrypto, mockStripe)

		url, err := svc.GetBillingPortalLink(ctx, userID, returnURL)

		require.Error(t, err)
		assert.Empty(t, url)
		assert.Contains(t, err.Error(), "create billing portal session")

		mockStripe.AssertExpectations(t)
	})
}
