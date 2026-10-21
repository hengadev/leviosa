package encryption

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestConfigFromEnv_ReadsEveryVariable(t *testing.T) {
	env := map[string]string{
		EnvKMSKeyAlias:        "alias/leviosa-dev",
		EnvPepper:             "cGVwcGVy",
		EnvAWSRegion:          "eu-central-1",
		EnvAWSAccessKeyID:     "AKIDEXAMPLE",
		EnvAWSSecretAccessKey: "secret",
		EnvAWSEndpointURL:     "http://localstack:4566",
		EnvDEKCacheMaxEntries: "2048",
		EnvDEKCacheTTL:        "5m",
	}

	got := ConfigFromEnv(func(key string) string { return env[key] })

	want := Config{
		KMSKeyAlias:        "alias/leviosa-dev",
		Pepper:             "cGVwcGVy",
		AWSRegion:          "eu-central-1",
		AWSAccessKeyID:     "AKIDEXAMPLE",
		AWSSecretAccessKey: "secret",
		AWSEndpointURL:     "http://localstack:4566",
		DEKCacheMaxEntries: "2048",
		DEKCacheTTL:        "5m",
	}
	if got != want {
		t.Fatalf("ConfigFromEnv = %+v, want %+v", got, want)
	}
}

func validConfig() Config {
	return Config{
		KMSKeyAlias:        "alias/leviosa-dev",
		Pepper:             "bGV2aW9zYS1sb2NhbC1kZXYtcGVwcGVyLTMyYnl0ZXM=",
		AWSRegion:          "us-east-1",
		AWSAccessKeyID:     "test",
		AWSSecretAccessKey: "test",
	}
}

func TestConfigValidate_AcceptsCompleteConfig(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestConfigValidate_ReportsEachMissingValue(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{"alias", func(c *Config) { c.KMSKeyAlias = "" }, EnvKMSKeyAlias},
		{"region", func(c *Config) { c.AWSRegion = "" }, EnvAWSRegion},
		{"access key", func(c *Config) { c.AWSAccessKeyID = "" }, EnvAWSAccessKeyID},
		{"secret key", func(c *Config) { c.AWSSecretAccessKey = "" }, EnvAWSSecretAccessKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(&cfg)
			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate error = %v, want it to name %s", err, tt.want)
			}
		})
	}
}

func TestConfigValidate_RejectsMissingPepper(t *testing.T) {
	cfg := validConfig()
	cfg.Pepper = ""
	if err := cfg.Validate(); !errors.Is(err, ErrPepperMissing) {
		t.Fatalf("Validate error = %v, want ErrPepperMissing", err)
	}
}

func TestConfigValidate_RejectsWrongLengthPepper(t *testing.T) {
	cfg := validConfig()
	cfg.Pepper = "c2hvcnQ=" // "short"
	if err := cfg.Validate(); !errors.Is(err, ErrPepperInvalid) {
		t.Fatalf("Validate error = %v, want ErrPepperInvalid", err)
	}
}

// encx's AWS provider prefixes anything that doesn't start with "alias/" with "alias/",
// so a key ID or ARN would silently become a nonexistent alias. Only alias names are accepted.
func TestConfigValidate_RequiresAnAliasName(t *testing.T) {
	for _, alias := range []string{
		"leviosa-dev",
		"1234abcd-12ab-34cd-56ef-1234567890ab",
		"arn:aws:kms:eu-central-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
		"arn:aws:kms:eu-central-1:111122223333:alias/leviosa-prod",
		"alias/",
	} {
		cfg := validConfig()
		cfg.KMSKeyAlias = alias
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), EnvKMSKeyAlias) {
			t.Errorf("Validate(%q) error = %v, want an %s error", alias, err, EnvKMSKeyAlias)
		}
	}
}

// Unset cache variables mean "enabled with encx's defaults"; explicit values must parse.
func TestDEKCache(t *testing.T) {
	tests := []struct {
		name             string
		maxEntries, ttl  string
		want             DEKCacheConfig
		wantProblemInErr string
	}{
		{
			name: "unset means enabled with encx defaults",
			want: DEKCacheConfig{},
		},
		{
			name: "blank means unset",
			want: DEKCacheConfig{},
		},
		{
			name:       "explicit values",
			maxEntries: "2000",
			ttl:        "15m",
			want:       DEKCacheConfig{MaxEntries: 2000, TTL: 15 * time.Minute},
		},
		{
			name: "compound durations",
			ttl:  "1h30m",
			want: DEKCacheConfig{TTL: 90 * time.Minute},
		},
		{
			name:       "zero entries disables the cache",
			maxEntries: "0",
			want:       DEKCacheConfig{Disabled: true},
		},
		{
			name:             "negative entries",
			maxEntries:       "-1",
			wantProblemInErr: EnvDEKCacheMaxEntries,
		},
		{
			name:             "non-integer entries",
			maxEntries:       "lots",
			wantProblemInErr: EnvDEKCacheMaxEntries,
		},
		{
			name:             "zero ttl",
			ttl:              "0s",
			wantProblemInErr: EnvDEKCacheTTL,
		},
		{
			name:             "negative ttl",
			ttl:              "-5m",
			wantProblemInErr: EnvDEKCacheTTL,
		},
		{
			name:             "unitless ttl",
			ttl:              "600",
			wantProblemInErr: EnvDEKCacheTTL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.DEKCacheMaxEntries = tt.maxEntries
			cfg.DEKCacheTTL = tt.ttl

			got, err := cfg.DEKCache()
			if tt.wantProblemInErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantProblemInErr) {
					t.Fatalf("DEKCache error = %v, want it to name %s", err, tt.wantProblemInErr)
				}
				if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), tt.wantProblemInErr) {
					t.Fatalf("Validate error = %v, want it to name %s too", err, tt.wantProblemInErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("DEKCache: %v", err)
			}
			if got != tt.want {
				t.Fatalf("DEKCache = %+v, want %+v", got, tt.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
		})
	}
}
