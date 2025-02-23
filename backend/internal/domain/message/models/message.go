package models

import "time"

type Message struct {
	ID                 string
	ConversationID     string
	SenderID           string
	Content            string
	CreatedAt          time.Time
	EncryptedCreatedAt string
}
