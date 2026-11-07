package aggregator

import (
	"context"
	"errors"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/ports"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePractitionerUsers struct {
	ports.UserService
	userID   uuid.UUID
	password string
	deleted  []uuid.UUID
}

func (f *fakePractitionerUsers) CreatePractitionerUser(_ context.Context, _ *domain.CreatePractitionerRequest, password string) (uuid.UUID, error) {
	f.password = password
	return f.userID, nil
}

func (f *fakePractitionerUsers) DeleteUser(_ context.Context, userID uuid.UUID) error {
	f.deleted = append(f.deleted, userID)
	return nil
}

type fakePractitionerPartners struct {
	ports.PartnerService
	partnerID  uuid.UUID
	bio        string
	occupation string
	verifiedBy uuid.UUID
}

func (f *fakePractitionerPartners) CreatePartner(_ context.Context, userID uuid.UUID, bio, _ string, _, _ []uuid.UUID) (*domain.Partner, error) {
	f.bio = bio
	return &domain.Partner{ID: f.partnerID, UserID: userID}, nil
}

func (f *fakePractitionerPartners) UpdatePartner(_ context.Context, _ uuid.UUID, req *domain.UpdatePartnerRequest) (*domain.PartnerResponse, error) {
	f.occupation = *req.Occupation
	return &domain.PartnerResponse{}, nil
}

func (f *fakePractitionerPartners) VerifyPartner(_ context.Context, _ uuid.UUID, verifiedBy uuid.UUID) (*domain.PartnerResponse, error) {
	f.verifiedBy = verifiedBy
	return &domain.PartnerResponse{}, nil
}

type fakePracticeRoom struct {
	ports.BookingClient
	err      error
	provided []uuid.UUID
}

func (f *fakePracticeRoom) EnsurePracticeRoom(_ context.Context, userID uuid.UUID) error {
	f.provided = append(f.provided, userID)
	return f.err
}

func practitionerRequest() *domain.CreatePractitionerRequest {
	return &domain.CreatePractitionerRequest{
		FirstName:  "Camille",
		LastName:   "Martin",
		Email:      "camille.martin@example.com",
		Telephone:  "0612345678",
		Occupation: "Ostéopathe",
		Bio:        "Bio",
	}
}

func TestCreatePractitioner_CreatesAccountProfileAndRoom(t *testing.T) {
	users := &fakePractitionerUsers{userID: uuid.New()}
	partners := &fakePractitionerPartners{partnerID: uuid.New()}
	room := &fakePracticeRoom{}
	svc := New(nil, users, nil, partners, room)
	adminID := uuid.New()

	res, err := svc.CreatePractitioner(context.Background(), adminID, practitionerRequest())

	require.NoError(t, err)
	assert.Equal(t, users.userID, res.UserID)
	assert.Equal(t, partners.partnerID, res.PartnerID)
	assert.Len(t, res.Password, domain.GeneratedPasswordLength)
	assert.Equal(t, users.password, res.Password, "the account gets the password returned")
	assert.Equal(t, "Bio", partners.bio)
	assert.Equal(t, "Ostéopathe", partners.occupation)
	assert.Equal(t, adminID, partners.verifiedBy)
	assert.Equal(t, []uuid.UUID{users.userID}, room.provided)
	assert.Empty(t, users.deleted)
}

func TestCreatePractitioner_DeletesTheAccountWhenTheRoomFails(t *testing.T) {
	users := &fakePractitionerUsers{userID: uuid.New()}
	partners := &fakePractitionerPartners{partnerID: uuid.New()}
	room := &fakePracticeRoom{err: errors.New("no room")}
	svc := New(nil, users, nil, partners, room)

	_, err := svc.CreatePractitioner(context.Background(), uuid.New(), practitionerRequest())

	require.Error(t, err)
	assert.Equal(t, []uuid.UUID{users.userID}, users.deleted)
}

func TestCreatePractitioner_RejectsAnInvalidRequest(t *testing.T) {
	users := &fakePractitionerUsers{userID: uuid.New()}
	svc := New(nil, users, nil, &fakePractitionerPartners{}, &fakePracticeRoom{})
	req := practitionerRequest()
	req.Email = "not an email"

	_, err := svc.CreatePractitioner(context.Background(), uuid.New(), req)

	require.Error(t, err)
	assert.Empty(t, users.password, "no account is created")
}
