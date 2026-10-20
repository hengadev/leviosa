package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	twilioClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/twilio"
)

// --- check-sms ---------------------------------------------------------------

const (
	testAccountSID   = "ACcheck0000000000000000000000000000"
	testAPIKeySID    = "SKcheck0000000000000000000000000000"
	testAPIKeySecret = "checkapikeysecret000000000000000"
	testAuthToken    = "checkauthtoken00000000000000000"
)

// setCheckSMSEnv configures Twilio for a check-sms run. The account's auth
// token is set too, to prove it is neither used nor printed.
func setCheckSMSEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ENVIRONMENT", "production") // deterministic: no development.env loading
	t.Setenv("TWILIO_ACCOUNT_SID", testAccountSID)
	t.Setenv("TWILIO_API_KEY_SID", testAPIKeySID)
	t.Setenv("TWILIO_API_KEY_SECRET", testAPIKeySecret)
	t.Setenv("TWILIO_REGION", "ie1")
	t.Setenv("TWILIO_EDGE", "dublin")
	t.Setenv("TWILIO_SENDER", "Leviosa")
	t.Setenv("TWILIO_AUTH_TOKEN", testAuthToken)
}

// fakeTwilioAPI stands in for Twilio's REST API: every request the SDK makes
// is sent to it, with the original Host header kept.
type fakeTwilioAPI struct {
	srv    *httptest.Server
	status int
	body   string

	mu   sync.Mutex
	reqs []*http.Request
}

func newFakeTwilioAPI(t *testing.T, status int, body string) *fakeTwilioAPI {
	t.Helper()
	f := &fakeTwilioAPI{status: status, body: body}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		f.reqs = append(f.reqs, r)
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		_, _ = w.Write([]byte(f.body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeTwilioAPI) option() twilioClient.Option {
	target, _ := url.Parse(f.srv.URL)
	return twilioClient.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r = r.Clone(r.Context())
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		return http.DefaultTransport.RoundTrip(r)
	})})
}

func (f *fakeTwilioAPI) requests() []*http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]*http.Request(nil), f.reqs...)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func assertNoSecrets(t *testing.T, output string, secrets ...string) {
	t.Helper()
	for _, s := range secrets {
		if s != "" && strings.Contains(output, s) {
			t.Fatalf("output leaks a configured secret value %q:\n%s", s, output)
		}
	}
}

func TestCheckSMSRequiresExactlyOneRecipient(t *testing.T) {
	for _, args := range [][]string{{"check-sms"}, {"check-sms", "+33612345678", "extra"}} {
		err := dispatch(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "usage: check-sms <to>") {
			t.Fatalf("%v: expected a usage error, got %v", args, err)
		}
	}
}

func TestCheckSMSFailsWhenTwilioIsNotConfigured(t *testing.T) {
	setCheckSMSEnv(t)
	t.Setenv("TWILIO_API_KEY_SID", "")
	t.Setenv("TWILIO_SENDER", "")

	var out bytes.Buffer
	err := runCheckSMS(context.Background(), []string{"+33612345678"}, &out)
	if err == nil {
		t.Fatal("expected check-sms to fail without a complete Twilio configuration")
	}
	if !strings.Contains(err.Error(), "TWILIO_API_KEY_SID") || !strings.Contains(err.Error(), "TWILIO_SENDER") {
		t.Fatalf("expected the missing variables to be named, got %v", err)
	}
	assertNoSecrets(t, out.String()+err.Error(), testAPIKeySecret, testAuthToken)
}

