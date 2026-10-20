package encryption

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigFromEnv_ReadsEveryVariable(t *testing.T) {
	env := map[string]string{
		EnvKMSKeyAlias:        "alias/leviosa-dev",
		EnvPepper:             "cGVwcGVy",
		EnvAWSRegion:          "eu-central-1",
		EnvAWSAccessKeyID:     "AKIDEXAMPLE",
		EnvAWSSecretAccessKey: "secret",
		EnvAWSEndpointURL:     "http://localstack:4566",
	}

	got := ConfigFromEnv(func(key string) string { return env[key] })

	want := Config{
		KMSKeyAlias:        "alias/leviosa-dev",
		Pepper:             "cGVwcGVy",
		AWSRegion:          "eu-central-1",
		AWSAccessKeyID:     "AKIDEXAMPLE",
		AWSSecretAccessKey: "secret",
		AWSEndpointURL:     "http://localstack:4566",
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
