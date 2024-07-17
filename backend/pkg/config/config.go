package config

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	viper *viper.Viper

	sqlite *sqliteCreds
	redis  *redisCreds
	s3     *s3Creds
}

// TODO: Add the other creds needed
func New(ctx context.Context, envFilename, envFileType string) *Config {
	vp := viper.New()
	vp.AddConfigPath(".")
	vp.SetConfigName(envFilename)
	vp.SetConfigType(envFileType)
	if err := vp.ReadInConfig(); err != nil {
		fmt.Println("viper reading :", err)
	}
	return &Config{
		viper:  vp,
		sqlite: &sqliteCreds{},
	}
}

func (c *Config) Load(ctx context.Context) error {
	envVarsToKeys := map[string]struct {
		required bool
		key      string
	}{
		"DATABASE_FILENAME": {required: true, key: "sqlite.filename"},
	}
	for envVar, requiredKey := range envVarsToKeys {
		if os.Getenv(envVar) == "" && requiredKey.required == true {
			return fmt.Errorf("missing required env variables: %s", envVar)
		}
		if err := c.viper.BindEnv(requiredKey.key, envVar); err != nil {
			return fmt.Errorf("bind env: %w", err)
		}
	}
	if err := c.setSQLITE(ctx); err != nil {
		return fmt.Errorf("set SQLITE: %w", err)
	}
	return nil
}
