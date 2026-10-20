package bootstrap_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrap "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/bootstrap"
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
)

// strictAdminEnv returns a complete strict-mode environment for the given
// email/password, like the one the deployed bootstrap container receives.
func strictAdminEnv(email, password string) map[string]string {
	return map[string]string{
		"ENVIRONMENT":       "production",
		"ADMIN_EMAIL":       email,
		"ADMIN_PASSWORD":    password,
		"ADMIN_FIRST_NAME":  "Admin",
		"ADMIN_LAST_NAME":   "User",
		"ADMIN_TELEPHONE":   "+33600000000",
		"ADMIN_POSTAL_CODE": "75001",
		"ADMIN_CITY":        "Paris",
		"ADMIN_ADDRESS1":    "1 Rue Example",
	}
}

func envMap(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}

// TestBootstrapAdminCreatesSuperadminWhoCanLogIn covers the first deploy on
// an empty database: the mode creates the superadmin, and the configured
// password verifies against the stored hash through the same code path
// SignIn uses.
func TestBootstrapAdminCreatesSuperadminWhoCanLogIn(t *testing.T) {
	ctx, logs := newTestLogger(context.Background())

	email := "bootstrap-create@example.com"
	password := "InitialAdmin9Pass"

	cfg, err := bootstrap.LoadConfig(envMap(strictAdminEnv(email, password)))
	require.NoError(t, err)

	created, err := svc.Run(ctx, cfg)
	require.NoError(t, err)
	assert.True(t, created, "on an empty database the superadmin is created")

	// "Can log in": GetUserByEmail + VerifyUserPassword is the SignIn path.
	user, err := users.GetUserByEmail(ctx, email)
	require.NoError(t, err, "the created admin must be findable by email")
	assert.Equal(t, identity.AdministratorStr, user.Role, "the account must hold the top admin role")
	assert.Equal(t, domain.Active, user.State, "the account must be active, not pending")
	assert.False(t, user.ProfileIncomplete)

	require.NoError(t, users.VerifyUserPassword(ctx, user.ID, password),
		"the configured password must verify against the stored hash")

	err = users.VerifyUserPassword(ctx, user.ID, "WrongPassword9x")
	assert.Error(t, err, "a different password must not verify")

	assert.NotContains(t, logs.String(), password, "logs must never contain the password")
}
