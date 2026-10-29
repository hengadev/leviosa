package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDepositValidate(t *testing.T) {
	tests := []struct {
		name    string
		deposit *Deposit
		wantErr bool
	}{
		{name: "nil deposit is valid (no deposit)", deposit: nil, wantErr: false},
		{name: "fixed deposit of zero cents is valid", deposit: &Deposit{Type: DepositFixed, Value: 0}, wantErr: false},
		{name: "fixed deposit is valid", deposit: &Deposit{Type: DepositFixed, Value: 2000}, wantErr: false},
		{name: "fixed deposit cannot be negative", deposit: &Deposit{Type: DepositFixed, Value: -1}, wantErr: true},
		{name: "percentage deposit is valid", deposit: &Deposit{Type: DepositPercent, Value: 30}, wantErr: false},
		{name: "percentage of 100 is valid", deposit: &Deposit{Type: DepositPercent, Value: 100}, wantErr: false},
		{name: "percentage of 0 is invalid", deposit: &Deposit{Type: DepositPercent, Value: 0}, wantErr: true},
		{name: "percentage below 0 is invalid", deposit: &Deposit{Type: DepositPercent, Value: -5}, wantErr: true},
		{name: "percentage above 100 is invalid", deposit: &Deposit{Type: DepositPercent, Value: 101}, wantErr: true},
		{name: "unknown type is invalid", deposit: &Deposit{Type: DepositType("arrhes"), Value: 100}, wantErr: true},
		{name: "empty type is invalid", deposit: &Deposit{Type: "", Value: 100}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.deposit.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAmountDueOnline(t *testing.T) {
	tests := []struct {
		name        string
		deposit     *Deposit
		priceCents  int
		want        int
		wantErr     bool
		description string
	}{
		{
			name:        "no deposit: the full price is due online",
			deposit:     nil,
			priceCents:  7000,
			want:        7000,
			description: "70 € session with no deposit configured",
		},
		{
			name:        "fixed deposit",
			deposit:     &Deposit{Type: DepositFixed, Value: 2000},
			priceCents:  7000,
			want:        2000,
			description: "20 € deposit on a 70 € session",
		},
		{
			name:        "fixed deposit equal to the price",
			deposit:     &Deposit{Type: DepositFixed, Value: 7000},
			priceCents:  7000,
			want:        7000,
			description: "the deposit may cover the whole price",
		},
		{
			name:        "fixed deposit of zero cents",
			deposit:     &Deposit{Type: DepositFixed, Value: 0},
			priceCents:  7000,
			want:        0,
			description: "nothing collected online",
		},
		{
			name:        "fixed deposit above the price is rejected",
			deposit:     &Deposit{Type: DepositFixed, Value: 7001},
			priceCents:  7000,
			wantErr:     true,
			description: "the online amount can never exceed the price",
		},
		{
			name:        "percentage deposit, exact division",
			deposit:     &Deposit{Type: DepositPercent, Value: 30},
			priceCents:  7000,
			want:        2100,
			description: "30% of 70 €",
		},
		{
			name:        "percentage deposit of 100 is the full price",
			deposit:     &Deposit{Type: DepositPercent, Value: 100},
			priceCents:  5500,
			want:        5500,
			description: "100% deposit",
		},
		{
			name:        "percentage deposit rounds half up to the cent",
			deposit:     &Deposit{Type: DepositPercent, Value: 33},
			priceCents:  999,
			want:        330,
			description: "33% of 9.99 € = 3.2967 € → 3.30 €",
		},
		{
			name:        "percentage deposit rounds half down to the cent",
			deposit:     &Deposit{Type: DepositPercent, Value: 15},
			priceCents:  999,
			want:        150,
			description: "15% of 9.99 € = 1.4985 € → 1.50 € (round half away from zero)",
		},
		{
			name:        "percentage deposit rounds a sub-cent result up",
			deposit:     &Deposit{Type: DepositPercent, Value: 10},
			priceCents:  105,
			want:        11,
			description: "10% of 1.05 € = 0.105 € → 0.11 €",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AmountDueOnline(tc.deposit, tc.priceCents)
			if tc.wantErr {
				require.Error(t, err, tc.description)
				return
			}
			require.NoError(t, err, tc.description)
			assert.Equal(t, tc.want, got, tc.description)
		})
	}
}

func TestDepositPatchUnmarshalJSON(t *testing.T) {
	t.Run("absent field leaves the patch zero-valued", func(t *testing.T) {
		var req struct {
			Deposit DepositPatch `json:"deposit,omitempty"`
		}
		require.NoError(t, json.Unmarshal([]byte(`{"name":"x"}`), &req))
		assert.False(t, req.Deposit.Present)
		assert.Nil(t, req.Deposit.Value)
	})

	t.Run("JSON null clears the deposit", func(t *testing.T) {
		var req struct {
			Deposit DepositPatch `json:"deposit,omitempty"`
		}
		require.NoError(t, json.Unmarshal([]byte(`{"deposit":null}`), &req))
		assert.True(t, req.Deposit.Present)
		assert.Nil(t, req.Deposit.Value)
	})

	t.Run("JSON object sets the deposit", func(t *testing.T) {
		var req struct {
			Deposit DepositPatch `json:"deposit,omitempty"`
		}
		require.NoError(t, json.Unmarshal([]byte(`{"deposit":{"type":"percent","value":30}}`), &req))
		assert.True(t, req.Deposit.Present)
		require.NotNil(t, req.Deposit.Value)
		assert.Equal(t, Deposit{Type: DepositPercent, Value: 30}, *req.Deposit.Value)
	})

	t.Run("a non-object, non-null value is rejected", func(t *testing.T) {
		var req struct {
			Deposit DepositPatch `json:"deposit,omitempty"`
		}
		assert.Error(t, json.Unmarshal([]byte(`{"deposit":"none"}`), &req))
	})
}

func TestDepositPatchRoundTrip(t *testing.T) {
	type req struct {
		Name    string       `json:"name"`
		Deposit DepositPatch `json:"deposit,omitzero"`
	}

	t.Run("an untouched patch is omitted", func(t *testing.T) {
		data, err := json.Marshal(req{Name: "x"})
		require.NoError(t, err)
		assert.JSONEq(t, `{"name":"x"}`, string(data))
	})

	t.Run("a removal marshals to null and back", func(t *testing.T) {
		data, err := json.Marshal(req{Deposit: DepositPatch{Present: true}})
		require.NoError(t, err)
		var got req
		require.NoError(t, json.Unmarshal(data, &got))
		assert.True(t, got.Deposit.Present)
		assert.Nil(t, got.Deposit.Value)
	})

	t.Run("a deposit marshals and unmarshals unchanged", func(t *testing.T) {
		in := req{Deposit: DepositPatch{Present: true, Value: &Deposit{Type: DepositFixed, Value: 2000}}}
		data, err := json.Marshal(in)
		require.NoError(t, err)
		var got req
		require.NoError(t, json.Unmarshal(data, &got))
		assert.Equal(t, in.Deposit, got.Deposit)
	})
}

func TestDepositValidateAgainstPrice(t *testing.T) {
	var none *Deposit
	assert.NoError(t, none.ValidateAgainstPrice(7000), "no deposit always fits")
	assert.NoError(t, (&Deposit{Type: DepositFixed, Value: 7000}).ValidateAgainstPrice(7000))
	assert.Error(t, (&Deposit{Type: DepositFixed, Value: 7001}).ValidateAgainstPrice(7000))
	assert.NoError(t, (&Deposit{Type: DepositPercent, Value: 100}).ValidateAgainstPrice(7000))
	assert.Error(t, (&Deposit{Type: DepositPercent, Value: 0}).ValidateAgainstPrice(7000))
}
