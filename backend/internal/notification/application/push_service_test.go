package application

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
)

// memoryPushRepo is an in-memory PushSubscriptionRepository keyed by
// endpoint hash, like the table's unique constraint.
type memoryPushRepo struct {
	mu   sync.Mutex
	rows map[string]*domain.PushSubscriptionEncx
}

func newMemoryPushRepo() *memoryPushRepo {
	return &memoryPushRepo{rows: map[string]*domain.PushSubscriptionEncx{}}
}

func (r *memoryPushRepo) Upsert(_ context.Context, sub *domain.PushSubscriptionEncx) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.rows[sub.EndpointHash]; ok {
		sub.ID = existing.ID
	}
	r.rows[sub.EndpointHash] = sub
	return nil
}

func (r *memoryPushRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]*domain.PushSubscriptionEncx, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.PushSubscriptionEncx
	for _, row := range r.rows {
		if row.UserID == userID {
			out = append(out, row)
		}
	}
	return out, nil
}

func (r *memoryPushRepo) DeleteByEndpointHash(_ context.Context, userID uuid.UUID, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if row, ok := r.rows[hash]; ok && row.UserID == userID {
		delete(r.rows, hash)
	}
	return nil
}

func (r *memoryPushRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, row := range r.rows {
		if row.ID == id {
			delete(r.rows, hash)
		}
	}
	return nil
}

func (r *memoryPushRepo) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rows)
}

// fakePushSender records what it sends; endpoints in gone answer as a push
// service does for a removed subscription, those in failing fail.
type fakePushSender struct {
	mu      sync.Mutex
	sent    map[string][]byte
	gone    map[string]bool
	failing map[string]bool
}

func newFakePushSender() *fakePushSender {
	return &fakePushSender{sent: map[string][]byte{}, gone: map[string]bool{}, failing: map[string]bool{}}
}

func (s *fakePushSender) Send(_ context.Context, sub domain.PushSubscription, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gone[sub.Endpoint] {
		return domain.ErrPushSubscriptionGone
	}
	if s.failing[sub.Endpoint] {
		return errors.New("push service answered 500")
	}
	s.sent[sub.Endpoint] = payload
	return nil
}

func newTestPushService(t *testing.T, sender ports.PushSender, publicKey string) (*PushService, *memoryPushRepo) {
	t.Helper()
	crypto, err := encx.NewTestCrypto(t)
	if err != nil {
		t.Fatalf("test crypto: %v", err)
	}
	repo := newMemoryPushRepo()
	return NewPushService(repo, sender, crypto, publicKey), repo
}

func subscription(endpoint string) ports.PushSubscriptionInput {
	return ports.PushSubscriptionInput{Endpoint: endpoint, P256dh: "BPk3", Auth: "c2VjcmV0", UserAgent: "iPhone"}
}

const (
	iphoneEndpoint = "https://web.push.apple.com/QGuQyavXutnMH"
	laptopEndpoint = "https://fcm.googleapis.com/fcm/send/abc"
)

func TestPushService_NotifyUserSendsToEveryDeviceOfTheUser(t *testing.T) {
	ctx := context.Background()
	sender := newFakePushSender()
	svc, _ := newTestPushService(t, sender, "public-key")
	practitioner, other := uuid.New(), uuid.New()

	for _, endpoint := range []string{iphoneEndpoint, laptopEndpoint} {
		if err := svc.Subscribe(ctx, practitioner, subscription(endpoint)); err != nil {
			t.Fatalf("subscribe: %v", err)
		}
	}
	if err := svc.Subscribe(ctx, other, subscription("https://web.push.apple.com/other")); err != nil {
		t.Fatalf("subscribe other: %v", err)
	}

	msg := domain.PushMessage{Title: "Nouvelle réservation", Body: "Marie D. · sam. 28 nov. 10:00", URL: "/admin"}
	if err := svc.NotifyUser(ctx, practitioner, msg); err != nil {
		t.Fatalf("notify: %v", err)
	}

	if len(sender.sent) != 2 {
		t.Fatalf("sent to %d devices, want the practitioner's 2", len(sender.sent))
	}
	var got domain.PushMessage
	if err := json.Unmarshal(sender.sent[iphoneEndpoint], &got); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if got != msg {
		t.Errorf("payload = %+v, want %+v", got, msg)
	}
}

