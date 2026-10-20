package encryption_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hengadev/encx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
)

// Integration suites get encx from tu.SetupCrypto; it must be encx as the backend builds
// it: DEKs wrapped by the KMS key behind the alias, and a 32-byte pepper in the hashes.
func TestSetupCrypto_ProvisionsEncxOnKMSWithAGeneratedPepper(t *testing.T) {
	ctx := context.Background()

	setup, err := tu.SetupCrypto(ctx, t)
	require.NoError(t, err)
	t.Cleanup(func() { tu.TeardownCrypto(ctx, t, setup) })

	require.NotNil(t, setup.Crypto)
	require.NoError(t, setup.Config.Validate())
	assert.Equal(t, tu.TestKMSKeyAlias, setup.Config.KMSKeyAlias)
	pepper, err := base64.StdEncoding.DecodeString(setup.Config.Pepper)
	require.NoError(t, err)
	assert.Len(t, pepper, encx.PepperLength)
	assert.NotEqual(t, make([]byte, encx.PepperLength), pepper, "the pepper must be random, not zero")

	// The DEK is a real KMS ciphertext under the aliased key.
	dek, err := setup.Crypto.GenerateDEK()
	require.NoError(t, err)
	encryptedDEK, err := setup.Crypto.EncryptDEK(ctx, dek)
	require.NoError(t, err)
	blob, err := base64.StdEncoding.DecodeString(string(encryptedDEK))
	require.NoError(t, err)
	direct, err := setup.KMS.Client().Decrypt(ctx, &kms.DecryptInput{CiphertextBlob: blob})
	require.NoError(t, err)
	assert.Equal(t, dek, direct.Plaintext)
	described, err := setup.KMS.Client().DescribeKey(ctx, &kms.DescribeKeyInput{KeyId: aws.String(tu.TestKMSKeyAlias)})
	require.NoError(t, err)
	assert.Equal(t, aws.ToString(described.KeyMetadata.Arn), aws.ToString(direct.KeyId))

	// Another process built from the same configuration reads what this one wrote.
	password, err := encx.SerializeValue("Correct-Horse-1")
	require.NoError(t, err)
	passwordHash, err := setup.Crypto.HashSecure(ctx, password)
	require.NoError(t, err)

	readerCfg := setup.Config
	readerCfg.KeyMetadataDir = t.TempDir()
	reader, err := encryption.NewCrypto(ctx, readerCfg)
	require.NoError(t, err)
	ok, err := reader.CompareSecureHashAndValue(ctx, "Correct-Horse-1", passwordHash)
	require.NoError(t, err)
	assert.True(t, ok, "a crypto service on the same KEK and pepper must verify the hash")
	version, err := setup.Crypto.GetCurrentKEKVersion(ctx, tu.TestKMSKeyAlias)
	require.NoError(t, err)
	readerDEK, err := reader.DecryptDEKWithVersion(ctx, encryptedDEK, version)
	require.NoError(t, err)
	assert.Equal(t, dek, readerDEK)
}
