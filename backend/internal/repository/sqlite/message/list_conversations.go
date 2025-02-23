package messageRepository

import (
	"context"
	"errors"

	"github.com/GaryHY/leviosa/internal/domain/message/models"
	rp "github.com/GaryHY/leviosa/internal/repository"
)

func (m *Repository) ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	query := `
        SELECT 
            id,
            user_id,
            admin_id,
            encrypted_created_at
        FROM conversations;`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return nil, rp.NewContextErr(err)
		default:
			return nil, rp.NewDatabaseErr(err)
		}
	}
	defer rows.Close()
	var conversations []*models.Conversation
	for rows.Next() {
		var conversation models.Conversation
		err := rows.Scan(
			&conversation.ID,
			&conversation.UserID,
			&conversation.AdminID,
			&conversation.EncryptedCreatedAt,
		)

		if err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		conversations = append(conversations, &conversation)
	}
	if err := rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	if len(conversations) == 0 {
		return []*models.Conversation{}, nil
	}
	return conversations, nil
}
