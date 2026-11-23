package ports

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"

	"github.com/google/uuid"
)

// PushNotifier sends a practitioner notification to every device of a user
// (ADR-0020). It is what other domains call ("notify the practitioner"),
// without knowing about Web Push. With Web Push off (no VAPID keys) it sends
// nothing and returns nil. Callers run it in the background: a push failure
// never fails the operation that triggered it.
type PushNotifier interface {
	NotifyUser(ctx context.Context, userID uuid.UUID, msg domain.PushMessage) error
}

// PushStatus says whether Web Push is configured, and the VAPID public key
// the browser subscribes with (the application server key).
type PushStatus struct {
	Enabled   bool
	PublicKey string
}

// PushSubscriptionInput is a device's subscription as the browser's
// PushManager gives it.
type PushSubscriptionInput struct {
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent string
}

// PushSubscriptionService manages the current user's devices (the Cabinet
// switch).
type PushSubscriptionService interface {
	Status() PushStatus
	// Subscribe stores the device for the user, or moves it to them when it
	// was another user's (one row per device). domain.ErrPushUnavailable
	// while Web Push is off.
	Subscribe(ctx context.Context, userID uuid.UUID, input PushSubscriptionInput) error
	// Unsubscribe deletes the user's device with that endpoint; nothing
	// happens when the user has no such device.
	Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) error
}

// PushSender delivers one encrypted payload to one device through its push
// service (Web Push with VAPID). It returns domain.ErrPushSubscriptionGone
// when the push service reports the subscription gone (404/410). Its errors
// never contain the endpoint.
type PushSender interface {
	Send(ctx context.Context, sub domain.PushSubscription, payload []byte) error
}

// PushSubscriptionRepository stores push subscriptions encrypted (encx).
type PushSubscriptionRepository interface {
	// Upsert inserts the subscription, or replaces the row with the same
	// endpoint hash (same device): its user, keys and user agent.
	Upsert(ctx context.Context, sub *domain.PushSubscriptionEncx) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.PushSubscriptionEncx, error)
	// DeleteByEndpointHash deletes the user's subscription with that
	// endpoint hash; no error when there is none.
	DeleteByEndpointHash(ctx context.Context, userID uuid.UUID, endpointHash string) error
	// Delete deletes a subscription by ID; no error when it is already gone.
	Delete(ctx context.Context, id uuid.UUID) error
}
