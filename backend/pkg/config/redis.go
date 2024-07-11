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
	addr := fmt.Sprintf("%s:%d", host, port)
	c.redis = &redisCreds{
		Addr: addr,
	}
	return nil
}
