// Package webpush sends practitioner notifications through Web Push with
// VAPID (RFC 8030, RFC 8291, RFC 8292), the way an iPhone home-screen web
// app receives them (ADR-0020).
package webpush

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	webpushgo "github.com/SherClockHolmes/webpush-go"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"
)

// Config holds the VAPID key pair and subject, from VAPID_PUBLIC_KEY,
// VAPID_PRIVATE_KEY and VAPID_SUBJECT (Infisical /webpush, ADR-0017). Any of
// them empty: Web Push is off.
type Config struct {
	PublicKey  string
	PrivateKey string
	// Subject is the contact the push services may reach the sender at:
	// "mailto:contact@leviosa.care" or an https URL.
	Subject string
}

// Complete reports whether every value is set, so Web Push can be on.
func (c Config) Complete() bool {
	return strings.TrimSpace(c.PublicKey) != "" &&
		strings.TrimSpace(c.PrivateKey) != "" &&
		strings.TrimSpace(c.Subject) != ""
}

// Partial reports whether some values are set but not all: a configuration
// mistake that leaves Web Push off.
func (c Config) Partial() bool {
	set := 0
	for _, v := range []string{c.PublicKey, c.PrivateKey, c.Subject} {
		if strings.TrimSpace(v) != "" {
			set++
		}
	}
	return set > 0 && set < 3
}

const (
	// defaultTTL is how long a push service keeps a notification for a
	// device that is offline: a day, so a phone switched on in the morning
	// still gets the evening's bookings.
	defaultTTL = 24 * 60 * 60
	// requestTimeout bounds one request to a push service.
	requestTimeout = 15 * time.Second
	// recordSize is the size of the encrypted body: webpush-go pads every
	// payload to it (4096 by default, the most push services accept).
	// Notifications are a few hundred bytes, so half of it is plenty.
	recordSize = 2048
)

// Sender delivers payloads to push services.
type Sender struct {
	publicKey  string
	privateKey string
	subscriber string
	client     webpushgo.HTTPClient
}

var _ ports.PushSender = (*Sender)(nil)

// NewSender builds a sender from a complete configuration.
func NewSender(cfg Config) (*Sender, error) {
	if !cfg.Complete() {
		return nil, errors.New("web push: VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY and VAPID_SUBJECT are required")
	}
	subject := strings.TrimSpace(cfg.Subject)
	if !strings.HasPrefix(subject, "mailto:") && !strings.HasPrefix(subject, "https:") {
		return nil, errors.New("web push: VAPID_SUBJECT must be a mailto: address or an https URL")
	}
	return &Sender{
		publicKey:  strings.TrimSpace(cfg.PublicKey),
		privateKey: strings.TrimSpace(cfg.PrivateKey),
		// webpush-go prefixes "mailto:" to any subscriber that is not an
		// https URL: give it the bare address.
		subscriber: strings.TrimPrefix(subject, "mailto:"),
		client:     &http.Client{Timeout: requestTimeout},
	}, nil
}

// WithHTTPClient replaces the HTTP client (tests).
func (s *Sender) WithHTTPClient(client webpushgo.HTTPClient) *Sender {
	s.client = client
	return s
}

// Send encrypts payload for the device and posts it to its push service.
// 404 and 410 mean the subscription is gone (domain.ErrPushSubscriptionGone).
// The endpoint, a bearer capability, never appears in a returned error.
func (s *Sender) Send(ctx context.Context, sub domain.PushSubscription, payload []byte) error {
	resp, err := webpushgo.SendNotificationWithContext(ctx, payload, &webpushgo.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpushgo.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpushgo.Options{
		HTTPClient:      s.client,
		Subscriber:      s.subscriber,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		RecordSize:      recordSize,
		TTL:             defaultTTL,
		Urgency:         webpushgo.UrgencyHigh,
	})
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			// url.Error carries the endpoint: keep only the cause.
			return fmt.Errorf("web push request failed: %w", urlErr.Err)
		}
		return fmt.Errorf("web push: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return domain.ErrPushSubscriptionGone
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		return fmt.Errorf("web push: push service answered %d", resp.StatusCode)
	}
	return nil
}
