package messageRepository

import (
	"context"
	"errors"

	"github.com/hengadev/leviosa/internal/domain/message/models"
	rp "github.com/hengadev/leviosa/internal/repository"
)

func (m *Repository) GetMessages(ctx context.Context, conversationID string) ([]*models.Message, error) {
	query := `
        SELECT 
            id,
            conversation_id,
            sender_id,
            encrypted_content,
            encrypted_created_at
        FROM messages;`
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

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderID,
			&message.Content,
			&message.EncryptedCreatedAt,
		)
		if err != nil {
			return nil, rp.NewDatabaseErr(err)
		}
		messages = append(messages, &message)
	}
	if err := rows.Err(); err != nil {
		return nil, rp.NewDatabaseErr(err)
	}
	if len(messages) == 0 {
		return []*models.Message{}, nil
	}
	return messages, nil
}
