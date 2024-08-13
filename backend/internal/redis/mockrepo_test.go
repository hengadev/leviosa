package redis_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
)

// Je fais le meme truc avec un autre client

type StubSessionRepository struct {
	Client *Client
}

func NewStubSessionRepositroy(client *Client) *StubSessionRepository {
	return &StubSessionRepository{Client: client}
}

// func (s *StubSessionRepository) FindSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
// 	return session, nil
// }

type StringCmd session.Values

func (s *StringCmd) Bytes() ([]byte, error) {
	return json.Marshal(s)
}

type KVMap map[string]StringCmd

type Client struct {
	sessions KVMap
}

func (c *Client) Get(ctx context.Context, key string) (*StringCmd, error) {
	val, ok := c.sessions[key]
	if !ok {
		return nil, fmt.Errorf("key does not exists in database")
	}
	return &val, nil
}

func (c *Client) Set(ctx context.Context, key string, value StringCmd) error {
	c.sessions[key] = value
	return nil
}

func (c *Client) Del(ctx context.Context, key string) error {
	delete(c.sessions, key)
	return nil
}
