package config

import (
	"context"
)

type sqliteCreds struct {
	Filename string `json:"filename"`
}

func (c *Config) GetSqlite() *sqliteCreds {
	return c.sqlite
}

func (c *Config) setSQLITE(context.Context) error {
	return nil
}
