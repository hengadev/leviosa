package redisutil

import "github.com/redis/go-redis/v9"

func WithDB(DB int) RedisOption {
	return func(r *redis.Options) {
		r.DB = DB
	}
}
