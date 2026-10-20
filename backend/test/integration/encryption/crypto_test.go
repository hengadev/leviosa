package encryption_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/google/uuid"
	"github.com/hengadev/encx"
	awskeys "github.com/hengadev/encx/providers/keys/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption"
)

// newPepper returns a fresh base64 pepper, as an operator would put in Infisical.
func newPepper(t *testing.T) string {
	t.Helper()
	pepper := make([]byte, encx.PepperLength)
	_, err := rand.Read(pepper)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(pepper)
}

// newKEK provisions a KMS key behind a test-unique alias and returns the alias.
func newKEK(t *testing.T) string {
	t.Helper()
	alias := "alias/leviosa-test-" + uuid.NewString()
	_, err := kmsContainer.CreateKey(context.Background(), alias)
	require.NoError(t, err)
	return alias
}

func configFor(t *testing.T, alias, pepper string) encryption.Config {
	cfg := kmsContainer.EncryptionConfig(alias, pepper)
	cfg.KeyMetadataDir = t.TempDir()
	return cfg
}

func keyCount(t *testing.T) int {
	t.Helper()
	out, err := kmsContainer.Client().ListKeys(context.Background(), &kms.ListKeysInput{})
	require.NoError(t, err)
	return len(out.Keys)
}

func TestNewCrypto_RoundTripsThroughKMS(t *testing.T) {
	ctx := context.Background()
	alias := newKEK(t)

	crypto, err := encryption.NewCrypto(ctx, configFor(t, alias, newPepper(t)))
	require.NoError(t, err)

	dek, err := crypto.GenerateDEK()
	require.NoError(t, err)
	encryptedDEK, err := crypto.EncryptDEK(ctx, dek)
	require.NoError(t, err)

	// The encrypted DEK is a real KMS ciphertext under the aliased key, not a local wrap.
	blob, err := base64.StdEncoding.DecodeString(string(encryptedDEK))
	require.NoError(t, err)
	direct, err := kmsContainer.Client().Decrypt(ctx, &kms.DecryptInput{CiphertextBlob: blob})
	require.NoError(t, err)
	assert.Equal(t, dek, direct.Plaintext)
	described, err := kmsContainer.Client().DescribeKey(ctx, &kms.DescribeKeyInput{KeyId: aws.String(alias)})
	require.NoError(t, err)
	assert.Equal(t, aws.ToString(described.KeyMetadata.Arn), aws.ToString(direct.KeyId))

	version, err := crypto.GetCurrentKEKVersion(ctx, alias)
	require.NoError(t, err)
	decryptedDEK, err := crypto.DecryptDEKWithVersion(ctx, encryptedDEK, version)
	require.NoError(t, err)
	assert.Equal(t, dek, decryptedDEK)

	ciphertext, err := crypto.EncryptData(ctx, []byte("jane.doe@example.com"), dek)
	require.NoError(t, err)
	plaintext, err := crypto.DecryptData(ctx, ciphertext, decryptedDEK)
	require.NoError(t, err)
	assert.Equal(t, "jane.doe@example.com", string(plaintext))
}

