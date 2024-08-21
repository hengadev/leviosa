package miniredis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
)

// Helper to run miniredis for redis related unit tests.

var redisServer *miniredis.Miniredis

type InitMap map[string]any

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

func Init(t testing.TB, ctx context.Context, client *redis.Client, initMap InitMap) error {
	t.Helper()
	fail := func(err string) error {
		return fmt.Errorf("init miniredis: %s,", err)
	}
	for key, value := range initMap {
		valueEncoded, err := json.Marshal(value)
		if err != nil {
			return fail("encoding values")
		}
		if err = client.Set(ctx, key, valueEncoded, time.Minute).Err(); err != nil {
			return fail("set values")
		}
	}
	return nil
}
