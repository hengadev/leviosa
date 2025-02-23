package redisutil

import "github.com/redis/go-redis/v9"

func WithAddr(addr string) RedisOption {
	return func(r *redis.Options) {
		r.Addr = addr
	}
}
