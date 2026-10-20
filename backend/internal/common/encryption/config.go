// Package encryption builds the backend's encx crypto service the one way every
// entry point uses (ADR-0015): the KEK is an AWS KMS key addressed by alias, and the
// pepper comes from configuration through PepperStore.
package encryption

import (
	"errors"
	"fmt"
	"strings"
)

// Environment variables read by ConfigFromEnv.
const (
	// EnvKMSKeyAlias is the KEK's KMS alias name, e.g. "alias/leviosa-prod".
	EnvKMSKeyAlias = "ENCX_KMS_KEY_ALIAS"
	// EnvPepper is the pepper: standard base64 of exactly 32 bytes.
	EnvPepper = "ENCX_PEPPER"
	// EnvAWSRegion is the region of the KMS key.
	EnvAWSRegion = "AWS_REGION"
	// EnvAWSAccessKeyID and EnvAWSSecretAccessKey are the credentials used to call KMS.
	EnvAWSAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	// EnvAWSEndpointURL overrides the KMS endpoint; set only for LocalStack.
	EnvAWSEndpointURL = "AWS_ENDPOINT_URL"
)

const aliasPrefix = "alias/"

// Config is everything needed to build the crypto service.
type Config struct {
	// KMSKeyAlias is the KEK's alias name ("alias/<name>"). Key IDs and ARNs are not
	// accepted: encx's AWS provider would turn them into a nonexistent alias.
	KMSKeyAlias string
	// Pepper is standard base64 of exactly 32 bytes.
	Pepper string

	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	// AWSEndpointURL is empty against real AWS and points at LocalStack locally and in tests.
	AWSEndpointURL string

	// KeyMetadataDir is where encx keeps its local KEK-version database. It is not read
	// from the environment; leave it empty to get a fresh directory per process (see NewCrypto).
	KeyMetadataDir string
}

// ConfigFromEnv reads the Env* variables through getenv (typically os.Getenv).
func ConfigFromEnv(getenv func(string) string) Config {
	return Config{
		KMSKeyAlias:        getenv(EnvKMSKeyAlias),
		Pepper:             getenv(EnvPepper),
		AWSRegion:          getenv(EnvAWSRegion),
		AWSAccessKeyID:     getenv(EnvAWSAccessKeyID),
		AWSSecretAccessKey: getenv(EnvAWSSecretAccessKey),
		AWSEndpointURL:     getenv(EnvAWSEndpointURL),
	}
}

// Validate reports every missing or malformed value. A missing or invalid pepper wraps
// ErrPepperMissing or ErrPepperInvalid.
func (c Config) Validate() error {
	var errs []error

	switch alias := strings.TrimSpace(c.KMSKeyAlias); {
	case alias == "":
		errs = append(errs, fmt.Errorf("%s is required", EnvKMSKeyAlias))
	case !strings.HasPrefix(alias, aliasPrefix) || len(alias) == len(aliasPrefix):
		errs = append(errs, fmt.Errorf("%s must be a KMS alias name like %q (key IDs and ARNs are not supported), got %q",
			EnvKMSKeyAlias, "alias/leviosa-prod", alias))
	}

	if _, err := NewPepperStore(c.Pepper); err != nil {
		errs = append(errs, err)
	}

	for _, required := range []struct{ name, value string }{
		{EnvAWSRegion, c.AWSRegion},
		{EnvAWSAccessKeyID, c.AWSAccessKeyID},
		{EnvAWSSecretAccessKey, c.AWSSecretAccessKey},
	} {
		if strings.TrimSpace(required.value) == "" {
			errs = append(errs, fmt.Errorf("%s is required", required.name))
		}
	}

	return errors.Join(errs...)
}
