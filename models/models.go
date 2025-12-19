package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ItemType string

const (
	ItemTypeCredential ItemType = "credential"
	ItemTypeText       ItemType = "text"
	ItemTypeBinary     ItemType = "binary"
	ItemTypeCard       ItemType = "card"
)

type Item struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	Type             ItemType        `json:"type"`
	Title            string          `json:"title"`
	Metadata         json.RawMessage `json:"metadata"`
	DataEncrypted    []byte          `json:"data_encrypted,omitempty"`
	DataEncryptedKey []byte          `json:"data_encrypted_key,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
type PlainItem struct {
	UserID   uuid.UUID
	Type     ItemType
	Title    string
	Metadata json.RawMessage
	Data     []byte
}
