package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"strings"

	"github.com/joho/godotenv"
	twilioRest "github.com/twilio/twilio-go/client"

	"github.com/Leviosa-care/leviosa/backend/internal/app"
	"github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	smtpClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/smtp"
	twilioClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/twilio"
)

// The check modes prove a provider's credentials end to end from wherever the
// binary runs (in deployed environments: inside the backend container). Each
// reads only its provider's settings — no database, Redis, RabbitMQ or KMS —
// sends one short message through the same adapter the app uses, prints a
// single success line, and otherwise fails (non-zero exit) with the
// provider's error code and message. Neither ever prints a credential.

const checkSMSBody = "Leviosa: test SMS from the check-sms command. No action needed."

// checkSMS sends one test SMS to args[0] through the configured Twilio
// settings (TWILIO_ACCOUNT_SID, TWILIO_API_KEY_SID, TWILIO_API_KEY_SECRET,
// TWILIO_REGION, TWILIO_EDGE, TWILIO_SENDER).
func checkSMS(ctx context.Context, args []string) error {
	return runCheckSMS(ctx, args, os.Stdout)
}

func runCheckSMS(ctx context.Context, args []string, out io.Writer, opts ...twilioClient.Option) error {
	if len(args) != 1 {
		return errors.New("usage: check-sms <to> (the recipient's phone number in E.164 form, e.g. +33612345678)")
	}
	to := args[0]

	loadDevelopmentEnv()
	cfg := app.TwilioConfigFromEnv()
	if missing := app.MissingTwilioEnv(cfg); len(missing) > 0 {
		return fmt.Errorf("check-sms: Twilio is not configured: missing %s", strings.Join(missing, ", "))
	}
	secrets := []string{cfg.AccountSID, cfg.APIKeySID, cfg.APIKeySecret}

	client, err := twilioClient.NewTwilioClient(cfg, opts...)
	if err != nil {
		return fmt.Errorf("check-sms: %s", redact(err.Error(), secrets...))
	}
	if err := client.SendSMS(ctx, domain.GenericSMSRequest{PhoneNumber: to, Message: checkSMSBody}); err != nil {
		return fmt.Errorf("check-sms: %s", redact(describeTwilioError(err), secrets...))
	}

	fmt.Fprintf(out, "check-sms: OK, Twilio accepted a test SMS to %s from %q (region %s, edge %s)\n",
		to, cfg.Sender, orDefault(cfg.Region), orDefault(cfg.Edge))
	return nil
}

// checkEmail sends one test email to args[0] through the configured SMTP
// settings (SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD).
func checkEmail(ctx context.Context, args []string) error {
	return runCheckEmail(ctx, args, os.Stdout)
}

func runCheckEmail(ctx context.Context, args []string, out io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: check-email <to> (the recipient's email address)")
	}
	to := args[0]

	loadDevelopmentEnv()
	cfg := app.SMTPConfigFromEnv()
	if missing := app.MissingSMTPEnv(cfg); len(missing) > 0 {
		return fmt.Errorf("check-email: SMTP is not configured: missing %s", strings.Join(missing, ", "))
	}

	if err := smtpClient.NewSMTPClient(cfg).SendCheckEmail(ctx, to); err != nil {
		return fmt.Errorf("check-email: %s", redact(describeSMTPError(err), cfg.Password))
	}

	fmt.Fprintf(out, "check-email: OK, %s:%d accepted a test email to %s from %s\n", cfg.Host, cfg.Port, to, cfg.Username)
	return nil
}

// describeTwilioError reduces a failed send to Twilio's error code and
// message when Twilio answered, and keeps the error as is otherwise (invalid
// recipient, network failure).
func describeTwilioError(err error) string {
	var restErr *twilioRest.TwilioRestError
	if !errors.As(err, &restErr) {
		return err.Error()
	}
	msg := fmt.Sprintf("Twilio error %d (HTTP %d): %s", restErr.Code, restErr.Status, restErr.Message)
	if restErr.MoreInfo != "" {
		msg += " (" + restErr.MoreInfo + ")"
	}
	return msg
}

// describeSMTPError reduces a failed send to the server's reply code and
// message when the server answered, and keeps the error as is otherwise.
func describeSMTPError(err error) string {
	var protoErr *textproto.Error
	if errors.As(err, &protoErr) {
		return fmt.Sprintf("SMTP error %d: %s", protoErr.Code, protoErr.Msg)
	}
	return err.Error()
}

// redact removes every configured credential from s, so a failure can never
// print one even if a library or provider echoes it back.
func redact(s string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			s = strings.ReplaceAll(s, secret, "[redacted]")
		}
	}
	return s
}

func orDefault(s string) string {
	if s == "" {
		return "default"
	}
	return s
}

// loadDevelopmentEnv applies the serve mode's development convenience: when
// ENVIRONMENT is unset or "development" it loads development.env, never
// overriding variables already set.
func loadDevelopmentEnv() {
	if env := os.Getenv("ENVIRONMENT"); env != "" && env != "development" {
		return
	}
	if err := godotenv.Load("development.env"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load development.env: %v\n", err)
	}
}