// Every process (serve, seed, one-off modes, a restarted container) starts with fresh local
// encx metadata; what one encrypted, another must decrypt from KMS and the pepper alone.
func TestNewCrypto_AnotherProcessDecryptsAndVerifies(t *testing.T) {
	ctx := context.Background()
	alias := newKEK(t)
	pepper := newPepper(t)

	writer, err := encryption.NewCrypto(ctx, configFor(t, alias, pepper))
	require.NoError(t, err)

	dek, err := writer.GenerateDEK()
	require.NoError(t, err)
	encryptedDEK, err := writer.EncryptDEK(ctx, dek)
	require.NoError(t, err)
	version, err := writer.GetCurrentKEKVersion(ctx, alias)
	require.NoError(t, err)
	ciphertext, err := writer.EncryptData(ctx, []byte("+33612345678"), dek)
	require.NoError(t, err)
	password, err := encx.SerializeValue("Correct-Horse-1") // as the domain's encx code does
	require.NoError(t, err)
	passwordHash, err := writer.HashSecure(ctx, password)
	require.NoError(t, err)
	emailHash := writer.HashBasic(ctx, []byte("jane.doe@example.com"))

	reader, err := encryption.NewCrypto(ctx, configFor(t, alias, pepper))
	require.NoError(t, err)

	readerDEK, err := reader.DecryptDEKWithVersion(ctx, encryptedDEK, version)
	require.NoError(t, err)
	plaintext, err := reader.DecryptData(ctx, ciphertext, readerDEK)
	require.NoError(t, err)
	assert.Equal(t, "+33612345678", string(plaintext))

	ok, err := reader.CompareSecureHashAndValue(ctx, "Correct-Horse-1", passwordHash)
	require.NoError(t, err)
	assert.True(t, ok, "password hash made by one process must verify in another with the same pepper")
	assert.Equal(t, emailHash, reader.HashBasic(ctx, []byte("jane.doe@example.com")))

	// The pepper really is part of the password hash: another pepper doesn't verify it.
	otherPepper, err := encryption.NewCrypto(ctx, configFor(t, alias, newPepper(t)))
	require.NoError(t, err)
	ok, err = otherPepper.CompareSecureHashAndValue(ctx, "Correct-Horse-1", passwordHash)
	require.NoError(t, err)
	assert.False(t, ok, "a different pepper must not verify the password hash")
}

func TestNewCrypto_MissingPepperStopsStartup(t *testing.T) {
	alias := newKEK(t)

	_, err := encryption.NewCrypto(context.Background(), configFor(t, alias, ""))
	require.Error(t, err)
	assert.ErrorIs(t, err, encryption.ErrPepperMissing)
}

func TestNewCrypto_WrongLengthPepperStopsStartup(t *testing.T) {
	alias := newKEK(t)

	for _, n := range []int{16, 31, 33} {
		short := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, n))
		_, err := encryption.NewCrypto(context.Background(), configFor(t, alias, short))
		require.Error(t, err, "%d-byte pepper", n)
		assert.ErrorIs(t, err, encryption.ErrPepperInvalid, "%d-byte pepper", n)
	}
}

// pepperAbsent forces encx down its "no pepper yet: generate and store one" path while
// storing through the real env-backed store.
type pepperAbsent struct{ *encryption.PepperStore }

func (pepperAbsent) PepperExists(context.Context, string) (bool, error) { return false, nil }

func TestStorePepper_RefusedEvenWhenEncxTriesToGenerateOne(t *testing.T) {
	ctx := context.Background()
	alias := newKEK(t)
	cfg := configFor(t, alias, newPepper(t))

	store, err := encryption.NewPepperStore(cfg.Pepper)
	require.NoError(t, err)
	require.ErrorIs(t, store.StorePepper(ctx, "leviosa", make([]byte, encx.PepperLength)), encryption.ErrPepperReadOnly)

	keys, err := awskeys.NewKMSService(ctx, awskeys.Config{AWSConfig: &aws.Config{
		Region:       cfg.AWSRegion,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
		BaseEndpoint: aws.String(cfg.AWSEndpointURL),
	}})
	require.NoError(t, err)

	_, err = encx.NewCrypto(ctx, keys, pepperAbsent{store}, encx.Config{
		KEKAlias:    alias,
		PepperAlias: "leviosa",
		DBPath:      t.TempDir(),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, encryption.ErrPepperReadOnly)
}

// The KEK is provisioned outside the app (Terraform, or the LocalStack init container).
// encx creates a key when its alias doesn't resolve; the app must refuse instead.
func TestNewCrypto_UnknownAliasStopsStartupWithoutCreatingAKey(t *testing.T) {
	before := keyCount(t)

	_, err := encryption.NewCrypto(context.Background(), configFor(t, "alias/leviosa-does-not-exist", newPepper(t)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alias/leviosa-does-not-exist")

	assert.Equal(t, before, keyCount(t), "no KMS key may be created at startup")
}

func TestNewCrypto_RejectsIncompleteConfigBeforeCallingKMS(t *testing.T) {
	cfg := configFor(t, "alias/leviosa-anything", newPepper(t))
	cfg.AWSRegion = ""

	_, err := encryption.NewCrypto(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), encryption.EnvAWSRegion)
	assert.False(t, errors.Is(err, encryption.ErrPepperMissing))
}
