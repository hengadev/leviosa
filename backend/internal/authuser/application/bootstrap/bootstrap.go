// Package bootstrap implements the `bootstrap-admin` run mode's use case:
// creating the first superadmin from operator-specified ADMIN_* configuration
// on a freshly deployed environment, without shell access and without the
// admin password ever reaching the running app (issue 013).
//
// The account is created only when no user with ADMIN_EMAIL exists; an
// existing account is never modified — neither its password nor its role —
// so re-running the mode on every deploy is safe.
package bootstrap

import (
	"fmt"
	"strings"
	"unicode"
)

// AdminPasswordMinLength is the minimum length for the superadmin's
// password. It is deliberately stronger than the app-wide client minimum
// (8, see domain.PasswordMinLength): the bootstrap password is chosen once
// by the operator in configuration and grants full platform access.
const AdminPasswordMinLength = 12

// ValidateAdminPassword enforces the superadmin password strength rule:
// at least AdminPasswordMinLength characters, containing at least one
// lowercase letter, one uppercase letter and one digit. It is checked
// outside development before anything is written, so a weak ADMIN_PASSWORD
// makes the bootstrap-admin mode exit non-zero.
func ValidateAdminPassword(password string) error {
	if len(password) < AdminPasswordMinLength {
		return fmt.Errorf("ADMIN_PASSWORD must be at least %d characters long", AdminPasswordMinLength)
	}

	var hasLower, hasUpper, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLower || !hasUpper || !hasDigit {
		return fmt.Errorf("ADMIN_PASSWORD must contain at least one lowercase letter, one uppercase letter and one digit")
	}

	return nil
}

// Config holds the superadmin's values, read from ADMIN_* environment
// variables by LoadConfig.
type Config struct {
	Email      string
	Password   string
	FirstName  string
	LastName   string
	Telephone  string
	PostalCode string
	City       string
	Address1   string
	Address2   string
}

// adminKeys maps each ADMIN_* variable to the Config field it fills, in a
// stable order for error messages. Every value is required outside
// development except ADMIN_ADDRESS2.
var adminKeys = []struct {
	env      string
	field    func(Config) string
	assign   func(*Config, string)
	optional bool
}{
	{"ADMIN_EMAIL", func(c Config) string { return c.Email }, func(c *Config, v string) { c.Email = v }, false},
	{"ADMIN_PASSWORD", func(c Config) string { return c.Password }, func(c *Config, v string) { c.Password = v }, false},
	{"ADMIN_FIRST_NAME", func(c Config) string { return c.FirstName }, func(c *Config, v string) { c.FirstName = v }, false},
	{"ADMIN_LAST_NAME", func(c Config) string { return c.LastName }, func(c *Config, v string) { c.LastName = v }, false},
	{"ADMIN_TELEPHONE", func(c Config) string { return c.Telephone }, func(c *Config, v string) { c.Telephone = v }, false},
	{"ADMIN_POSTAL_CODE", func(c Config) string { return c.PostalCode }, func(c *Config, v string) { c.PostalCode = v }, false},
	{"ADMIN_CITY", func(c Config) string { return c.City }, func(c *Config, v string) { c.City = v }, false},
	{"ADMIN_ADDRESS1", func(c Config) string { return c.Address1 }, func(c *Config, v string) { c.Address1 = v }, false},
	{"ADMIN_ADDRESS2", func(c Config) string { return c.Address2 }, func(c *Config, v string) { c.Address2 = v }, true},
}

// devDefaults are the non-secret values used in development when an ADMIN_*
// variable is not set. They exist only so local environments bootstrap
// without configuration; every deployed environment must set its own.
var devDefaults = map[string]string{
	"ADMIN_EMAIL":       "admin@example.com",
	"ADMIN_PASSWORD":    "ChangeMe1234!",
	"ADMIN_FIRST_NAME":  "Admin",
	"ADMIN_LAST_NAME":   "User",
	"ADMIN_TELEPHONE":   "+33600000000",
	"ADMIN_POSTAL_CODE": "75001",
	"ADMIN_CITY":        "Paris",
	"ADMIN_ADDRESS1":    "1 Rue Example",
	"ADMIN_ADDRESS2":    "",
}

// LoadConfig builds the superadmin's Config from the ADMIN_* values read
// through getenv (os.Getenv in the binary).
//
// Outside development (ENVIRONMENT set to anything but "development", which
// is also the default when unset) every value except ADMIN_ADDRESS2 is
// required and the password must satisfy the strength rule below; a missing
// value or a weak password is an error, so the bootstrap-admin mode exits
// non-zero instead of silently creating a broken or weak account. In
// development, missing values fall back to the non-secret defaults above.
func LoadConfig(getenv func(string) string) (Config, error) {
	development := getenv("ENVIRONMENT") == "" || getenv("ENVIRONMENT") == "development"

	var cfg Config
	var missing []string
	for _, key := range adminKeys {
		value := strings.TrimSpace(getenv(key.env))
		if value == "" {
			if development {
				value = devDefaults[key.env]
			} else if !key.optional {
				missing = append(missing, key.env)
			}
		}
		key.assign(&cfg, value)
	}

	if !development && len(missing) > 0 {
		return Config{}, fmt.Errorf(
			"missing required values: %s (every ADMIN_* value except ADMIN_ADDRESS2 must be set outside development; they belong in the Infisical /admin folder, rendered into the bootstrap's own env file)",
			strings.Join(missing, ", "),
		)
	}

	if !development {
		if err := ValidateAdminPassword(cfg.Password); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}
