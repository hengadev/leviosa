package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption"
)

// LocalStack KMS settings shared by the container and its clients. LocalStack accepts
// any credentials; keys are scoped to the region they were created in.
const (
	KMSRegion          = "us-east-1"
	KMSAccessKeyID     = "test"
	KMSSecretAccessKey = "test"
)

// KMSContainer is a LocalStack container serving AWS KMS — the local stand-in for the
// per-environment KMS key that holds the encx KEK (ADR-0015).
type KMSContainer struct {
	testcontainers.Container
	// Endpoint is the host-reachable LocalStack URL, e.g. http://localhost:32768.
	Endpoint string
}

// SetupKMS starts LocalStack with only the KMS service enabled. Create a KEK with
// CreateKey, then build encx from EncryptionConfig. Pass a nil t from TestMain.
func SetupKMS(ctx context.Context, t *testing.T) (*KMSContainer, error) {
	setupCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	container, err := testcontainers.GenericContainer(setupCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "localstack/localstack:4.6",
			ExposedPorts: []string{"4566/tcp"},
			Env: map[string]string{
				"SERVICES":           "kms",
				"AWS_DEFAULT_REGION": KMSRegion,
				"PERSISTENCE":        "0",
			},
			WaitingFor: wait.ForLog("Ready.").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start LocalStack KMS container: %w", err)
	}

	host, err := container.Host(setupCtx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get LocalStack host: %w", err)
	}
	port, err := container.MappedPort(setupCtx, "4566")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get LocalStack port: %w", err)
	}

	c := &KMSContainer{Container: container, Endpoint: fmt.Sprintf("http://%s:%s", host, port.Port())}
	if t != nil {
		t.Logf("LocalStack KMS endpoint: %s", c.Endpoint)
	}
	return c, nil
}

// TeardownKMS terminates the container; nil is a no-op.
func TeardownKMS(ctx context.Context, t *testing.T, c *KMSContainer) {
	if c == nil {
		return
	}
	if err := c.Terminate(ctx); err != nil && t != nil {
		t.Logf("terminate LocalStack KMS container: %v", err)
	}
}

// Client returns a KMS client for the container, for arranging test state directly.
func (c *KMSContainer) Client() *kms.Client {
	return kms.NewFromConfig(aws.Config{
		Region:       KMSRegion,
		Credentials:  credentials.NewStaticCredentialsProvider(KMSAccessKeyID, KMSSecretAccessKey, ""),
		BaseEndpoint: aws.String(c.Endpoint),
	})
}

// CreateKey creates a symmetric encryption key and points alias (e.g. "alias/leviosa-test")
// at it, the way Terraform provisions the KEK in deployed environments. It returns the key ID.
func (c *KMSContainer) CreateKey(ctx context.Context, alias string) (string, error) {
	client := c.Client()

	key, err := client.CreateKey(ctx, &kms.CreateKeyInput{Description: aws.String("encx KEK for " + alias)})
	if err != nil {
		return "", fmt.Errorf("create KMS key: %w", err)
	}
	keyID := aws.ToString(key.KeyMetadata.KeyId)

	if _, err := client.CreateAlias(ctx, &kms.CreateAliasInput{
		AliasName:   aws.String(alias),
		TargetKeyId: aws.String(keyID),
	}); err != nil {
		return "", fmt.Errorf("create KMS alias %s: %w", alias, err)
	}
	return keyID, nil
}

// EncryptionConfig returns the configuration the backend would load to use alias on this
// container with the given base64 pepper.
func (c *KMSContainer) EncryptionConfig(alias, pepper string) encryption.Config {
	return encryption.Config{
		KMSKeyAlias:        alias,
		Pepper:             pepper,
		AWSRegion:          KMSRegion,
		AWSAccessKeyID:     KMSAccessKeyID,
		AWSSecretAccessKey: KMSSecretAccessKey,
		AWSEndpointURL:     c.Endpoint,
	}
}
