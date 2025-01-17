package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/GaryHY/leviosa/pkg/flags"
)

type sqliteCreds struct {
	Filename string `json:"filename"`
}

func (c *Config) GetSQLITE() *sqliteCreds {
	return c.sqlite
}

func (c *Config) setSQLITE(env mode.EnvMode) error {
	databaseFilename := c.viper.GetString("sqlite.filename")
	if databaseFilename == "" {
		return errors.New("'DATABASE_FILENAME' environment variable not set; please define it to specify SQLite file name")
	}
	var prefix string
	switch env {
	case mode.ModeStaging, mode.ModeDev, mode.ModeProd:
		prefix = env.String()
	default:
		return fmt.Errorf("mode value can only be 'development', 'production' or 'staging', got : %q", env)
	}
	sqliteFile := fmt.Sprintf("%s_%s", prefix, databaseFilename)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory to set SQLite database: %w", err)
	}
	c.sqlite.Filename = path.Join(wd, "data", sqliteFile)
	return nil
}
