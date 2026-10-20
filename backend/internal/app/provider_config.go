package app

import (
	smtpClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/smtp"
	twilioClient "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/twilio"
)

// TwilioConfigFromEnv reads the Twilio settings. The client authenticates with
// an API key (TWILIO_API_KEY_SID/TWILIO_API_KEY_SECRET) on the account
// TWILIO_ACCOUNT_SID; the account's auth token is never read. TWILIO_REGION
// and TWILIO_EDGE are optional (empty means Twilio's default, US1; deployed
// environments use ie1/dublin). TWILIO_SENDER is an alphanumeric sender ID or
// an E.164 number.
//
// LoadConfig uses it, and so does the check-sms run mode, which must work
// without the rest of the app's configuration.
func TwilioConfigFromEnv() twilioClient.Config {
	return twilioClient.Config{
		AccountSID:   getEnv("TWILIO_ACCOUNT_SID", ""),
		APIKeySID:    getEnv("TWILIO_API_KEY_SID", ""),
		APIKeySecret: getEnv("TWILIO_API_KEY_SECRET", ""),
		Region:       getEnv("TWILIO_REGION", ""),
		Edge:         getEnv("TWILIO_EDGE", ""),
		Sender:       getEnv("TWILIO_SENDER", ""),
	}
}

// MissingTwilioEnv names the required Twilio variables that are empty in cfg.
func MissingTwilioEnv(cfg twilioClient.Config) []string {
	return missing(
		"TWILIO_ACCOUNT_SID", cfg.AccountSID,
		"TWILIO_API_KEY_SID", cfg.APIKeySID,
		"TWILIO_API_KEY_SECRET", cfg.APIKeySecret,
		"TWILIO_SENDER", cfg.Sender,
	)
}

// SMTPConfigFromEnv reads the SMTP settings. LoadConfig uses it, and so does
// the check-email run mode, which must work without the rest of the app's
// configuration.
func SMTPConfigFromEnv() smtpClient.SMTPConfig {
	return smtpClient.SMTPConfig{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     getEnvInt("SMTP_PORT", 587),
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
	}
}

// MissingSMTPEnv names the SMTP credentials that are empty in cfg (the host
// and port have defaults).
func MissingSMTPEnv(cfg smtpClient.SMTPConfig) []string {
	return missing(
		"SMTP_USERNAME", cfg.Username,
		"SMTP_PASSWORD", cfg.Password,
	)
}

// missing takes name/value pairs and returns the names whose value is empty.
func missing(pairs ...string) []string {
	var names []string
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] == "" {
			names = append(names, pairs[i])
		}
	}
	return names
}
