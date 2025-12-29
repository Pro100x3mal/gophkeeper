// Package models provides data structures and types used across the GophKeeper application.
//
// This package defines the core domain models including users, items, and encrypted data
// structures used for storing and managing sensitive information.
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain errors

var (
	// ErrUserAlreadyExists is returned when attempting to create a user with an existing username.
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrUserNotFound is returned when a user cannot be found by ID or username.
	ErrUserNotFound = errors.New("user not found")

	// ErrItemNotFound is returned when an item cannot be found.
	ErrItemNotFound = errors.New("item not found")
)

// User represents a registered user in the system.
type User struct {
	// ID is the unique identifier for the user.
	ID uuid.UUID `json:"id"`
	// Username is the unique username for authentication.
	Username string `json:"username"`
	// PasswordHash is the hashed password (never exposed in JSON).
	PasswordHash string `json:"-"`
	// CreatedAt is the timestamp when the user was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the user was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemType represents the type of stored item.
type ItemType string

const (
	// ItemTypeCredential represents username/password credentials.
	ItemTypeCredential ItemType = "credential"
	// ItemTypeText represents arbitrary text data.
	ItemTypeText ItemType = "text"
	// ItemTypeBinary represents binary data (files).
	ItemTypeBinary ItemType = "binary"
	// ItemTypeCard represents credit card information.
	ItemTypeCard ItemType = "card"
)

// Item represents a stored item with metadata.
type Item struct {
	// ID is the unique identifier for the item.
	ID uuid.UUID `json:"id"`
	// UserID is the ID of the user who owns this item.
	UserID uuid.UUID `json:"user_id"`
	// Type is the type of data stored in this item.
	Type ItemType `json:"type"`
	// Title is the user-friendly name of the item.
	Title string `json:"title"`
	// Metadata contains additional information about the item in JSON format.
	Metadata string `json:"metadata"`
	// CreatedAt is the timestamp when the item was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the item was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// EncryptedData represents encrypted data associated with an item.
type EncryptedData struct {
	// ID is the unique identifier for the encrypted data record.
	ID uuid.UUID `json:"id"`
	// ItemID is the ID of the item this encrypted data belongs to.
	ItemID uuid.UUID `json:"item_id"`
	// DataEncrypted is the encrypted data content.
	DataEncrypted []byte `json:"data_encrypted"`
	// DataKeyEncrypted is the encrypted data encryption key.
	DataKeyEncrypted []byte `json:"data_key_encrypted"`
}

// CreateItemRequest represents a request to create a new item.
type CreateItemRequest struct {
	// Type is the type of item to create.
	Type ItemType `json:"type"`
	// Title is the user-friendly name of the item.
	Title string `json:"title"`
	// Metadata contains additional information in JSON format.
	Metadata string `json:"metadata"`
	// DataBase64 is the base64-encoded data content (optional).
	DataBase64 string `json:"data_base64,omitempty"`
}

// UpdateItemRequest represents a request to update an existing item.
// All fields are optional and only provided fields will be updated.
type UpdateItemRequest struct {
	// Type is the new type for the item (optional).
	Type *ItemType `json:"type,omitempty"`
	// Title is the new title for the item (optional).
	Title *string `json:"title,omitempty"`
	// Metadata is the new metadata for the item (optional).
	Metadata *string `json:"metadata,omitempty"`
	// DataBase64 is the new base64-encoded data content (optional).
	DataBase64 *string `json:"data_base64,omitempty"`
}
