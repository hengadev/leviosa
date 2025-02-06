package redisutil

import (
	"github.com/redis/go-redis/v9"
)

type RedisOption func(*redis.Options)
