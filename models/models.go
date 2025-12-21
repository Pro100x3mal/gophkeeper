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
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	Type             ItemType  `json:"type"`
	Title            string    `json:"title"`
	Metadata         string    `json:"metadata"`
	DataEncrypted    []byte    `json:"data_encrypted,omitempty"`
	DataKeyEncrypted []byte    `json:"data_key_encrypted,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
type CreateItemRequest struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Metadata   string `json:"metadata"`
	DataBase64 string `json:"data_base64,omitempty"`
}