func TestCheckSMSSucceedsThroughTheConfiguredTwilio(t *testing.T) {
	setCheckSMSEnv(t)
	fake := newFakeTwilioAPI(t, http.StatusCreated, `{"sid":"SM00000000000000000000000000000000","status":"queued"}`)

	var out bytes.Buffer
	if err := runCheckSMS(context.Background(), []string{"+33612345678"}, &out, fake.option()); err != nil {
		t.Fatalf("expected check-sms to succeed, got %v", err)
	}

	reqs := fake.requests()
	if len(reqs) != 1 {
		t.Fatalf("expected one request to Twilio, got %d", len(reqs))
	}
	req := reqs[0]
	if user, pass, _ := req.BasicAuth(); user != testAPIKeySID || pass != testAPIKeySecret {
		t.Fatalf("expected API-key basic auth, got user %q", user)
	}
	if req.Host != "api.dublin.ie1.twilio.com" {
		t.Fatalf("expected the IE1 Dublin host, got %q", req.Host)
	}
	if req.PostForm.Get("To") != "+33612345678" || req.PostForm.Get("From") != "Leviosa" || req.PostForm.Get("Body") == "" {
		t.Fatalf("unexpected message: %v", req.PostForm)
	}

	if !strings.HasPrefix(out.String(), "check-sms: OK") || !strings.Contains(out.String(), "+33612345678") {
		t.Fatalf("unexpected success output: %q", out.String())
	}
	assertNoSecrets(t, out.String(), testAccountSID, testAPIKeySID, testAPIKeySecret, testAuthToken)
}

func TestCheckSMSReportsTheProviderErrorWithoutSecrets(t *testing.T) {
	setCheckSMSEnv(t)
	fake := newFakeTwilioAPI(t, http.StatusUnauthorized,
		`{"code":20003,"message":"Authenticate","more_info":"https://www.twilio.com/docs/errors/20003","status":401}`)

	var out bytes.Buffer
	err := runCheckSMS(context.Background(), []string{"+33612345678"}, &out, fake.option())
	if err == nil {
		t.Fatal("expected check-sms to fail on a Twilio error")
	}
	if !strings.Contains(err.Error(), "20003") || !strings.Contains(err.Error(), "Authenticate") {
		t.Fatalf("expected Twilio's error code and message, got %v", err)
	}
	assertNoSecrets(t, out.String()+err.Error(), testAccountSID, testAPIKeySID, testAPIKeySecret, testAuthToken)
}

func TestCheckSMSRedactsCredentialsFromTransportErrors(t *testing.T) {
	setCheckSMSEnv(t)
	fake := newFakeTwilioAPI(t, http.StatusCreated, `{}`)
	fake.srv.Close() // the request fails before reaching "Twilio"; the error carries the request URL

	var out bytes.Buffer
	err := runCheckSMS(context.Background(), []string{"+33612345678"}, &out, fake.option())
	if err == nil {
		t.Fatal("expected check-sms to fail when Twilio is unreachable")
	}
	assertNoSecrets(t, out.String()+err.Error(), testAccountSID, testAPIKeySID, testAPIKeySecret, testAuthToken)
}

// --- check-email -------------------------------------------------------------

const (
	testSMTPUser     = "noreply@example.com"
	testSMTPPassword = "checksmtppassword000"
)

type fakeMail struct {
	from string
	to   []string
	data string
}

// fakeSMTP is a minimal in-process SMTP server: EHLO advertising AUTH PLAIN,
// AUTH, MAIL, RCPT, DATA and QUIT. It accepts only user/pass.
type fakeSMTP struct {
	host string
	port int
	user string
	pass string

	mu    sync.Mutex
	mails []fakeMail
}

func newFakeSMTP(t *testing.T, user, pass string) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	addr := ln.Addr().(*net.TCPAddr)
	f := &fakeSMTP{host: "127.0.0.1", port: addr.Port, user: user, pass: pass}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(conn)
		}
	}()
	return f
}

func (f *fakeSMTP) serve(conn net.Conn) {
	defer conn.Close()
	tp := textproto.NewConn(conn)
	reply := func(line string) { _ = tp.PrintfLine("%s", line) }

	reply("220 fake.smtp ESMTP")
	var mail fakeMail
	for {
		line, err := tp.ReadLine()
		if err != nil {
			return
		}
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"):
			reply("250-fake.smtp")
			reply("250 AUTH PLAIN")
		case strings.HasPrefix(upper, "AUTH PLAIN "):
			raw, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(line[len("AUTH PLAIN "):]))
			parts := strings.Split(string(raw), "\x00")
			if len(parts) == 3 && parts[1] == f.user && parts[2] == f.pass {
				reply("235 2.7.0 Authentication successful")
			} else {
				reply("535 5.7.8 Authentication credentials invalid")
			}
		case strings.HasPrefix(upper, "MAIL FROM:"):
			mail = fakeMail{from: strings.Trim(line[len("MAIL FROM:"):], "<> ")}
			reply("250 2.1.0 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			mail.to = append(mail.to, strings.Trim(line[len("RCPT TO:"):], "<> "))
			reply("250 2.1.5 OK")
		case upper == "DATA":
			reply("354 Start mail input; end with <CRLF>.<CRLF>")
			data, err := tp.ReadDotBytes()
			if err != nil {
				return
			}
			mail.data = string(data)
			f.mu.Lock()
			f.mails = append(f.mails, mail)
			f.mu.Unlock()
			reply("250 2.0.0 OK queued")
		case upper == "QUIT":
			reply("221 2.0.0 Bye")
			return
		case upper == "*":
			reply("501 5.7.0 Authentication cancelled")
		default:
			reply("502 5.5.2 Command not recognized")
		}
	}
}

