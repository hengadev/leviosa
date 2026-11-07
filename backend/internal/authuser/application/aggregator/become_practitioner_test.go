package aggregator

import (
	"context"
	"fmt"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeOwnProfile is the admin's practitioner profile, or none when nil.
type fakeOwnProfile struct {
	fakePractitionerPartners
	existing *domain.PartnerResponse
	created  bool
}

func (f *fakeOwnProfile) GetPartnerByUserID(_ context.Context, _ uuid.UUID) (*domain.PartnerResponse, error) {
	if f.existing == nil {
		return nil, fmt.Errorf("get partner by user ID: %w", errs.ErrRepositoryNotFound)
	}
	return f.existing, nil
}

func (f *fakeOwnProfile) CreatePartner(ctx context.Context, userID uuid.UUID, bio, experience string, categoryIDs, productIDs []uuid.UUID) (*domain.Partner, error) {
	f.created = true
	return f.fakePractitionerPartners.CreatePartner(ctx, userID, bio, experience, categoryIDs, productIDs)
}

func TestBecomePractitioner_CreatesTheProfileAndTheRoom(t *testing.T) {
	partners := &fakeOwnProfile{fakePractitionerPartners: fakePractitionerPartners{partnerID: uuid.New()}}
	room := &fakePracticeRoom{}
	svc := New(nil, nil, nil, partners, room)
	adminID := uuid.New()

	partnerID, err := svc.BecomePractitioner(context.Background(), adminID, &domain.BecomePractitionerRequest{Occupation: "Ostéopathe", Bio: "Bio"})

	require.NoError(t, err)
	assert.Equal(t, partners.partnerID, partnerID)
	assert.True(t, partners.created)
	assert.Equal(t, "Bio", partners.bio)
	assert.Equal(t, "Ostéopathe", partners.occupation)
	assert.Equal(t, adminID, partners.verifiedBy)
	assert.Equal(t, []uuid.UUID{adminID}, room.provided)
}

func TestBecomePractitioner_FinishesAnExistingProfile(t *testing.T) {
	existing := &domain.PartnerResponse{ID: uuid.New()}
	partners := &fakeOwnProfile{existing: existing}
	room := &fakePracticeRoom{}
	svc := New(nil, nil, nil, partners, room)
	adminID := uuid.New()

	partnerID, err := svc.BecomePractitioner(context.Background(), adminID, &domain.BecomePractitionerRequest{})

	require.NoError(t, err)
	assert.Equal(t, existing.ID, partnerID)
	assert.False(t, partners.created, "the existing profile is kept")
	assert.Equal(t, adminID, partners.verifiedBy, "the unverified profile is verified")
	assert.Equal(t, []uuid.UUID{adminID}, room.provided)
}

func TestBecomePractitioner_DoesNotVerifyTwice(t *testing.T) {
	existing := &domain.PartnerResponse{
		ID:                       uuid.New(),
		StripeAccountStatus:      domain.StripeAccountStatusActive,
		StripeOnboardingComplete: true,
	}
	partners := &fakeOwnProfile{existing: existing}
	room := &fakePracticeRoom{}
	svc := New(nil, nil, nil, partners, room)

	_, err := svc.BecomePractitioner(context.Background(), uuid.New(), &domain.BecomePractitionerRequest{})

	require.NoError(t, err)
	assert.Equal(t, uuid.Nil, partners.verifiedBy)
	assert.Len(t, room.provided, 1)
}
