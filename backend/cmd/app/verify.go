package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption"
	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption/canary"
)

// Exit statuses of verify's failures, so a restore test can tell them apart.
// Any other error (configuration, connectivity) exits 1.
const (
	exitNothingToVerify  = 3
	exitCanaryMissing    = 4
	exitDecryptionFailed = 5
	exitPepperMismatch   = 6
)

// usersTable holds the records verify decrypts.
const usersTable = "auth.users"

// verify proves the database is readable with the configured KMS key
// (ENCX_KMS_KEY_ALIAS) and pepper (ENCX_PEPPER), and exits: `leviosa verify`.
// The restore test runs it against a restored dump. Inside a read-only
// transaction it decrypts one auth.users record (the KEK on real data), then
// decrypts the pepper canary and checks its argon2id hash (the KEK again, and
// the pepper). It prints a one-line summary — table, record id, never a
// plaintext — and fails with a distinct message and exit status when there is
// nothing to verify, the canary is missing, decryption fails or the canary's
// hash does not match. It reads only DATABASE_URL and the encx settings, and
// never writes.
func verify(ctx context.Context, args []string) error {
	return runVerify(ctx, args, os.Stdout)
}

func runVerify(ctx context.Context, args []string, out io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("verify takes no arguments, got %q", args)
	}

	loadDevelopmentEnv()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("verify: DATABASE_URL is required")
	}

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("verify: connect to postgres: %w", err)
	}
	defer conn.Close(ctx)

	crypto, err := encryption.NewCrypto(ctx, encryption.ConfigFromEnv(os.Getenv))
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("verify: begin read-only transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	summary, err := verifyDatabase(ctx, tx, crypto)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, summary)
	return nil
}

// verifyDatabase runs the checks and returns the success summary.
func verifyDatabase(ctx context.Context, tx pgx.Tx, crypto encx.CryptoService) (string, error) {
	alias := crypto.GetAlias()

	user, fields, err := loadUser(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("verify: %w", err)
	}
	if user != nil {
		// The decrypted user is dropped right away: verify never keeps or prints plaintext.
		if _, err := domain.DecryptUserEncx(ctx, crypto, user); err != nil {
			return "", failVerify(exitDecryptionFailed,
				"verify: decryption failed with KMS key %s (%s): %s record %s does not decrypt: %w",
				alias, encryption.EnvKMSKeyAlias, usersTable, user.ID, err)
		}
	}

	err = canary.Verify(ctx, tx, crypto)
	switch {
	case errors.Is(err, canary.ErrMissing) && user == nil:
		return "", failVerify(exitNothingToVerify,
			"verify: nothing to verify: the database has no %s record and no pepper canary (%s) — is DATABASE_URL the restored database, and did the restore succeed?",
			usersTable, canary.Table)
	case errors.Is(err, canary.ErrMissing):
		return "", failVerify(exitCanaryMissing,
			"verify: %w — start serve once against this database (it writes the canary at startup), then run verify again; %s record %s decrypted with KMS key %s, but the pepper can only be checked against the canary",
			err, usersTable, user.ID, alias)
	case errors.Is(err, canary.ErrDecryption):
		return "", failVerify(exitDecryptionFailed,
			"verify: decryption failed with KMS key %s (%s): %w",
			alias, encryption.EnvKMSKeyAlias, err)
	case errors.Is(err, canary.ErrPepperMismatch):
		return "", failVerify(exitPepperMismatch,
			"verify: %w: the canary (%s) decrypted with KMS key %s, but its hash does not match under %s — the pepper differs from the one this database was written with",
			err, canary.Table, alias, encryption.EnvPepper)
	case err != nil:
		return "", fmt.Errorf("verify: %w", err)
	}

	decrypted := fmt.Sprintf("no %s record to decrypt", usersTable)
	if user != nil {
		decrypted = fmt.Sprintf("%s record %s decrypted (%d encrypted fields)", usersTable, user.ID, fields)
	}
	return fmt.Sprintf("verify: OK: %s; pepper canary matched (%s); KMS key %s and %s both read this database",
		decrypted, canary.Table, alias, encryption.EnvPepper), nil
}

// loadUser reads the oldest auth.users record's encrypted columns and counts
// the non-empty ones. It returns a nil user when there is none, including on a
// database that was never migrated.
func loadUser(ctx context.Context, tx pgx.Tx) (*domain.UserEncx, int, error) {
	var tableExists bool
	if err := tx.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", usersTable).Scan(&tableExists); err != nil {
		return nil, 0, fmt.Errorf("check for the %s table: %w", usersTable, err)
	}
	if !tableExists {
		return nil, 0, nil
	}

	user := &domain.UserEncx{}
	encrypted := []*[]byte{
		&user.EmailEncrypted, &user.PictureEncrypted, &user.CreatedAtEncrypted,
		&user.LoggedInAtEncrypted, &user.RoleEncrypted, &user.BirthDateEncrypted,
		&user.LastNameEncrypted, &user.FirstNameEncrypted, &user.GenderEncrypted,
		&user.TelephoneEncrypted, &user.PostalCodeEncrypted, &user.CityEncrypted,
		&user.Address1Encrypted, &user.Address2Encrypted, &user.GoogleIDEncrypted,
		&user.AppleIDEncrypted, &user.StripeCustomerIDEncrypted,
	}
	dest := []any{&user.ID, &user.DEKEncrypted, &user.KeyVersion}
	for _, field := range encrypted {
		dest = append(dest, field)
	}

	err := tx.QueryRow(ctx, `
		SELECT id, dek_encrypted, key_version,
			email_encrypted, picture_encrypted, created_at_encrypted,
			logged_in_at_encrypted, role_encrypted, birth_date_encrypted,
			last_name_encrypted, first_name_encrypted, gender_encrypted,
			telephone_encrypted, postal_code_encrypted, city_encrypted,
			address1_encrypted, address2_encrypted, google_id_encrypted,
			apple_id_encrypted, stripe_customer_id_encrypted
		FROM `+usersTable+`
		ORDER BY created_at, id
		LIMIT 1`,
	).Scan(dest...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("read a %s record: %w", usersTable, err)
	}

	fields := 0
	for _, field := range encrypted {
		if len(*field) > 0 {
			fields++
		}
	}
	return user, fields, nil
}

// verifyError is a verify failure that carries its own exit status.
type verifyError struct {
	code int
	err  error
}

func (e *verifyError) Error() string { return e.err.Error() }
func (e *verifyError) Unwrap() error { return e.err }
func (e *verifyError) ExitCode() int { return e.code }

func failVerify(code int, format string, args ...any) error {
	return &verifyError{code: code, err: fmt.Errorf(format, args...)}
}
