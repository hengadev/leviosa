package bootstrap_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrap "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/bootstrap"
)

// envMap adapts a map to the getenv function LoadConfig reads its values
// through (os.Getenv in the real binary).
func envMap(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}

// completeAdminEnv returns a valid strict-mode ADMIN_* environment.
func completeAdminEnv() map[string]string {
	return map[string]string{
		"ENVIRONMENT":       "production",
		"ADMIN_EMAIL":       "admin@example.com",
		"ADMIN_PASSWORD":    "CorrectHorse9",
		"ADMIN_FIRST_NAME":  "Admin",
		"ADMIN_LAST_NAME":   "User",
		"ADMIN_TELEPHONE":   "+33600000000",
		"ADMIN_POSTAL_CODE": "75001",
		"ADMIN_CITY":        "Paris",
		"ADMIN_ADDRESS1":    "1 Rue Example",
	}
}

func TestLoadConfigOutsideDevelopmentRefusesMissingValues(t *testing.T) {
	env := completeAdminEnv()
	delete(env, "ADMIN_EMAIL")
	delete(env, "ADMIN_PASSWORD")
	delete(env, "ADMIN_LAST_NAME")

	_, err := bootstrap.LoadConfig(envMap(env))
	require.Error(t, err, "missing values must be refused outside development")

	msg := err.Error()
	assert.Contains(t, msg, "ADMIN_EMAIL", "the missing keys should be named")
	assert.Contains(t, msg, "ADMIN_PASSWORD")
	assert.Contains(t, msg, "ADMIN_LAST_NAME")
	assert.NotContains(t, msg, "ADMIN_FIRST_NAME", "set values should not be reported missing")
}

func TestLoadConfigOutsideDevelopmentAcceptsCompleteValues(t *testing.T) {
	env := completeAdminEnv()
	env["ADMIN_ADDRESS2"] = "Appt 4"

	cfg, err := bootstrap.LoadConfig(envMap(env))
	require.NoError(t, err)

	assert.Equal(t, "admin@example.com", cfg.Email)
	assert.Equal(t, "CorrectHorse9", cfg.Password)
	assert.Equal(t, "Admin", cfg.FirstName)
	assert.Equal(t, "User", cfg.LastName)
	assert.Equal(t, "+33600000000", cfg.Telephone)
	assert.Equal(t, "75001", cfg.PostalCode)
	assert.Equal(t, "Paris", cfg.City)
	assert.Equal(t, "1 Rue Example", cfg.Address1)
	assert.Equal(t, "Appt 4", cfg.Address2)
}

func TestLoadConfigEmptyAddress2IsOptionalOutsideDevelopment(t *testing.T) {
	env := completeAdminEnv()
	env["ADMIN_ADDRESS2"] = ""

	_, err := bootstrap.LoadConfig(envMap(env))
	assert.NoError(t, err, "ADMIN_ADDRESS2 is the only optional value")
}

func TestLoadConfigDefaultsToDevelopment(t *testing.T) {
	// An unset ENVIRONMENT means development, like the rest of the app.
	_, err := bootstrap.LoadConfig(envMap(map[string]string{}))
	assert.NoError(t, err)
}

func TestLoadConfigDevelopmentAcceptsMissingValuesWithDefaults(t *testing.T) {
	cfg, err := bootstrap.LoadConfig(envMap(map[string]string{"ENVIRONMENT": "development"}))
	require.NoError(t, err, "missing values fall back to development defaults")

	assert.NotEmpty(t, cfg.Email)
	assert.NotEmpty(t, cfg.Password)
	assert.NotEmpty(t, cfg.FirstName)
	assert.NotEmpty(t, cfg.LastName)
	assert.True(t, strings.Contains(cfg.Email, "@"), "the default email should stay a usable address")
}
