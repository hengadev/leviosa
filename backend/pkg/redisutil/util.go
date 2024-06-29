package redisutil

import (
	"context"
	"fmt"
	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rp "github.com/GaryHY/event-reservation-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

type RedisOption func(*redis.Options)

func WithPort(addr int) RedisOption {
	return func(r *redis.Options) {
		r.Addr = fmt.Sprintf("localhost:%d", addr)
	}
}

func WithPassword(pwd string) RedisOption {
	return func(r *redis.Options) {
		r.Password = pwd
	}
}

func Connect(ctx context.Context, opts ...RedisOption) (*redis.Client, error) {
	// TODO: Use the env variables or vault instances
	// TODO: Put default values so that if they are not set, I still have something
	r := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	// use the values set if there are any
	for _, opt := range opts {
		opt(r)
	}
	client := redis.NewClient(r)
	ping, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println(ping)
	return client, nil
}

// TODO: find a value for the expiration of the value set

func Init(ctx context.Context, client *redis.Client, queries map[string]interface{}) error {
	for k, v := range queries {
		err := client.Set(ctx, k, v, session.SessionExpirationDuration).Err()
		if err != nil {
			return rp.NewRessourceCreationErr(err)
		}
	}
	return nil
}
