package session_repo

import (
	"context"

	rdu "github.com/GaryHY/event-reservation-app/pkg/redisutil"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	Client *redis.Client
}

func NewSessionRepository(ctx context.Context, opts ...rdu.RedisOption) (*SessionRepository, error) {
	// Add the redis options with it, if there are any ?
	db, err := rdu.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	store := &SessionRepository{db}
	// The admin user for testing purposes.
	// the admin has the id 3439...
	// queries := make(map[string]interface{})
	// queries["session:3439434532245"] = struct {
	// 	ID     string `json:"id"`
	// 	UserID string `json:"userID"`
	// }{
	// 	"U2343r23490J4", // the session ID for the admin user.
	// 	"3439434532245",
	// }
	// rdu.Init(ctx, store.client, queries)
	return store, nil
}
