package redisutil

import (
	"context"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

type RedisOption func(*redis.Options)

// the with things here do not serve because they are not optional, soo.....
func WithAddr(addr string) RedisOption {
	return func(r *redis.Options) {
		r.Addr = addr
	}
}

func WithDB(DB int) RedisOption {
	return func(r *redis.Options) {
		r.DB = DB
	}
}

func WithPassword(pwd string) RedisOption {
	return func(r *redis.Options) {
		r.Password = pwd
	}
}

func Connect(ctx context.Context, opts ...RedisOption) (*redis.Client, error) {
	r := &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}
	for _, opt := range opts {
		opt(r)
	}
	client := redis.NewClient(r)
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// TODO: find a value for the expiration of the value set
func Init(ctx context.Context, client *redis.Client, queries map[string]interface{}) error {
	for k, v := range queries {
		err := client.Set(ctx, k, v, session.SessionDuration).Err()
		if err != nil {
			return rp.NewRessourceCreationErr(err)
		}
	}
	return nil
}
