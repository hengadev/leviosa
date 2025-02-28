package messageRepository

import (
	"context"
	"errors"

	"github.com/hengadev/leviosa/internal/domain/message/models"
	rp "github.com/hengadev/leviosa/internal/repository"
)

func (m *Repository) SendMessage(ctx context.Context, message *models.Message) error {
	query := `
        INSERT INTO messages (
            id,
            conversation_id,
            sender_id,
            encrypted_content,
            encrypted_created_at
        ) VALUES (?, ?, ?, ?, ?);`
	result, err := m.DB.ExecContext(
		ctx,
		query,
		message.ID,
		message.ConversationID,
		message.SenderID,
		message.Content,
		message.EncryptedCreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			return rp.NewContextErr(err)
		default:
			return rp.NewDatabaseErr(err)
		}

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return rp.NewDatabaseErr(err)
	}
	if rowsAffected == 0 {
		return rp.NewNotCreatedErr(errors.New("no rows affected by insertion statement"), "message")
	}
	return nil
}
