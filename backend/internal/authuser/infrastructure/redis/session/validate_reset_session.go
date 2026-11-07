package sessionRepository

import (
	"context"

	"github.com/Leviosa-care/leviosa/backend/internal/common/errs"
)

func (r *SessionRepository) ValidateResetSession(ctx context.Context, tokenHash string) (string, error) {
	key := FormatResetSessionKey(tokenHash)

	// Read and delete in one step (single-use): with a GET then a DEL, two
	// concurrent confirmations could both read the token before either
	// deleted it.
	userEmailHash, err := r.client.GetDel(ctx, key).Result()
	if err != nil {
		return "", errs.ClassifyRedisError("validate reset session", err)
	}

	return userEmailHash, nil
}

