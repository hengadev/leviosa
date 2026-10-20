package bootstrap_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrap "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/bootstrap"
)

func TestValidateAdminPasswordAcceptsStrongPasswords(t *testing.T) {
	assert.NoError(t, bootstrap.ValidateAdminPassword("CorrectHorse9"))
	assert.NoError(t, bootstrap.ValidateAdminPassword("aB1aaaaaaaaaa"))
}

func TestValidateAdminPasswordRefusesWeakPasswords(t *testing.T) {
	cases := map[string]string{
		"too short":           "Ab1aaaaaaa", // 10 characters
		"no lowercase letter": "ABCDEFGH1234",
		"no uppercase letter": "abcdefgh1234",
		"no digit":            "Abcdefghijkl",
		"empty":               "",
		"only whitespace":     "             ",
		"length without mix":  "aaaaaaaaaaaa",
	}
	for name, password := range cases {
		t.Run(name, func(t *testing.T) {
			err := bootstrap.ValidateAdminPassword(password)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "ADMIN_PASSWORD", "the error should name the value to fix")
			if password != "" {
				assert.NotContains(t, err.Error(), password, "the error must not echo the password")
			}
		})
	}
}

func TestLoadConfigOutsideDevelopmentRefusesWeakPassword(t *testing.T) {
	for _, password := range []string{"short", "nodigitsatallAB", "nolowercase12345", "12charsnoupper1"} {
		env := completeAdminEnv()
		env["ADMIN_PASSWORD"] = password

		_, err := bootstrap.LoadConfig(envMap(env))
		assert.Error(t, err, "a weak ADMIN_PASSWORD must be refused outside development: %q", password)
	}
}

func TestLoadConfigDevelopmentAcceptsWeakPassword(t *testing.T) {
	env := map[string]string{
		"ENVIRONMENT":    "development",
		"ADMIN_EMAIL":    "weak@example.com",
		"ADMIN_PASSWORD": "weak",
	}

	cfg, err := bootstrap.LoadConfig(envMap(env))
	require.NoError(t, err, "development accepts a weak password")
	assert.Equal(t, "weak", cfg.Password)
	assert.Equal(t, "weak@example.com", cfg.Email, "set values are used as-is in development")
	assert.True(t, strings.Contains(cfg.Email, "@"), "sanity: email field populated")
}
