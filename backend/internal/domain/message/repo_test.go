package messageService_test

import (
	"context"

	"github.com/GaryHY/leviosa/internal/domain/message/models"
)

type MockRepo struct {
	GetMessagesFunc        func(ctx context.Context, conversationID string) ([]*models.Message, error)
	ListConversationsFunc  func(ctx context.Context, userID string) ([]*models.Conversation, error)
	CreateConversationFunc func(ctx context.Context, conversation *models.Conversation) error
	SendMessageFunc        func(ctx context.Context, message *models.Message) error
}

func (m *MockRepo) GetMessages(ctx context.Context, conversationID string) ([]*models.Message, error) {
	if m.GetMessagesFunc != nil {
		return m.GetMessagesFunc(ctx, conversationID)
	}
	return nil, nil
}

func (m *MockRepo) ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	if m.ListConversationsFunc != nil {
		return m.ListConversationsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepo) CreateConversation(ctx context.Context, conversation *models.Conversation) error {
	if m.CreateConversationFunc != nil {
		return m.CreateConversationFunc(ctx, conversation)
	}
	return nil
}

func (m *MockRepo) SendMessage(ctx context.Context, message *models.Message) error {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, message)
	}
	return nil
}
