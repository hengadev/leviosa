package twilio

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	twilioclient "github.com/twilio/twilio-go/client"

	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
)

// fakeTwilio stands in for the Twilio REST API. Its HTTP client sends every
// request to the fake server while keeping the original Host header, so a
// test can see which Twilio host (region/edge) the SDK targeted.
type fakeTwilio struct {
	srv    *httptest.Server
	status int
	body   string

	mu   sync.Mutex
	reqs []*http.Request
	form []url.Values
}

func newFakeTwilio(t *testing.T, status int, body string) *fakeTwilio {
	t.Helper()
	f := &fakeTwilio{status: status, body: body}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		f.mu.Lock()
		f.reqs = append(f.reqs, r)
		f.form = append(f.form, r.PostForm)
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		_, _ = w.Write([]byte(f.body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeTwilio) httpClient() *http.Client {
	target, _ := url.Parse(f.srv.URL)
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r = r.Clone(r.Context())
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host // r.Host keeps the Twilio host the SDK built
		return http.DefaultTransport.RoundTrip(r)
	})}
}

func (f *fakeTwilio) onlyRequest(t *testing.T) (*http.Request, url.Values) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.reqs) != 1 {
		t.Fatalf("expected exactly one request to Twilio, got %d", len(f.reqs))
	}
	return f.reqs[0], f.form[0]
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

const createdMessage = `{"sid":"SM00000000000000000000000000000000","status":"queued"}`

func testConfig() Config {
	return Config{
		AccountSID:   "ACaccount0000000000000000000000000",
		APIKeySID:    "SKapikey00000000000000000000000000",
		APIKeySecret: "apikeysecret0000000000000000000",
		Sender:       "Leviosa",
	}
}

func sendTestSMS(t *testing.T, cfg Config, fake *fakeTwilio) error {
	t.Helper()
	c, err := NewTwilioClient(cfg, WithHTTPClient(fake.httpClient()))
	if err != nil {
		t.Fatalf("NewTwilioClient: %v", err)
	}
	return c.SendSMS(context.Background(), domain.GenericSMSRequest{PhoneNumber: "+33612345678", Message: "hello"})
}

func TestTwilioClientAuthenticatesWithAPIKey(t *testing.T) {
	// The account's auth token must never be used, even when the SDK's own
	// fallback variables are present in the environment.
	t.Setenv("TWILIO_AUTH_TOKEN", "authtokenmustnotbeused")
	t.Setenv("TWILIO_ACCOUNT_SID", "ACfromenvironment")

	fake := newFakeTwilio(t, http.StatusCreated, createdMessage)
	cfg := testConfig()

	if err := sendTestSMS(t, cfg, fake); err != nil {
		t.Fatalf("SendSMS: %v", err)
	}

	req, form := fake.onlyRequest(t)
	user, pass, ok := req.BasicAuth()
	if !ok {
		t.Fatal("expected HTTP basic auth on the Twilio request")
	}
	if user != cfg.APIKeySID || pass != cfg.APIKeySecret {
		t.Fatalf("expected basic auth with the API key SID and secret, got user %q", user)
	}
	if wantPath := "/2010-04-01/Accounts/" + cfg.AccountSID + "/Messages.json"; req.URL.Path != wantPath {
		t.Fatalf("expected the message to be created under the account SID: path %q, want %q", req.URL.Path, wantPath)
	}
	if got := form.Get("To"); got != "+33612345678" {
		t.Fatalf("To = %q", got)
	}
	if got := form.Get("Body"); got != "hello" {
		t.Fatalf("Body = %q", got)
	}
}

func TestTwilioClientAppliesRegionAndEdgeWhenSet(t *testing.T) {
	fake := newFakeTwilio(t, http.StatusCreated, createdMessage)
	cfg := testConfig()
	cfg.Region = "ie1"
	cfg.Edge = "dublin"

	if err := sendTestSMS(t, cfg, fake); err != nil {
		t.Fatalf("SendSMS: %v", err)
	}

	req, _ := fake.onlyRequest(t)
	if req.Host != "api.dublin.ie1.twilio.com" {
		t.Fatalf("expected the IE1 Dublin host, got %q", req.Host)
	}
}

func TestTwilioClientOmitsRegionAndEdgeWhenEmpty(t *testing.T) {
	// The SDK picks these up from the environment on its own; the configured
	// (empty) values must win so that empty really means Twilio's default.
	t.Setenv("TWILIO_REGION", "au1")
	t.Setenv("TWILIO_EDGE", "sydney")

	fake := newFakeTwilio(t, http.StatusCreated, createdMessage)

	if err := sendTestSMS(t, testConfig(), fake); err != nil {
		t.Fatalf("SendSMS: %v", err)
	}

	req, _ := fake.onlyRequest(t)
	if req.Host != "api.twilio.com" {
		t.Fatalf("expected Twilio's default host, got %q", req.Host)
	}
}

func TestTwilioClientSendsFromConfiguredSender(t *testing.T) {
	for _, sender := range []string{"Leviosa", "+33757000000"} {
		t.Run(sender, func(t *testing.T) {
			fake := newFakeTwilio(t, http.StatusCreated, createdMessage)
			cfg := testConfig()
			cfg.Sender = sender

			if err := sendTestSMS(t, cfg, fake); err != nil {
				t.Fatalf("SendSMS: %v", err)
			}

			_, form := fake.onlyRequest(t)
			if got := form.Get("From"); got != sender {
				t.Fatalf("From = %q, want %q", got, sender)
			}
		})
	}
}

func TestNewTwilioClientRequiresAPIKeyCredentialsAndSender(t *testing.T) {
	for name, mutate := range map[string]func(*Config){
		"account SID":    func(c *Config) { c.AccountSID = "" },
		"API key SID":    func(c *Config) { c.APIKeySID = "" },
		"API key secret": func(c *Config) { c.APIKeySecret = "" },
		"sender":         func(c *Config) { c.Sender = "" },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := testConfig()
			mutate(&cfg)
			if cfg.Complete() {
				t.Fatal("expected Complete() to be false")
			}
			if _, err := NewTwilioClient(cfg); err == nil {
				t.Fatal("expected an error for an incomplete configuration")
			} else if cfg.APIKeySecret != "" && strings.Contains(err.Error(), cfg.APIKeySecret) {
				t.Fatalf("error leaks the API key secret: %v", err)
			}
		})
	}

	if !testConfig().Complete() {
		t.Fatal("expected a full configuration (without region/edge) to be complete")
	}
}

func TestTwilioClientReturnsProviderError(t *testing.T) {
	fake := newFakeTwilio(t, http.StatusUnauthorized,
		`{"code":20003,"message":"Authenticate","more_info":"https://www.twilio.com/docs/errors/20003","status":401}`)

	err := sendTestSMS(t, testConfig(), fake)
	if err == nil {
		t.Fatal("expected an error")
	}
	var restErr *twilioclient.TwilioRestError
	if !errors.As(err, &restErr) {
		t.Fatalf("expected a wrapped TwilioRestError, got %T: %v", err, err)
	}
	if restErr.Code != 20003 || restErr.Message != "Authenticate" {
		t.Fatalf("unexpected provider error: %+v", restErr)
	}
}
