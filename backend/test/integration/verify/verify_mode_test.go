package verify_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption/canary"
)

// Exit statuses of the verify mode, one per failure (1 is any other error).
const (
	exitNothingToVerify  = 3
	exitCanaryMissing    = 4
	exitDecryptionFailed = 5
	exitPepperMismatch   = 6
)

// admin is the user every populated database holds, created by the
// bootstrap-admin mode through the authuser domain path (encrypted fields,
// argon2id password, searchable email hash). Its values are distinctive so a
// leak into verify's output is easy to spot.
var admin = map[string]string{
	"ADMIN_EMAIL":       "verify.canary.admin@example.test",
	"ADMIN_PASSWORD":    "Verify-Secret-Passw0rd",
	"ADMIN_FIRST_NAME":  "Zephyrine",
	"ADMIN_LAST_NAME":   "Quillfeather",
	"ADMIN_TELEPHONE":   "+33611223344",
	"ADMIN_POSTAL_CODE": "69007",
	"ADMIN_CITY":        "Montbrison-sur-Verify",
	"ADMIN_ADDRESS1":    "12 Rue du Canari",
	"ADMIN_ADDRESS2":    "Batiment Pepper",
}

// createUser runs bootstrap-admin, which creates the admin user through the app.
func createUser(t *testing.T, dsn string) {
	t.Helper()
	mustRun(t, "bootstrap-admin", appEnv(dsn, kekAlias, pepper, admin))
}

// populate returns a database migrated, holding one user, with the pepper
// canary written by serve's startup: what a deployed environment's dump holds.
func populate(t *testing.T, name string) string {
	t.Helper()

	dsn := createTestDB(t, name)
	mustRun(t, "migrate", envFor(dsn))
	createUser(t, dsn)
	serveOnce(t, envFor(dsn))
	return dsn
}

// assertNoPlaintext fails if output contains any of the admin's plaintext
// values, the canary constant or the pepper.
func assertNoPlaintext(t *testing.T, output string) {
	t.Helper()

	secrets := []string{canary.Value, pepper}
	for _, v := range admin {
		secrets = append(secrets, v)
	}
	for _, s := range secrets {
		if strings.Contains(output, s) {
			t.Errorf("verify printed the plaintext %q; output:\n%s", s, output)
		}
	}
}

func assertExit(t *testing.T, run appRun, code int, wants ...string) {
	t.Helper()
	t.Logf("verify exited %d; output:\n%s", run.exitCode, run.output)

	if run.exitCode != code {
		t.Fatalf("expected exit code %d, got %d; output:\n%s", code, run.exitCode, run.output)
	}
	for _, want := range wants {
		if !strings.Contains(run.output, want) {
			t.Errorf("expected the output to mention %q; output:\n%s", want, run.output)
		}
	}
}

func TestVerifyPopulatedDatabase(t *testing.T) {
	dsn := populate(t, "verify_populated")

	var userID string
	if err := openDB(t, dsn).QueryRow("SELECT id::text FROM auth.users").Scan(&userID); err != nil {
		t.Fatalf("read the user's id: %v", err)
	}

	t.Run("configured KMS key and pepper", func(t *testing.T) {
		run := runApp(t, "verify", envFor(dsn))
		assertExit(t, run, 0, "auth.users", userID, "decrypted", "pepper canary matched", kekAlias)
		assertNoPlaintext(t, run.output)
	})

	t.Run("different KMS key", func(t *testing.T) {
		run := runApp(t, "verify", appEnv(dsn, otherKEKAlias, pepper, nil))
		assertExit(t, run, exitDecryptionFailed, "decryption failed", "auth.users", userID, otherKEKAlias)
		assertNoPlaintext(t, run.output)
	})

	t.Run("different pepper", func(t *testing.T) {
		run := runApp(t, "verify", appEnv(dsn, kekAlias, newPepper(), nil))
		assertExit(t, run, exitPepperMismatch, "pepper canary mismatch", "ENCX_PEPPER")
		assertNoPlaintext(t, run.output)
	})
}

func TestVerifyEmptyDatabase(t *testing.T) {
	t.Run("never migrated", func(t *testing.T) {
		dsn := createTestDB(t, "verify_empty_unmigrated")
		run := runApp(t, "verify", envFor(dsn))
		assertExit(t, run, exitNothingToVerify, "nothing to verify")
	})

	t.Run("migrated, no data", func(t *testing.T) {
		dsn := createTestDB(t, "verify_empty_migrated")
		mustRun(t, "migrate", envFor(dsn))
		run := runApp(t, "verify", envFor(dsn))
		assertExit(t, run, exitNothingToVerify, "nothing to verify")
	})
}

func TestVerifyUserWithoutCanary(t *testing.T) {
	dsn := createTestDB(t, "verify_no_canary")
	mustRun(t, "migrate", envFor(dsn))
	createUser(t, dsn) // serve never started: no canary

	run := runApp(t, "verify", envFor(dsn))
	assertExit(t, run, exitCanaryMissing, "pepper canary missing", "start serve once")
	assertNoPlaintext(t, run.output)

	// verify is read-only: it must not have written the canary itself.
	var rows int
	if err := openDB(t, dsn).QueryRow("SELECT count(*) FROM " + canary.Table).Scan(&rows); err != nil {
		t.Fatalf("count canary rows: %v", err)
	}
	if rows != 0 {
		t.Fatalf("verify wrote %d canary row(s); it must never write", rows)
	}
}

// A database serve has started against but no user has signed up to yet: the
// canary alone proves both the KMS key and the pepper.
func TestVerifyCanaryWithoutUsers(t *testing.T) {
	dsn := createTestDB(t, "verify_canary_only")
	mustRun(t, "migrate", envFor(dsn))
	serveOnce(t, envFor(dsn))

	run := runApp(t, "verify", envFor(dsn))
	assertExit(t, run, 0, "pepper canary matched", canary.Table)

	run = runApp(t, "verify", appEnv(dsn, otherKEKAlias, pepper, nil))
	assertExit(t, run, exitDecryptionFailed, "decryption failed", canary.Table)
}

// serve writes the canary once and never rewrites it — not even when started
// with another pepper, which would otherwise make the canary vouch for it.
func TestServeNeverOverwritesCanary(t *testing.T) {
	dsn := populate(t, "verify_canary_kept")
	db := openDB(t, dsn)

	readCanary := func() (value []byte, hash string) {
		t.Helper()
		if err := db.QueryRow(
			"SELECT value_encrypted, value_hash_secure FROM "+canary.Table,
		).Scan(&value, &hash); err != nil {
			t.Fatalf("read the canary: %v", err)
		}
		return value, hash
	}
	value, hash := readCanary()

	otherPepper := newPepper()
	serveOnce(t, appEnv(dsn, kekAlias, otherPepper, nil))

	gotValue, gotHash := readCanary()
	if !bytes.Equal(gotValue, value) || gotHash != hash {
		t.Fatal("serve rewrote the existing canary")
	}

	assertExit(t, runApp(t, "verify", envFor(dsn)), 0, "pepper canary matched")
	assertExit(t, runApp(t, "verify", appEnv(dsn, kekAlias, otherPepper, nil)), exitPepperMismatch, "pepper canary mismatch")
}
