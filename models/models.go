package models

import (
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
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      ItemType  `json:"type"`
	Title     string    `json:"title"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EncryptedData struct {
	ID               uuid.UUID `json:"id"`
	ItemID           uuid.UUID `json:"item_id"`
	DataEncrypted    []byte    `json:"data_encrypted"`
	DataKeyEncrypted []byte    `json:"data_key_encrypted"`
}

type CreateItemRequest struct {
	Type       ItemType `json:"type"`
	Title      string   `json:"title"`
	Metadata   string   `json:"metadata"`
	DataBase64 string   `json:"data_base64,omitempty"`
}
