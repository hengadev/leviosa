package redis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	rd "github.com/GaryHY/event-reservation-app/internal/redis"

	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
)

const sessionID = "a0rg34tWfQ33009_K"

var sessionTest = session.Session{
	ID:         sessionID,
	UserID:     1,
	Role:       "basic",
	LoggedInAt: time.Now(),
	CreatedAt:  time.Now(),
	ExpiresAt:  time.Now().Add(time.Hour),
}

var redisServer *miniredis.Miniredis

type Init struct {
	key   string
	value *session.Values
}

func setupSessionRepo(ctx context.Context, t *testing.T, inits ...Init) (*rd.SessionRepository, error) {
	t.Helper()
	redisServer, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("create database server: %w", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})
	repo := rd.NewSessionRepository(ctx, client)
	return repo, nil
}

func initRepo(ctx context.Context, repo *rd.SessionRepository, inits ...Init) (string, error) {
	for _, i := range inits {
		valueEncoded, err := json.Marshal(i.value)
		if err != nil {
			return "", fmt.Errorf("encoding")
		}
		if err = repo.Client.Set(ctx, i.key, valueEncoded, time.Minute).Err(); err != nil {
			fmt.Println("the key is", i.key)
			fmt.Println("the value is", i.value)
			return "", fmt.Errorf("init database: %w", err)
		}
	}
	return "", nil
}
