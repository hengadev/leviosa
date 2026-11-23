package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
)

const (
	maxPushEndpointLength  = 2048
	maxPushKeyLength       = 256
	maxPushUserAgentLength = 512
)

// PushService stores the practitioner's devices and sends them practitioner
// notifications through Web Push (ADR-0020). Without a sender (no VAPID
// keys) the feature is off: Status says so, Subscribe refuses, NotifyUser
// sends nothing and returns nil.
type PushService struct {
	repo      ports.PushSubscriptionRepository
	sender    ports.PushSender
	crypto    encx.CryptoService
	publicKey string
	now       func() time.Time
}

var (
	_ ports.PushNotifier            = (*PushService)(nil)
	_ ports.PushSubscriptionService = (*PushService)(nil)
)

// NewPushService builds the push service. sender is nil, or publicKey
// empty, when Web Push is not configured.
func NewPushService(repo ports.PushSubscriptionRepository, sender ports.PushSender, crypto encx.CryptoService, publicKey string) *PushService {
	return &PushService{
		repo:      repo,
		sender:    sender,
		crypto:    crypto,
		publicKey: publicKey,
		now:       time.Now,
	}
}

func (s *PushService) enabled() bool {
	return s != nil && s.sender != nil && s.publicKey != ""
}

// Status says whether Web Push is on, with the VAPID public key when it is.
func (s *PushService) Status() ports.PushStatus {
	if !s.enabled() {
		return ports.PushStatus{}
	}
	return ports.PushStatus{Enabled: true, PublicKey: s.publicKey}
}

// Subscribe stores the user's device, encrypted.
func (s *PushService) Subscribe(ctx context.Context, userID uuid.UUID, input ports.PushSubscriptionInput) error {
	if !s.enabled() {
		return domain.ErrPushUnavailable
	}
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	input.P256dh = strings.TrimSpace(input.P256dh)
	input.Auth = strings.TrimSpace(input.Auth)
	if err := validatePushSubscription(input); err != nil {
		return err
	}
	userAgent := strings.TrimSpace(input.UserAgent)
	if len(userAgent) > maxPushUserAgentLength {
		userAgent = userAgent[:maxPushUserAgentLength]
	}

	sub := &domain.PushSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		Endpoint:  input.Endpoint,
		P256dh:    input.P256dh,
		Auth:      input.Auth,
		UserAgent: userAgent,
		CreatedAt: s.now().UTC(),
	}
	subEncx, err := domain.ProcessPushSubscriptionEncx(ctx, s.crypto, sub)
	if err != nil {
		return fmt.Errorf("encrypt push subscription: %w", err)
	}
	if err := s.repo.Upsert(ctx, subEncx); err != nil {
		return fmt.Errorf("store push subscription: %w", err)
	}
	return nil
}

// Unsubscribe deletes the user's device with that endpoint. It works with
// Web Push off too, so a device can always be switched off.
func (s *PushService) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return errs.NewInvalidValueErr("endpoint is required")
	}
	hash, err := s.endpointHash(ctx, endpoint)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteByEndpointHash(ctx, userID, hash); err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}
	return nil
}

// NotifyUser sends msg to every device of the user. A device the push
// service reports gone is deleted; the other failures are returned joined,
// after every device was tried. Nothing is logged that identifies a device
// beyond its subscription ID.
func (s *PushService) NotifyUser(ctx context.Context, userID uuid.UUID, msg domain.PushMessage) error {
	if !s.enabled() {
		return nil
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encode push payload: %w", err)
	}
	subs, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("list push subscriptions: %w", err)
	}

	var failures []error
	for _, subEncx := range subs {
		sub, err := domain.DecryptPushSubscriptionEncx(ctx, s.crypto, subEncx)
		if err != nil {
			failures = append(failures, fmt.Errorf("decrypt push subscription %s: %w", subEncx.ID, err))
			continue
		}
		err = s.sender.Send(ctx, *sub, payload)
		switch {
		case err == nil:
		case errors.Is(err, domain.ErrPushSubscriptionGone):
			if delErr := s.repo.Delete(ctx, sub.ID); delErr != nil {
				failures = append(failures, fmt.Errorf("delete gone push subscription %s: %w", sub.ID, delErr))
				continue
			}
			slog.InfoContext(ctx, "push subscription gone, deleted",
				"subscription_id", sub.ID, "user_id", userID)
		default:
			failures = append(failures, fmt.Errorf("push to subscription %s: %w", sub.ID, err))
		}
	}
	return errors.Join(failures...)
}

func (s *PushService) endpointHash(ctx context.Context, endpoint string) (string, error) {
	bytes, err := encx.SerializeValue(endpoint)
	if err != nil {
		return "", errs.NewInvalidValueErr("invalid endpoint")
	}
	return s.crypto.HashBasic(ctx, bytes), nil
}

// validatePushSubscription accepts what a browser's PushManager gives: an
// https endpoint and its two base64url keys.
func validatePushSubscription(input ports.PushSubscriptionInput) error {
	if input.Endpoint == "" || len(input.Endpoint) > maxPushEndpointLength {
		return errs.NewInvalidValueErr("endpoint is required")
	}
	u, err := url.Parse(input.Endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return errs.NewInvalidValueErr("endpoint must be an https URL")
	}
	if input.P256dh == "" || len(input.P256dh) > maxPushKeyLength {
		return errs.NewInvalidValueErr("keys.p256dh is required")
	}
	if input.Auth == "" || len(input.Auth) > maxPushKeyLength {
		return errs.NewInvalidValueErr("keys.auth is required")
	}
	return nil
}
