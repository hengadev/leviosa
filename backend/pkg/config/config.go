package config

import (
	"context"
	"fmt"
	"os"

	"github.com/GaryHY/leviosa/pkg/errsx"
	"github.com/GaryHY/leviosa/pkg/flags"

	"github.com/spf13/viper"
)

type Config struct {
	viper    *viper.Viper
	sqlite   *sqliteCreds
	redis    *redisCreds
	s3       *s3Creds
	security *SecurityConfig
}

func New(ctx context.Context, envFilename, envFileType string) *Config {
	vp := viper.New()
	if os.Getenv("APP_ENV") != "production" {
		vp.AddConfigPath(".")
		vp.SetConfigName(envFilename)
		vp.SetConfigType(envFileType)
		if err := vp.ReadInConfig(); err != nil {
			fmt.Println("viper reading :", err)
		}
	}
	return &Config{
		viper:    vp,
		sqlite:   &sqliteCreds{},
		redis:    &redisCreds{},
		s3:       &s3Creds{},
		security: &SecurityConfig{},
	}
}

func (c *Config) Load(ctx context.Context, mode mode.EnvMode) errsx.Map {
	var errs errsx.Map

	envVarsToKeys := map[string]struct {
		required bool
		key      string
	}{
		"DATABASE_FILENAME": {required: true, key: "sqlite.filename"},

		"REDIS_ADDR":     {required: true, key: "redis.addr"},
		"REDIS_DB":       {required: true, key: "redis.db"},
		"REDIS_PASSWORD": {required: true, key: "redis.password"},

		"STRIPE_SECRET_KEY": {required: true, key: "stripe.secret.key"},

		"GMAIL_EMAIL":    {required: true, key: "gmail.email"},
		"GMAIL_PASSWORD": {required: true, key: "gmail.password"},

		"AWS_REGION":            {required: true, key: "aws.region"},
		"AWS_ACCESS_KEY_ID":     {required: true, key: "aws.access.key.id"},
		"AWS_SECRET_ACCESS_KEY": {required: true, key: "aws.secret.access.key"},
		"BUCKETNAME":            {required: true, key: "s3.bucketname"},

		"USER_ENCRYPTION_KEY": {required: true, key: "user.encryption.key"},
	}
	for envVar, requiredKey := range envVarsToKeys {
		if os.Getenv(envVar) == "" && requiredKey.required == true {
			errs.Set("get environment variable", fmt.Errorf("missing required env variables: %s", envVar))
		}
		if err := c.viper.BindEnv(requiredKey.key, envVar); err != nil {
			errs.Set("bind environment variable", fmt.Errorf("bind env: %w", err))
		}
	}
	if err := c.setSQLITE(mode); err != nil {
		errs.Set("sqlite configuration", fmt.Errorf("set SQLITE: %w", err))
	}
	if err := c.setRedis(mode); err != nil {
		errs.Set("redis configuration", fmt.Errorf("set Redis: %w", err))
	}
	if err := c.setS3(mode); err != nil {
		errs.Set("S3 configuration", fmt.Errorf("set S3: %w", err))
	}
	if err := c.setSecurityConfig(ctx); err != nil {
		errs.Set("user security configuration", fmt.Errorf("set user security config: %w", err))
	}
	return errs
}