func TestPushService_GoneSubscriptionsAreDeleted(t *testing.T) {
	ctx := context.Background()
	sender := newFakePushSender()
	svc, repo := newTestPushService(t, sender, "public-key")
	practitioner := uuid.New()
	for _, endpoint := range []string{iphoneEndpoint, laptopEndpoint} {
		if err := svc.Subscribe(ctx, practitioner, subscription(endpoint)); err != nil {
			t.Fatalf("subscribe: %v", err)
		}
	}
	sender.gone[laptopEndpoint] = true

	if err := svc.NotifyUser(ctx, practitioner, domain.PushMessage{Title: "t"}); err != nil {
		t.Fatalf("a gone subscription is not an error: %v", err)
	}
	if repo.count() != 1 {
		t.Fatalf("%d subscriptions left, want 1 (the gone one deleted)", repo.count())
	}
	if _, ok := sender.sent[iphoneEndpoint]; !ok {
		t.Error("the live device was not notified")
	}
}

func TestPushService_OtherFailuresKeepTheSubscriptionAndAreReturned(t *testing.T) {
	ctx := context.Background()
	sender := newFakePushSender()
	svc, repo := newTestPushService(t, sender, "public-key")
	practitioner := uuid.New()
	for _, endpoint := range []string{iphoneEndpoint, laptopEndpoint} {
		if err := svc.Subscribe(ctx, practitioner, subscription(endpoint)); err != nil {
			t.Fatalf("subscribe: %v", err)
		}
	}
	sender.failing[laptopEndpoint] = true

	err := svc.NotifyUser(ctx, practitioner, domain.PushMessage{Title: "t"})
	if err == nil {
		t.Fatal("want the failure returned")
	}
	if repo.count() != 2 {
		t.Errorf("a failing subscription must be kept, %d left", repo.count())
	}
	if _, ok := sender.sent[iphoneEndpoint]; !ok {
		t.Error("one failing device must not stop the others")
	}
}

func TestPushService_NoKeysMeansNothingSentNoError(t *testing.T) {
	ctx := context.Background()
	// No sender and no public key: Web Push is not configured.
	svc, repo := newTestPushService(t, nil, "")
	practitioner := uuid.New()

	if status := svc.Status(); status.Enabled || status.PublicKey != "" {
		t.Errorf("status = %+v, want off", status)
	}
	if err := svc.NotifyUser(ctx, practitioner, domain.PushMessage{Title: "t"}); err != nil {
		t.Errorf("notify with no keys: %v, want nil", err)
	}
	if err := svc.Subscribe(ctx, practitioner, subscription(iphoneEndpoint)); !errors.Is(err, domain.ErrPushUnavailable) {
		t.Errorf("subscribe with no keys: %v, want ErrPushUnavailable", err)
	}
	if repo.count() != 0 {
		t.Errorf("nothing may be stored with no keys")
	}
	// A device can still be switched off.
	if err := svc.Unsubscribe(ctx, practitioner, iphoneEndpoint); err != nil {
		t.Errorf("unsubscribe with no keys: %v", err)
	}
}

func TestPushService_StatusWithKeys(t *testing.T) {
	svc, _ := newTestPushService(t, newFakePushSender(), "BPublic")
	if status := svc.Status(); !status.Enabled || status.PublicKey != "BPublic" {
		t.Errorf("status = %+v", status)
	}
}

func TestPushService_SubscribeValidatesAndEncrypts(t *testing.T) {
	ctx := context.Background()
	svc, repo := newTestPushService(t, newFakePushSender(), "public-key")
	user := uuid.New()

	for name, input := range map[string]ports.PushSubscriptionInput{
		"no endpoint":   {P256dh: "k", Auth: "a"},
		"http endpoint": {Endpoint: "http://push.example/x", P256dh: "k", Auth: "a"},
		"no p256dh":     {Endpoint: iphoneEndpoint, Auth: "a"},
		"no auth":       {Endpoint: iphoneEndpoint, P256dh: "k"},
	} {
		if err := svc.Subscribe(ctx, user, input); !errors.Is(err, errs.ErrInvalidValue) {
			t.Errorf("%s: %v, want invalid value", name, err)
		}
	}

	if err := svc.Subscribe(ctx, user, subscription(iphoneEndpoint)); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	// The same device again replaces its row.
	if err := svc.Subscribe(ctx, user, subscription(iphoneEndpoint)); err != nil {
		t.Fatalf("resubscribe: %v", err)
	}
	rows, _ := repo.ListByUser(ctx, user)
	if len(rows) != 1 {
		t.Fatalf("%d rows, want one per device", len(rows))
	}
	if string(rows[0].EndpointEncrypted) == iphoneEndpoint || rows[0].EndpointHash == "" {
		t.Error("the endpoint must be stored encrypted, with its hash")
	}

	if err := svc.Unsubscribe(ctx, uuid.New(), iphoneEndpoint); err != nil {
		t.Fatalf("unsubscribe by another user: %v", err)
	}
	if repo.count() != 1 {
		t.Error("another user may not unsubscribe this device")
	}
	if err := svc.Unsubscribe(ctx, user, iphoneEndpoint); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if repo.count() != 0 {
		t.Error("the device was not unsubscribed")
	}
}
