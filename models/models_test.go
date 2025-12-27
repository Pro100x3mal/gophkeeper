package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestItemType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		itemType ItemType
		expected string
	}{
		{"Credential type", ItemTypeCredential, "credential"},
		{"Text type", ItemTypeText, "text"},
		{"Binary type", ItemTypeBinary, "binary"},
		{"Card type", ItemTypeCard, "card"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.itemType))
		})
	}
}

func TestUser_Struct(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	user := User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: "hashed_password",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestItem_Struct(t *testing.T) {
	itemID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	item := Item{
		ID:        itemID,
		UserID:    userID,
		Type:      ItemTypeCredential,
		Title:     "Test Item",
		Metadata:  `{"key": "value"}`,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, userID, item.UserID)
	assert.Equal(t, ItemTypeCredential, item.Type)
	assert.Equal(t, "Test Item", item.Title)
	assert.Equal(t, `{"key": "value"}`, item.Metadata)
	assert.Equal(t, now, item.CreatedAt)
	assert.Equal(t, now, item.UpdatedAt)
}

func TestEncryptedData_Struct(t *testing.T) {
	id := uuid.New()
	itemID := uuid.New()
	dataEncrypted := []byte("encrypted_data")
	dataKeyEncrypted := []byte("encrypted_key")

	encData := EncryptedData{
		ID:               id,
		ItemID:           itemID,
		DataEncrypted:    dataEncrypted,
		DataKeyEncrypted: dataKeyEncrypted,
	}

	assert.Equal(t, id, encData.ID)
	assert.Equal(t, itemID, encData.ItemID)
	assert.Equal(t, dataEncrypted, encData.DataEncrypted)
	assert.Equal(t, dataKeyEncrypted, encData.DataKeyEncrypted)
}

func TestCreateItemRequest_Struct(t *testing.T) {
	req := CreateItemRequest{
		Type:       ItemTypeText,
		Title:      "My Note",
		Metadata:   `{"tags": ["personal"]}`,
		DataBase64: "dGVzdCBkYXRh",
	}

	assert.Equal(t, ItemTypeText, req.Type)
	assert.Equal(t, "My Note", req.Title)
	assert.Equal(t, `{"tags": ["personal"]}`, req.Metadata)
	assert.Equal(t, "dGVzdCBkYXRh", req.DataBase64)
}

func TestUpdateItemRequest_Struct(t *testing.T) {
	newType := ItemTypeBinary
	newTitle := "Updated Title"
	newMetadata := `{"updated": true}`
	newData := "bmV3IGRhdGE="

	req := UpdateItemRequest{
		Type:       &newType,
		Title:      &newTitle,
		Metadata:   &newMetadata,
		DataBase64: &newData,
	}

	assert.NotNil(t, req.Type)
	assert.Equal(t, ItemTypeBinary, *req.Type)
	assert.NotNil(t, req.Title)
	assert.Equal(t, "Updated Title", *req.Title)
	assert.NotNil(t, req.Metadata)
	assert.Equal(t, `{"updated": true}`, *req.Metadata)
	assert.NotNil(t, req.DataBase64)
	assert.Equal(t, "bmV3IGRhdGE=", *req.DataBase64)
}

func TestUpdateItemRequest_NilFields(t *testing.T) {
	req := UpdateItemRequest{}

	assert.Nil(t, req.Type)
	assert.Nil(t, req.Title)
	assert.Nil(t, req.Metadata)
	assert.Nil(t, req.DataBase64)
}
