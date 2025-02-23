package models

import "time"

type Conversation struct {
	ID                 string
	UserID             string
	AdminID            string
	CreatedAt          time.Time
	EncryptedCreatedAt string
}
