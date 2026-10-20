package bootstrap_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrap "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/bootstrap"
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
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

// TestBootstrapAdminSkipsExistingWithDifferentPassword covers the second
// deploy: re-running with a different ADMIN_PASSWORD exits 0 (no error),
// and neither the stored password hash nor the role changes.
func TestBootstrapAdminSkipsExistingWithDifferentPassword(t *testing.T) {
	ctx, logs := newTestLogger(context.Background())

	email := "bootstrap-skip@example.com"
	firstPassword := "FirstAdmin9Pass"
	secondPassword := "SecondAdmin8Pass"

	cfg, err := bootstrap.LoadConfig(envMap(strictAdminEnv(email, firstPassword)))
	require.NoError(t, err)
	_, err = svc.Run(ctx, cfg)
	require.NoError(t, err)

	emailBytes, err := encx.SerializeValue(email)
	require.NoError(t, err)
	storedFirst, err := repo.GetUserByEmailHash(ctx, crypto.HashBasic(ctx, emailBytes))
	require.NoError(t, err)
	hashAfterFirst := storedFirst.PasswordHashSecure

	// Second run with a different ADMIN_PASSWORD: must succeed and change
	// nothing.
	cfgSecond, err := bootstrap.LoadConfig(envMap(strictAdminEnv(email, secondPassword)))
	require.NoError(t, err)
	created, err := svc.Run(ctx, cfgSecond)
	require.NoError(t, err, "re-running must exit 0")
	assert.False(t, created, "an existing account is never recreated")

	storedSecond, err := repo.GetUserByEmailHash(ctx, crypto.HashBasic(ctx, emailBytes))
	require.NoError(t, err)
	assert.Equal(t, hashAfterFirst, storedSecond.PasswordHashSecure,
		"the stored password hash must be unchanged")

	// The first password still verifies; the second one must not.
	require.NoError(t, users.VerifyUserPassword(ctx, storedSecond.ID, firstPassword),
		"the original password must still log in")
	assert.Error(t, users.VerifyUserPassword(ctx, storedSecond.ID, secondPassword),
		"the new ADMIN_PASSWORD must not have replaced the stored one")

	decrypted, err := domain.DecryptUserEncx(ctx, crypto, storedSecond)
	require.NoError(t, err)
	assert.Equal(t, identity.AdministratorStr, decrypted.Role, "the role must be unchanged")

	assert.NotContains(t, logs.String(), firstPassword)
	assert.NotContains(t, logs.String(), secondPassword)
	assert.NotContains(t, logs.String(), hashAfterFirst, "logs must not contain the password hash either")
}

// TestBootstrapAdminNeverModifiesExistingNonAdminAccount covers the guard
// role of the mode: a pre-existing standard user with the same email is
// neither re-passworded nor promoted.
func TestBootstrapAdminNeverModifiesExistingNonAdminAccount(t *testing.T) {
	ctx, _ := newTestLogger(context.Background())

	email := "bootstrap-existing@example.com"
	existingPassword := "Existing9Pass"

	existing := &domain.User{
		ID:        uuid.New(),
		State:     domain.Active,
		Email:     email,
		Password:  existingPassword,
		FirstName: "Some",
		LastName:  "Client",
		Role:      identity.Standard.String(),
		CreatedAt: time.Now(),
	}
	existingEncx, err := domain.ProcessUserEncx(ctx, crypto, existing)
	require.NoError(t, err)
	require.NoError(t, repo.CreateUser(ctx, existingEncx))

	cfg, err := bootstrap.LoadConfig(envMap(strictAdminEnv(email, "Replacement9Pass")))
	require.NoError(t, err)
	created, err := svc.Run(ctx, cfg)
	require.NoError(t, err)
	assert.False(t, created)

	emailBytes, err := encx.SerializeValue(email)
	require.NoError(t, err)
	stored, err := repo.GetUserByEmailHash(ctx, crypto.HashBasic(ctx, emailBytes))
	require.NoError(t, err)

	decrypted, err := domain.DecryptUserEncx(ctx, crypto, stored)
	require.NoError(t, err)
	assert.Equal(t, identity.StandardStr, decrypted.Role, "an existing account is never promoted")
	assert.Equal(t, existing.ID, decrypted.ID, "an existing account is never replaced")

	require.NoError(t, users.VerifyUserPassword(ctx, stored.ID, existingPassword),
		"the existing password must still verify")
	assert.Error(t, users.VerifyUserPassword(ctx, stored.ID, "Replacement9Pass"),
		"the bootstrap password must not have been applied")
}

// TestBootstrapAdminDevelopmentAcceptsMissingAndWeakValues covers the
// development escape hatch end to end: with no ADMIN_* values at all (and a
// weak password explicitly set), the run still creates an admin who can log
// in with the resulting password.
func TestBootstrapAdminDevelopmentAcceptsMissingAndWeakValues(t *testing.T) {
	ctx, logs := newTestLogger(context.Background())

	env := map[string]string{
		"ENVIRONMENT":    "development",
		"ADMIN_PASSWORD": "weak",
	}
	cfg, err := bootstrap.LoadConfig(envMap(env))
	require.NoError(t, err, "development accepts missing values and weak passwords")
	assert.Equal(t, "weak", cfg.Password, "an explicitly set weak password is used as-is")
	assert.NotEmpty(t, cfg.Email, "the missing values fall back to development defaults")

	created, err := svc.Run(ctx, cfg)
	require.NoError(t, err)
	assert.True(t, created)

	user, err := users.GetUserByEmail(ctx, cfg.Email)
	require.NoError(t, err)
	require.NoError(t, users.VerifyUserPassword(ctx, user.ID, "weak"),
		"the development admin must be able to log in")

	assert.NotContains(t, logs.String(), "weak")
}

// TestBootstrapAdminOutsideDevelopmentRefusalRunsBeforeAnyWrite covers the
// refusal path against the real configuration loader the mode calls: a
// missing value or a weak password must make the run fail before any user
// is created, which is what turns into the mode's non-zero exit.
func TestBootstrapAdminOutsideDevelopmentRefusalRunsBeforeAnyWrite(t *testing.T) {
	email := "bootstrap-refused@example.com"

	t.Run("missing value", func(t *testing.T) {
		env := strictAdminEnv(email, "ValidAdmin9Pass")
		delete(env, "ADMIN_FIRST_NAME")

		_, err := bootstrap.LoadConfig(envMap(env))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ADMIN_FIRST_NAME")
	})

	t.Run("weak password", func(t *testing.T) {
		env := strictAdminEnv(email, "short")

		_, err := bootstrap.LoadConfig(envMap(env))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ADMIN_PASSWORD")
		assert.NotContains(t, err.Error(), "short", "the refusal must not echo the password")
	})

	// Nothing was written by the refused runs.
	ctx := context.Background()
	emailBytes, err := encx.SerializeValue(email)
	require.NoError(t, err)
	exists, err := repo.ExistsByEmailHash(ctx, crypto.HashBasic(ctx, emailBytes))
	require.NoError(t, err)
	assert.False(t, exists, "a refused run must not create the account")
}
