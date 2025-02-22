package messageService

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/message/models"
)

type Reader interface {
	GetMessages(ctx context.Context, conversationID string) ([]*models.Message, error)
	ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error)
}

type Writer interface {
	CreateConversation(ctx context.Context, conversation *models.Conversation) error
	SendMessage(ctx context.Context, message *models.Message) error
}

type ReadWriter interface {
	Reader
	Writer
}
