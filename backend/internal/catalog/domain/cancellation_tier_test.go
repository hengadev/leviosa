package domain_test

import (
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"

	"github.com/stretchr/testify/assert"
)

func TestValidateCancellationTiers(t *testing.T) {
	tests := []struct {
		name    string
		tiers   []domain.CancellationTier
		wantErr bool
	}{
		{
			name:    "nil tiers is valid (feature not opted into)",
			tiers:   nil,
			wantErr: false,
		},
		{
			name:    "empty tiers is valid (feature not opted into)",
			tiers:   []domain.CancellationTier{},
			wantErr: false,
		},
		{
			name: "single tier covering from 0h is valid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: false,
		},
		{
			name: "standard three-tier schedule is valid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 48, RefundPercentage: 100},
				{MinHoursBefore: 24, RefundPercentage: 50},
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: false,
		},
		{
			name: "refund percentage above 100 is invalid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 48, RefundPercentage: 150},
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: true,
		},
		{
			name: "negative refund percentage is invalid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 48, RefundPercentage: -10},
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: true,
		},
		{
			name: "negative minHoursBefore is invalid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: -1, RefundPercentage: 100},
			},
			wantErr: true,
		},
		{
			name: "tiers not sorted descending is invalid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 24, RefundPercentage: 50},
				{MinHoursBefore: 48, RefundPercentage: 100},
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: true,
		},
		{
			name: "duplicate minHoursBefore thresholds is invalid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 48, RefundPercentage: 100},
				{MinHoursBefore: 48, RefundPercentage: 50},
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: true,
		},
		{
			name: "missing 0h tier does not fully cover the range",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 48, RefundPercentage: 100},
				{MinHoursBefore: 24, RefundPercentage: 50},
			},
			wantErr: true,
		},
		{
			name: "single tier not anchored at 0h is invalid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 12, RefundPercentage: 100},
			},
			wantErr: true,
		},
		{
			name: "boundary refund percentages 0 and 100 are valid",
			tiers: []domain.CancellationTier{
				{MinHoursBefore: 1, RefundPercentage: 100},
				{MinHoursBefore: 0, RefundPercentage: 0},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateCancellationTiers(tc.tiers)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductValid_CancellationTiers(t *testing.T) {
	base := func() domain.Product {
		return domain.Product{
			Name:         "Swedish Massage",
			Duration:     60,
			Status:       domain.Draft,
			Availability: domain.Online,
			BufferTime:   15,
		}
	}

	t.Run("product with valid tiers passes validation w.r.t. tiers", func(t *testing.T) {
		p := base()
		p.CancellationTiers = []domain.CancellationTier{
			{MinHoursBefore: 48, RefundPercentage: 100},
			{MinHoursBefore: 24, RefundPercentage: 50},
			{MinHoursBefore: 0, RefundPercentage: 0},
		}
		err := p.Valid(t.Context())
		assert.NoError(t, err)
	})

	t.Run("product with invalid tiers fails validation", func(t *testing.T) {
		p := base()
		p.CancellationTiers = []domain.CancellationTier{
			{MinHoursBefore: 24, RefundPercentage: 150},
		}
		err := p.Valid(t.Context())
		assert.Error(t, err)
	})
}
