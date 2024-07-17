package config

import (
	"context"
	"fmt"
)

type redisCreds struct {
	Addr string
	DB   string
}

func (c *Config) GetRedis() *redisCreds {
	return c.redis
}

func (c *Config) setRedis(context.Context) error {
	host := c.viper.GetString("redis.host")
	port := c.viper.GetInt("redis.port")
	c.redis = &redisCreds{
		Addr: fmt.Sprintf("%s:%d", host, port),
	}
	return nil
}
