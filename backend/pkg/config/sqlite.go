package config

import (
	"context"
)

type sqliteCreds struct {
	Filename string `json:"filename"`
}

func (c *Config) GetSQLITE() *sqliteCreds {
	return c.sqlite
}

func (c *Config) setSQLITE(context.Context) error {
	c.sqlite.Filename = c.viper.GetString("sqlite.filename")
	return nil
}
