package webpush

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	webpushgo "github.com/SherClockHolmes/webpush-go"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

func testConfig(t *testing.T) Config {
	t.Helper()
	private, public, err := webpushgo.GenerateVAPIDKeys()
	if err != nil {
		t.Fatalf("vapid keys: %v", err)
	}
	return Config{PublicKey: public, PrivateKey: private, Subject: "mailto:contact@leviosa.care"}
}

// deviceSubscription is a subscription with real device keys, as a browser
// gives it, pointing at endpoint.
func deviceSubscription(t *testing.T, endpoint string) domain.PushSubscription {
	t.Helper()
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("device key: %v", err)
	}
	auth := make([]byte, 16)
	_, _ = rand.Read(auth)
	return domain.PushSubscription{
		Endpoint: endpoint,
		P256dh:   base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()),
		Auth:     base64.RawURLEncoding.EncodeToString(auth),
	}
}

func TestSender_PostsAnEncryptedVAPIDRequest(t *testing.T) {
	var got *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	sender, err := NewSender(testConfig(t))
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	sender.WithHTTPClient(server.Client())

	if err := sender.Send(context.Background(), deviceSubscription(t, server.URL+"/push/abc"), []byte(`{"title":"Nouvelle réservation"}`)); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got == nil {
		t.Fatal("no request reached the push service")
	}
	if got.Header.Get("Content-Encoding") != "aes128gcm" || got.Header.Get("TTL") == "" || got.Header.Get("Urgency") != "high" {
		t.Errorf("headers = %v", got.Header)
	}

	// The VAPID JWT's subject is the configured mailto: address, once.
	authz := got.Header.Get("Authorization")
	if !strings.HasPrefix(authz, "vapid t=") {
		t.Fatalf("authorization = %q", authz)
	}
	jwt := strings.SplitN(strings.TrimPrefix(authz, "vapid t="), ",", 2)[0]
	claimsPart := strings.Split(jwt, ".")[1]
	raw, err := base64.RawURLEncoding.DecodeString(claimsPart)
	if err != nil {
		t.Fatalf("jwt claims: %v", err)
	}
	var claims map[string]any
	_ = json.Unmarshal(raw, &claims)
	if claims["sub"] != "mailto:contact@leviosa.care" {
		t.Errorf("sub = %v, want mailto:contact@leviosa.care", claims["sub"])
	}
}

func TestSender_GoneAndFailures(t *testing.T) {
	for _, tc := range []struct {
		status   int
		wantGone bool
		wantErr  bool
	}{
		{http.StatusNotFound, true, true},
		{http.StatusGone, true, true},
		{http.StatusInternalServerError, false, true},
		{http.StatusTooManyRequests, false, true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		}))
		sender, _ := NewSender(testConfig(t))
		sender.WithHTTPClient(server.Client())
		endpoint := server.URL + "/push/secret-capability"

		err := sender.Send(context.Background(), deviceSubscription(t, endpoint), []byte(`{}`))
		server.Close()

		if (err != nil) != tc.wantErr || errors.Is(err, domain.ErrPushSubscriptionGone) != tc.wantGone {
			t.Errorf("status %d: err = %v", tc.status, err)
		}
		if err != nil && strings.Contains(err.Error(), "secret-capability") {
			t.Errorf("status %d: the endpoint leaked into the error: %v", tc.status, err)
		}
	}
}

func TestSender_NetworkErrorHidesTheEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	endpoint := server.URL + "/push/secret-capability"
	server.Close() // nothing listens any more

	sender, _ := NewSender(testConfig(t))
	err := sender.Send(context.Background(), deviceSubscription(t, endpoint), []byte(`{}`))
	if err == nil {
		t.Fatal("want a network error")
	}
	if strings.Contains(err.Error(), "secret-capability") {
		t.Errorf("the endpoint leaked into the error: %v", err)
	}
}

func TestConfig(t *testing.T) {
	if (Config{}).Complete() || (Config{}).Partial() {
		t.Error("an empty config is off, and not partial")
	}
	if !(Config{PublicKey: "p"}).Partial() {
		t.Error("one value set is partial")
	}
	if _, err := NewSender(Config{PublicKey: "p", PrivateKey: "k", Subject: "contact@leviosa.care"}); err == nil {
		t.Error("a subject without mailto: or https: is refused")
	}
}
