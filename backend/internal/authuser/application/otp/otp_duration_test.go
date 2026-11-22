package otp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeOTPSettings answers a fixed duration or error.
type fakeOTPSettings struct {
	duration time.Duration
	err      error
}

func (f fakeOTPSettings) GetOTPDuration(context.Context) (time.Duration, error) {
	return f.duration, f.err
}

func TestOTPDuration(t *testing.T) {
	ctx := context.Background()
	def := time.Duration(defaultOTPDuration) * time.Minute

	tests := []struct {
		name     string
		settings *fakeOTPSettings
		want     time.Duration
	}{
		{"no settings provider: the default", nil, def},
		{"the setting", &fakeOTPSettings{duration: 5 * time.Minute}, 5 * time.Minute},
		{"lowest accepted value", &fakeOTPSettings{duration: time.Minute}, time.Minute},
		{"highest accepted value", &fakeOTPSettings{duration: time.Hour}, time.Hour},
		{"unreadable setting: the default", &fakeOTPSettings{err: errors.New("db down")}, def},
		{"too short (an old value in minutes): the default", &fakeOTPSettings{duration: 15 * time.Second}, def},
		{"too long: the default", &fakeOTPSettings{duration: 2 * time.Hour}, def},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &OTPService{}
			if tt.settings != nil {
				s.settings = *tt.settings
			}
			assert.Equal(t, tt.want, s.otpDuration(ctx))
		})
	}
}

// A new code lasts the duration set in the admin settings: in the store and
// in the email.
func TestOTPUsesTheDurationSetting(t *testing.T) {
	ctx := context.Background()
	newService := func(t *testing.T, repo *inMemRepo, notif *mockNotifier) *OTPService {
		return &OTPService{
			repo:            repo,
			crypto:          newTestCrypto(t),
			cache:           NewMockOTPCache(defaultOTPLength, defaultOTPDuration, defaultOTPMaxAttempts),
			notificationSvc: notif,
			settings:        fakeOTPSettings{duration: 5 * time.Minute},
		}
	}

	t.Run("RequestOTP", func(t *testing.T) {
		repo, notif := newInMemRepo(), &mockNotifier{}
		require.NoError(t, newService(t, repo, notif).RequestOTP(ctx, "user@example.com"))
		assert.Equal(t, 5*time.Minute, repo.lastTTL)
		require.Len(t, notif.calls, 1)
		assert.Equal(t, 5*time.Minute, notif.calls[0].ValidFor)
	})

	t.Run("CreateOTP", func(t *testing.T) {
		repo, notif := newInMemRepo(), &mockNotifier{}
		require.NoError(t, newService(t, repo, notif).CreateOTP(ctx, "user@example.com"))
		assert.Equal(t, 5*time.Minute, repo.lastTTL)
		require.Len(t, notif.calls, 1)
		assert.Equal(t, 5*time.Minute, notif.calls[0].ValidFor)
	})

	t.Run("ResendOTP states the time left", func(t *testing.T) {
		repo, notif := newInMemRepo(), &mockNotifier{}
		svc := newService(t, repo, notif)
		require.NoError(t, svc.RequestOTP(ctx, "user@example.com"))
		notif.calls = nil

		require.NoError(t, svc.ResendOTP(ctx, "user@example.com"))
		require.Len(t, notif.calls, 1)
		left := notif.calls[0].ValidFor
		assert.True(t, left > 4*time.Minute && left <= 5*time.Minute, "time left = %s", left)
	})
}
