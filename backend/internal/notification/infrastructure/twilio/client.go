package twilio

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/twilio/twilio-go"
	twilioclient "github.com/twilio/twilio-go/client"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/ports"
)

// Config is what the SMS adapter needs to send through Twilio. It
// authenticates with an API key (SID and secret) scoped to the account; the
// account's auth token is never used.
type Config struct {
	AccountSID   string
	APIKeySID    string
	APIKeySecret string

	// Region and Edge select Twilio's regional processing, e.g. "ie1" and
	// "dublin". Empty means Twilio's default (US1). An API key only works in
	// the region it was created in.
	Region string
	Edge   string

	// Sender is the From of every message: an alphanumeric sender ID
	// ("Leviosa") or a phone number in E.164 form.
	Sender string
}

// Complete reports whether every value needed to send is set. Region and
// edge are optional.
func (c Config) Complete() bool {
	return c.AccountSID != "" && c.APIKeySID != "" && c.APIKeySecret != "" && c.Sender != ""
}

// Option customises the client built by NewTwilioClient.
type Option func(*twilioclient.Client)

// WithHTTPClient makes the client send its requests through hc (tests point
// it at a fake Twilio API).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *twilioclient.Client) { c.HTTPClient = hc }
}

var _ ports.SMSService = (*TwilioClient)(nil)

type TwilioClient struct {
	client *twilio.RestClient
	from   string
}

// NewTwilioClient builds the SMS adapter from cfg. It refuses an incomplete
// configuration rather than letting the SDK fall back to its own environment
// variables (TWILIO_AUTH_TOKEN, TWILIO_REGION, TWILIO_EDGE).
func NewTwilioClient(cfg Config, opts ...Option) (*TwilioClient, error) {
	if !cfg.Complete() {
		return nil, errors.New("twilio: account SID, API key SID, API key secret and sender are all required")
	}

	// Building the base client ourselves (rather than passing Username and
	// Password in ClientParams) keeps the SDK from ever reading
	// TWILIO_AUTH_TOKEN, even for an empty value.
	base := &twilioclient.Client{
		Credentials: twilioclient.NewCredentials(cfg.APIKeySID, cfg.APIKeySecret),
	}
	base.SetAccountSid(cfg.AccountSID)
	for _, opt := range opts {
		opt(base)
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{Client: base})
	// Always set both, even when empty: the SDK otherwise defaults them from
	// TWILIO_REGION/TWILIO_EDGE, and empty must mean Twilio's default.
	client.SetRegion(cfg.Region)
	client.SetEdge(cfg.Edge)

	return &TwilioClient{
		client: client,
		from:   cfg.Sender,
	}, nil
}

func (c *TwilioClient) SendOTP(ctx context.Context, req domain.OTPSMSRequest) error {
	// Validate and normalize phone number
	smsReq, err := domain.NewSMSRequest(req.PhoneNumber, fmt.Sprintf("Your OTP code is: %s", req.OTP))
	if err != nil {
		return fmt.Errorf("invalid phone number for OTP SMS: %w", err)
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(smsReq.Phone)
	params.SetFrom(c.from)
	params.SetBody(smsReq.Message)

	_, err = c.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send OTP SMS via Twilio: %w", err)
	}

	return nil
}

func (c *TwilioClient) SendSMS(ctx context.Context, req domain.GenericSMSRequest) error {
	// Validate and normalize phone number
	smsReq, err := domain.NewSMSRequest(req.PhoneNumber, req.Message)
	if err != nil {
		return fmt.Errorf("invalid phone number for SMS: %w", err)
	}

	params := &openapi.CreateMessageParams{}
	params.SetTo(smsReq.Phone)
	params.SetFrom(c.from)
	params.SetBody(smsReq.Message)

	_, err = c.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS via Twilio: %w", err)
	}

	return nil
}
