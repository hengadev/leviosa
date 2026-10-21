// Package encryption builds the backend's encx crypto service the one way every
// entry point uses (ADR-0015): the KEK is an AWS KMS key addressed by alias, and the
// pepper comes from configuration through PepperStore.
package encryption

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
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
	// EnvDEKCacheMaxEntries bounds encx's in-memory cache of decrypted data keys
	// (issue 031): without it, every read of an encrypted record makes one KMS call.
	// Unset leaves encx's default bound (4096 entries); "0" disables the cache; any
	// other value must be a positive integer and is the bound itself.
	EnvDEKCacheMaxEntries = "ENCX_DEK_CACHE_MAX_ENTRIES"
	// EnvDEKCacheTTL is how long a decrypted data key stays cached, as a Go duration
	// string ("10m", "45s", "1h30m"). Unset leaves encx's default (10 minutes).
	EnvDEKCacheTTL = "ENCX_DEK_CACHE_TTL"
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

	// DEKCacheMaxEntries and DEKCacheTTL are the raw EnvDEKCache* values. DEKCache
	// parses them and Validate checks them, so a malformed value surfaces at startup
	// together with every other config error.
	DEKCacheMaxEntries string
	DEKCacheTTL        string
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
		DEKCacheMaxEntries: getenv(EnvDEKCacheMaxEntries),
		DEKCacheTTL:        getenv(EnvDEKCacheTTL),
	}
}

// DEKCacheConfig is the parsed form of the EnvDEKCache* variables: encx's bounded
// in-memory cache of decrypted data keys (encx ADR-0003), which NewCrypto enables by
// default. Public pages that decrypt partner data would otherwise make one sequential
// KMS round trip per record on every anonymous visit.
type DEKCacheConfig struct {
	// Disabled is true when EnvDEKCacheMaxEntries is "0": NewCrypto builds encx without
	// the cache, so every read goes back to the KMS. The escape hatch for an operator who
	// needs a KEK revocation or the KMS audit trail to see every read.
	Disabled bool
	// MaxEntries is the cache's entry bound. Zero leaves the bound to encx's default
	// (4096 entries; a few thousand, per Leviosa's ADR-0015).
	MaxEntries int
	// TTL is how long a decrypted data key is kept. Zero leaves it to encx's default
	// (10 minutes). The TTL is absolute: reads never extend it.
	TTL time.Duration
}

// DEKCache parses the EnvDEKCache* values. An empty MaxEntries enables the cache with
// encx's default bound, "0" disables it, and any other value must be a positive integer.
// An empty TTL leaves encx's default; any other value must be a positive duration.
func (c Config) DEKCache() (DEKCacheConfig, error) {
	var parsed DEKCacheConfig

	switch raw := strings.TrimSpace(c.DEKCacheMaxEntries); {
	case raw == "":
		// Enabled, encx's default bound.
	case raw == "0":
		parsed.Disabled = true
	default:
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return parsed, fmt.Errorf("%s must be a non-negative integer (0 disables the DEK cache), got %q",
				EnvDEKCacheMaxEntries, c.DEKCacheMaxEntries)
		}
		parsed.MaxEntries = n
	}

	if raw := strings.TrimSpace(c.DEKCacheTTL); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return parsed, fmt.Errorf("%s must be a positive duration like %q or %q, got %q",
				EnvDEKCacheTTL, "10m", "45s", c.DEKCacheTTL)
		}
		parsed.TTL = d
	}

	return parsed, nil
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

	if _, err := c.DEKCache(); err != nil {
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
