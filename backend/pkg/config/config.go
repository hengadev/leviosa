package config

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	viper  *viper.Viper
	sqlite *sqliteCreds
	redis  *redisCreds
	s3     *s3Creds
}

func New(context.Context) *Config {
	return &Config{
		viper: viper.New(),
	}
}

func (c *Config) Load(ctx context.Context) error {
	// TODO: add all the env variables that I need, sqlite, redis, S3 bucket etc..
	envVarsToKeys := map[string]struct {
		required bool
		key      string
	}{
		"DATABASE_FILENAME": {true, "database.filename"},
	}
	for envVar, requiredKey := range envVarsToKeys {
		if os.Getenv(envVar) == "" && requiredKey.required == true {
			return fmt.Errorf("Missing required env variables: %s", envVar)
		}
		if err := c.viper.BindEnv(envVar, requiredKey.key); err != nil {
			return fmt.Errorf("bind env: %w", err)
		}
	}
	if err := c.setSQLITE(ctx); err != nil {
		return fmt.Errorf("set SQLITE: %w", err)
	}
	return nil
}
