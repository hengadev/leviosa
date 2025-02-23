package redisutil

import "github.com/redis/go-redis/v9"

func WithPassword(pwd string) RedisOption {
	return func(r *redis.Options) {
		r.Password = pwd
	}
}
