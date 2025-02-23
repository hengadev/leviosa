package miniredis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
)

type InitMap[T any] map[string]T

func Setup(t testing.TB, ctx context.Context) (*redis.Client, error) {
	t.Helper()
	redisServer, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("create miniredis server: %w", err)
	}
	return redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	}), nil
}

func Init[T any](t testing.TB, ctx context.Context, client *redis.Client, prefix string, initMap InitMap[T]) error {
	t.Helper()
	for key, value := range initMap {
		valueEncoded, err := json.Marshal(value)
		if err != nil {
			return errors.New("encoding values")
		}
		if err = client.Set(ctx, prefix+key, valueEncoded, time.Minute).Err(); err != nil {
			return errors.New("set values")
		}
	}
	return nil
}

type redisRepository interface {
	GetClient() *redis.Client
}

type repoConstructor[T redisRepository] func(ctx context.Context, client *redis.Client) T

func SetupRepository[T redisRepository, V any](
	t testing.TB,
	ctx context.Context,
	prefix string,
	initMap InitMap[V],
	constructor repoConstructor[T],
) T {
	t.Helper()
	// setup miniredis
	client, err := Setup(t, ctx)
	if err != nil {
		t.Fatalf("setup miniredis: %s", err)
	}

	// init miniredis
	if err := Init(t, ctx, client, prefix, initMap); err != nil {
		t.Fatalf("init miniredis: %s", err)
	}

	return constructor(ctx, client)
}
