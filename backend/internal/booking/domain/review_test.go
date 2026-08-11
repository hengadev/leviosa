package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReview(t *testing.T) {
	bookingID := uuid.New()
	clientID := uuid.New()
	partnerID := uuid.New()
	productID := uuid.New()

	t.Run("valid review is created", func(t *testing.T) {
		review, err := NewReview(bookingID, clientID, partnerID, productID, 5, "Great session")
		require.NoError(t, err)
		require.NotNil(t, review)

		assert.NotEqual(t, uuid.Nil, review.ID)
		assert.Equal(t, bookingID, review.BookingID)
		assert.Equal(t, clientID, review.ClientID)
		assert.Equal(t, partnerID, review.PartnerID)
		assert.Equal(t, productID, review.ProductID)
		assert.Equal(t, 5, review.Rating)
		assert.Equal(t, "Great session", review.Comment)
		assert.False(t, review.Hidden)
		assert.False(t, review.CreatedAt.IsZero())
		assert.False(t, review.UpdatedAt.IsZero())
	})

	t.Run("comment is optional", func(t *testing.T) {
		review, err := NewReview(bookingID, clientID, partnerID, productID, 3, "")
		require.NoError(t, err)
		assert.Empty(t, review.Comment)
	})

	t.Run("rejects nil booking ID", func(t *testing.T) {
		_, err := NewReview(uuid.Nil, clientID, partnerID, productID, 5, "")
		assert.ErrorIs(t, err, ErrInvalidReviewBookingID)
	})

	t.Run("rejects nil client ID", func(t *testing.T) {
		_, err := NewReview(bookingID, uuid.Nil, partnerID, productID, 5, "")
		assert.ErrorIs(t, err, ErrInvalidClientID)
	})

	t.Run("rejects nil partner ID", func(t *testing.T) {
		_, err := NewReview(bookingID, clientID, uuid.Nil, productID, 5, "")
		assert.ErrorIs(t, err, ErrInvalidPartnerID)
	})

	t.Run("rejects rating bounds", func(t *testing.T) {
		tests := []struct {
			name   string
			rating int
		}{
			{"zero", 0},
			{"negative", -1},
			{"above max", 6},
			{"far above max", 100},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := NewReview(bookingID, clientID, partnerID, productID, tt.rating, "")
				assert.ErrorIs(t, err, ErrInvalidReviewRating)
			})
		}
	})

	t.Run("accepts rating bounds inclusive", func(t *testing.T) {
		for rating := MinReviewRating; rating <= MaxReviewRating; rating++ {
			_, err := NewReview(bookingID, clientID, partnerID, productID, rating, "")
			assert.NoError(t, err)
		}
	})

	t.Run("rejects comment exceeding max length", func(t *testing.T) {
		tooLong := strings.Repeat("a", MaxReviewCommentLength+1)
		_, err := NewReview(bookingID, clientID, partnerID, productID, 4, tooLong)
		assert.ErrorIs(t, err, ErrReviewCommentTooLong)
	})

	t.Run("accepts comment at exactly max length", func(t *testing.T) {
		exact := strings.Repeat("a", MaxReviewCommentLength)
		_, err := NewReview(bookingID, clientID, partnerID, productID, 4, exact)
		assert.NoError(t, err)
	})
}

func TestReviewHide(t *testing.T) {
	newVisibleReview := func(t *testing.T) *Review {
		t.Helper()
		review, err := NewReview(uuid.New(), uuid.New(), uuid.New(), uuid.New(), 4, "")
		require.NoError(t, err)
		return review
	}

	t.Run("hides a visible review", func(t *testing.T) {
		review := newVisibleReview(t)
		originalUpdatedAt := review.UpdatedAt

		err := review.Hide()

		require.NoError(t, err)
		assert.True(t, review.Hidden)
		assert.False(t, review.IsVisible())
		assert.True(t, review.UpdatedAt.Equal(originalUpdatedAt) || review.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("rejects hiding an already hidden review", func(t *testing.T) {
		review := newVisibleReview(t)
		require.NoError(t, review.Hide())

		err := review.Hide()

		assert.True(t, errors.Is(err, ErrReviewAlreadyHidden))
		assert.True(t, review.Hidden) // state unchanged
	})
}

func TestReviewUnhide(t *testing.T) {
	newHiddenReview := func(t *testing.T) *Review {
		t.Helper()
		review, err := NewReview(uuid.New(), uuid.New(), uuid.New(), uuid.New(), 4, "")
		require.NoError(t, err)
		require.NoError(t, review.Hide())
		return review
	}

	t.Run("unhides a hidden review", func(t *testing.T) {
		review := newHiddenReview(t)

		err := review.Unhide()

		require.NoError(t, err)
		assert.False(t, review.Hidden)
		assert.True(t, review.IsVisible())
	})

	t.Run("rejects unhiding a review that is not hidden", func(t *testing.T) {
		review, err := NewReview(uuid.New(), uuid.New(), uuid.New(), uuid.New(), 4, "")
		require.NoError(t, err)

		err = review.Unhide()

		assert.True(t, errors.Is(err, ErrReviewNotHidden))
		assert.False(t, review.Hidden) // state unchanged
	})
}

func TestReviewIsVisible(t *testing.T) {
	review, err := NewReview(uuid.New(), uuid.New(), uuid.New(), uuid.New(), 4, "")
	require.NoError(t, err)

	assert.True(t, review.IsVisible())

	require.NoError(t, review.Hide())
	assert.False(t, review.IsVisible())

	require.NoError(t, review.Unhide())
	assert.True(t, review.IsVisible())
}
