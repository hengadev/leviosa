// Package canary keeps the pepper canary: the single row of encryption.canary,
// holding a fixed public constant (Value) encrypted through encx — its DEK
// wrapped by the KMS KEK — and the constant's argon2id HashSecure, which mixes
// in the pepper.
//
// encx's searchable hashes (HashBasic, e.g. email_hash) are plain SHA-256
// without the pepper, and the only peppered values in the app's data —
// password hashes — need the plaintext password to check. The canary is the
// one record whose hash can be recomputed from a known value, so it is what
// proves that a database (typically a restored dump) is readable with a given
// pepper; decrypting it proves the KMS key as well.
//
// serve writes it at startup when it is absent (Ensure) and never updates it;
// the verify mode checks it (Verify). It must never be deleted or rewritten:
// a canary rewritten under another pepper would vouch for the wrong one.
package canary

import (
	"context"
	"errors"
	"fmt"

	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Value is the canary's plaintext. It is public and fixed: it proves keys, it
// is not a secret. Changing it would make every existing canary fail Verify.
const Value = "leviosa pepper canary v1"

// Table is the canary's table (migration 20261020120000_encryption_add_table_canary.sql).
const Table = "encryption.canary"

var (
	// ErrMissing means the database has no canary (or no canary table): serve
	// has not started against it since the canary migration.
	ErrMissing = errors.New("pepper canary missing")
	// ErrDecryption means the canary does not decrypt to Value with the
	// configured KMS key.
	ErrDecryption = errors.New("the pepper canary (" + Table + ") does not decrypt")
	// ErrPepperMismatch means the canary decrypts but its hash does not match
	// Value under the configured pepper.
	ErrPepperMismatch = errors.New("pepper canary mismatch")
)

// DB is what the canary needs from a connection; *pgxpool.Pool, *pgx.Conn and
// pgx.Tx all satisfy it.
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Ensure writes the canary with crypto's KEK and pepper when the database has
// none, and reports whether it did. An existing canary is never updated or
// overwritten, whatever crypto's keys are; concurrent starts are safe (the
// insert is ON CONFLICT DO NOTHING). The schema must already be migrated.
func Ensure(ctx context.Context, db DB, crypto encx.CryptoService) (bool, error) {
	var exists bool
	if err := db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM "+Table+")").Scan(&exists); err != nil {
		return false, fmt.Errorf("check for the pepper canary: %w", err)
	}
	if exists {
		return false, nil
	}

	plaintext, err := encx.SerializeValue(Value)
	if err != nil {
		return false, fmt.Errorf("serialize the pepper canary: %w", err)
	}
	dek, err := crypto.GenerateDEK()
	if err != nil {
		return false, fmt.Errorf("generate the pepper canary's DEK: %w", err)
	}
	valueEncrypted, err := crypto.EncryptData(ctx, plaintext, dek)
	if err != nil {
		return false, fmt.Errorf("encrypt the pepper canary: %w", err)
	}
	dekEncrypted, err := crypto.EncryptDEK(ctx, dek)
	if err != nil {
		return false, fmt.Errorf("encrypt the pepper canary's DEK: %w", err)
	}
	keyVersion, err := crypto.GetCurrentKEKVersion(ctx, crypto.GetAlias())
	if err != nil {
		return false, fmt.Errorf("read the KEK version for the pepper canary: %w", err)
	}
	hash, err := crypto.HashSecure(ctx, plaintext)
	if err != nil {
		return false, fmt.Errorf("hash the pepper canary: %w", err)
	}

	tag, err := db.Exec(ctx,
		"INSERT INTO "+Table+" (id, value_encrypted, value_hash_secure, dek_encrypted, key_version)"+
			" VALUES (1, $1, $2, $3, $4) ON CONFLICT (id) DO NOTHING",
		valueEncrypted, hash, dekEncrypted, keyVersion,
	)
	if err != nil {
		return false, fmt.Errorf("write the pepper canary: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

// Verify checks the canary against crypto: it must decrypt to Value with the
// KEK (else ErrDecryption) and its hash must match Value under the pepper (else
// ErrPepperMismatch). A database without the canary, or without its table,
// yields ErrMissing. Verify only reads, so it runs inside a read-only
// transaction.
func Verify(ctx context.Context, db DB, crypto encx.CryptoService) error {
	var tableExists bool
	if err := db.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", Table).Scan(&tableExists); err != nil {
		return fmt.Errorf("check for the %s table: %w", Table, err)
	}
	if !tableExists {
		return fmt.Errorf("%w: no %s table (the database predates the canary migration)", ErrMissing, Table)
	}

	var (
		valueEncrypted, dekEncrypted []byte
		hash                         string
		keyVersion                   int
	)
	err := db.QueryRow(ctx,
		"SELECT value_encrypted, value_hash_secure, dek_encrypted, key_version FROM "+Table+" WHERE id = 1",
	).Scan(&valueEncrypted, &hash, &dekEncrypted, &keyVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %s is empty", ErrMissing, Table)
	}
	if err != nil {
		return fmt.Errorf("read the pepper canary: %w", err)
	}

	dek, err := crypto.DecryptDEKWithVersion(ctx, dekEncrypted, keyVersion)
	if err != nil {
		return fmt.Errorf("%w: decrypt its DEK: %w", ErrDecryption, err)
	}
	plaintext, err := crypto.DecryptData(ctx, valueEncrypted, dek)
	if err != nil {
		return fmt.Errorf("%w: decrypt its value: %w", ErrDecryption, err)
	}
	var value string
	if err := encx.DeserializeValue(plaintext, &value); err != nil || value != Value {
		return fmt.Errorf("%w: it does not decrypt to the canary constant", ErrDecryption)
	}

	matched, err := crypto.CompareSecureHashAndValue(ctx, Value, hash)
	if err != nil {
		return fmt.Errorf("check the pepper canary's hash: %w", err)
	}
	if !matched {
		return ErrPepperMismatch
	}
	return nil
}