func (f *fakeSMTP) received() []fakeMail {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]fakeMail(nil), f.mails...)
}

func setCheckEmailEnv(t *testing.T, server *fakeSMTP, password string) {
	t.Helper()
	t.Setenv("ENVIRONMENT", "production") // deterministic: no development.env loading
	t.Setenv("SMTP_HOST", server.host)
	t.Setenv("SMTP_PORT", strconv.Itoa(server.port))
	t.Setenv("SMTP_USERNAME", testSMTPUser)
	t.Setenv("SMTP_PASSWORD", password)
}

func TestCheckEmailRequiresExactlyOneRecipient(t *testing.T) {
	for _, args := range [][]string{{"check-email"}, {"check-email", "a@example.com", "b@example.com"}} {
		err := dispatch(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "usage: check-email <to>") {
			t.Fatalf("%v: expected a usage error, got %v", args, err)
		}
	}
}

func TestCheckEmailFailsWhenSMTPIsNotConfigured(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")

	var out bytes.Buffer
	err := runCheckEmail(context.Background(), []string{"someone@example.com"}, &out)
	if err == nil || !strings.Contains(err.Error(), "SMTP_USERNAME") || !strings.Contains(err.Error(), "SMTP_PASSWORD") {
		t.Fatalf("expected the missing SMTP variables to be named, got %v", err)
	}
}

func TestCheckEmailSucceedsThroughTheConfiguredSMTPServer(t *testing.T) {
	server := newFakeSMTP(t, testSMTPUser, testSMTPPassword)
	setCheckEmailEnv(t, server, testSMTPPassword)

	var out bytes.Buffer
	if err := runCheckEmail(context.Background(), []string{"someone@example.com"}, &out); err != nil {
		t.Fatalf("expected check-email to succeed, got %v", err)
	}

	mails := server.received()
	if len(mails) != 1 {
		t.Fatalf("expected one email, got %d", len(mails))
	}
	mail := mails[0]
	if mail.from != testSMTPUser || len(mail.to) != 1 || mail.to[0] != "someone@example.com" {
		t.Fatalf("unexpected envelope: from %q to %v", mail.from, mail.to)
	}
	if !strings.Contains(mail.data, "Subject: Leviosa SMTP check") {
		t.Fatalf("expected the check subject in the message, got:\n%s", mail.data)
	}

	if !strings.HasPrefix(out.String(), "check-email: OK") || !strings.Contains(out.String(), "someone@example.com") {
		t.Fatalf("unexpected success output: %q", out.String())
	}
	assertNoSecrets(t, out.String(), testSMTPPassword)
}

func TestCheckEmailReportsTheProviderErrorWithoutSecrets(t *testing.T) {
	server := newFakeSMTP(t, testSMTPUser, "theserverexpectsanotherpassword")
	setCheckEmailEnv(t, server, testSMTPPassword)

	var out bytes.Buffer
	err := runCheckEmail(context.Background(), []string{"someone@example.com"}, &out)
	if err == nil {
		t.Fatal("expected check-email to fail on rejected credentials")
	}
	if !strings.Contains(err.Error(), "535") || !strings.Contains(err.Error(), "Authentication credentials invalid") {
		t.Fatalf("expected the SMTP error code and message, got %v", err)
	}
	if len(server.received()) != 0 {
		t.Fatal("no email should have been accepted")
	}
	assertNoSecrets(t, out.String()+err.Error(), testSMTPPassword)
}
