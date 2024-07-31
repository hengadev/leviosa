package config

import (
	"context"

	"github.com/go-redis/redis"
)

type redisCreds struct {
	*redis.Options
}

func (c *Config) GetRedis() *redisCreds {
	return c.redis
}

func (c *Config) setRedis(context.Context) error {
	c.redis = &redisCreds{
		&redis.Options{
			Addr:     c.viper.GetString("redis.addr"),
			Password: c.viper.GetString("redis.password"),
			DB:       c.viper.GetInt("redis.db"),
		},
	}
	return nil
}
