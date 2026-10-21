package encryption

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/hengadev/encx"
	awskeys "github.com/hengadev/encx/providers/keys/aws"
)

// pepperAlias is the encx pepper alias. PepperStore serves the same pepper for any alias;
// encx only requires a non-empty one.
const pepperAlias = "leviosa"

// ErrKeyCreationRefused is returned when encx asks the KMS provider to create a key. The
// KEK is provisioned outside the app — by Terraform in deployed environments and by the
// LocalStack init container locally — so an unresolvable alias must stop startup instead.
var ErrKeyCreationRefused = errors.New("refusing to create a KMS key: the KEK must already exist behind " + EnvKMSKeyAlias)

// NewCrypto validates cfg and builds the encx crypto service on AWS KMS with the
// configured pepper. Every entry point that encrypts or decrypts (serve, the seed
// tool, one-off run modes) builds encx through this function.
//
// The DEK cache (encx ADR-0003) is enabled unless cfg disables it: repeated reads of
// the same record within the TTL make no KMS call, so public pages that decrypt
// partner data cost one KMS round trip per record per TTL window instead of one per
// request. A zero MaxEntries or TTL in cfg leaves encx's defaults (4096 entries,
// 10 minutes).
//
// encx records which KMS key backs each KEK version in a local SQLite file. With KMS the
// alias is the source of truth (a key ID can differ between LocalStack restarts, and
// the app never rotates the KEK itself), so unless cfg.KeyMetadataDir is set that file
// lives in a fresh temporary directory and is rebuilt from the alias at every start.
func NewCrypto(ctx context.Context, cfg Config) (encx.CryptoService, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("encryption config: %w", err)
	}

	pepper, err := NewPepperStore(cfg.Pepper)
	if err != nil {
		return nil, err
	}

	keys, err := newKMSKeyProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Resolve the alias up front so a missing key or unusable credentials report
	// the KMS error itself instead of encx's attempt to create a key.
	if _, err := keys.GetKeyID(ctx, cfg.KMSKeyAlias); err != nil {
		return nil, fmt.Errorf("resolve KMS key %s: %w", cfg.KMSKeyAlias, err)
	}

	metadataDir := cfg.KeyMetadataDir
	if metadataDir == "" {
		if metadataDir, err = os.MkdirTemp("", "leviosa-encx-"); err != nil {
			return nil, fmt.Errorf("create encx key metadata directory: %w", err)
		}
	}

	deks, err := cfg.DEKCache()
	if err != nil {
		// Unreachable after Validate; kept so a malformed value can never pass silently.
		return nil, fmt.Errorf("encryption config: %w", err)
	}
	opts := []encx.Option{}
	if !deks.Disabled {
		// Zero values select encx's defaults (4096 entries, 10 minutes).
		opts = append(opts, encx.WithDEKCache(deks.MaxEntries, deks.TTL))
	}

	crypto, err := encx.NewCrypto(ctx, keys, pepper, encx.Config{
		KEKAlias:    cfg.KMSKeyAlias,
		PepperAlias: pepperAlias,
		DBPath:      metadataDir,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("create crypto service: %w", err)
	}
	return crypto, nil
}

// kmsKeyProvider is encx's AWS KMS provider with key creation disabled.
type kmsKeyProvider struct {
	*awskeys.KMSService
}

func (kmsKeyProvider) CreateKey(context.Context, string) (string, error) {
	return "", ErrKeyCreationRefused
}

func newKMSKeyProvider(ctx context.Context, cfg Config) (kmsKeyProvider, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, "",
		)),
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return kmsKeyProvider{}, fmt.Errorf("load AWS config for KMS: %w", err)
	}
	// The endpoint comes from cfg alone: LoadDefaultConfig also picks up AWS_ENDPOINT_URL
	// from the process environment, which callers building cfg by hand don't expect.
	awsCfg.BaseEndpoint = nil
	if cfg.AWSEndpointURL != "" {
		awsCfg.BaseEndpoint = aws.String(cfg.AWSEndpointURL)
	}

	service, err := awskeys.NewKMSService(ctx, awskeys.Config{AWSConfig: &awsCfg})
	if err != nil {
		return kmsKeyProvider{}, fmt.Errorf("create KMS key provider: %w", err)
	}
	return kmsKeyProvider{service}, nil
}
