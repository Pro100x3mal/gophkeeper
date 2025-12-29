package services

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockKeyRepo is a mock implementation of KeyRepo
type MockKeyRepo struct {
	mock.Mock
}

func (m *MockKeyRepo) Save(ctx context.Context, userID uuid.UUID, enc []byte) error {
	args := m.Called(ctx, userID, enc)
	return args.Error(0)
}

func (m *MockKeyRepo) Load(ctx context.Context, userID uuid.UUID) ([]byte, bool, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]byte), args.Bool(1), args.Error(2)
}

// MockItemRepo is a mock implementation of ItemRepoInterface
type MockItemRepo struct {
	mock.Mock
}

func (m *MockItemRepo) Create(ctx context.Context, item *models.Item, encData *models.EncryptedData) error {
	args := m.Called(ctx, item, encData)
	return args.Error(0)
}

func (m *MockItemRepo) GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, *models.EncryptedData, error) {
	args := m.Called(ctx, userID, itemID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var encData *models.EncryptedData
	if args.Get(1) != nil {
		encData = args.Get(1).(*models.EncryptedData)
	}
	return args.Get(0).(*models.Item), encData, args.Error(2)
}

func (m *MockItemRepo) DeleteByID(ctx context.Context, userID, itemID uuid.UUID) error {
	args := m.Called(ctx, userID, itemID)
	return args.Error(0)
}

func (m *MockItemRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func (m *MockItemRepo) Update(
	ctx context.Context,
	userID, itemID uuid.UUID,
	req *models.UpdateItemRequest,
	encData *models.EncryptedData,
) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID, req, encData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func TestNewItemService(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")

	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	assert.NotNil(t, service)
	assert.Equal(t, mockKeyRepo, service.keyRepo)
	assert.Equal(t, mockItemRepo, service.itemRepo)
	assert.Equal(t, masterKey, service.masterKey)
}

func TestIsValidType(t *testing.T) {
	tests := []struct {
		name     string
		itemType models.ItemType
		expected bool
	}{
		{"Valid credential type", models.ItemTypeCredential, true},
		{"Valid text type", models.ItemTypeText, true},
		{"Valid binary type", models.ItemTypeBinary, true},
		{"Valid card type", models.ItemTypeCard, true},
		{"Invalid type", models.ItemType("invalid"), false},
		{"Empty type", models.ItemType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidType(tt.itemType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecodeBase64_Success(t *testing.T) {
	data := "test data"
	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	decoded, err := decodeBase64(encoded)

	require.NoError(t, err)
	assert.Equal(t, []byte(data), decoded)
}

func TestDecodeBase64_EmptyString(t *testing.T) {
	decoded, err := decodeBase64("")

	require.NoError(t, err)
	assert.Nil(t, decoded)
}

func TestDecodeBase64_InvalidBase64(t *testing.T) {
	decoded, err := decodeBase64("not-valid-base64!!!")

	assert.Error(t, err)
	assert.Nil(t, decoded)
}

func TestItemService_CreateItem_InvalidType(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	req := &models.CreateItemRequest{
		Type:     models.ItemType("invalid"),
		Title:    "Test",
		Metadata: "{}",
	}

	item, err := service.CreateItem(ctx, req, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidItemType, err)
	assert.Nil(t, item)
}

func TestItemService_CreateItem_WithoutData(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	req := &models.CreateItemRequest{
		Type:       models.ItemTypeText,
		Title:      "Test Item",
		Metadata:   "{}",
		DataBase64: "",
	}

	mockItemRepo.On("Create", ctx, mock.AnythingOfType("*models.Item"), (*models.EncryptedData)(nil)).Return(nil)

	item, err := service.CreateItem(ctx, req, userID)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, req.Title, item.Title)
	assert.Equal(t, req.Type, item.Type)

	mockItemRepo.AssertExpectations(t)
}

func TestItemService_CreateItem_InvalidBase64(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	req := &models.CreateItemRequest{
		Type:       models.ItemTypeText,
		Title:      "Test",
		Metadata:   "{}",
		DataBase64: "invalid-base64!!!",
	}

	item, err := service.CreateItem(ctx, req, userID)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to decode base64")
}

func TestItemService_ListItems_Success(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	expectedItems := []*models.Item{
		{ID: uuid.New(), UserID: userID, Type: models.ItemTypeText, Title: "Item 1"},
		{ID: uuid.New(), UserID: userID, Type: models.ItemTypeCredential, Title: "Item 2"},
	}

	mockItemRepo.On("ListByUser", ctx, userID).Return(expectedItems, nil)

	items, err := service.ListItems(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, expectedItems, items)
	assert.Len(t, items, 2)

	mockItemRepo.AssertExpectations(t)
}

func TestItemService_ListItems_Error(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()

	mockItemRepo.On("ListByUser", ctx, userID).Return(nil, errors.New("database error"))

	items, err := service.ListItems(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "failed to list items")

	mockItemRepo.AssertExpectations(t)
}

func TestItemService_DeleteItem_Success(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()

	mockItemRepo.On("DeleteByID", ctx, userID, itemID).Return(nil)

	err := service.DeleteItem(ctx, userID, itemID)

	require.NoError(t, err)
	mockItemRepo.AssertExpectations(t)
}

func TestItemService_DeleteItem_Error(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()

	mockItemRepo.On("DeleteByID", ctx, userID, itemID).Return(errors.New("database error"))

	err := service.DeleteItem(ctx, userID, itemID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete item")

	mockItemRepo.AssertExpectations(t)
}

func TestItemService_UpdateItem_InvalidType(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()
	invalidType := models.ItemType("invalid")
	req := &models.UpdateItemRequest{
		Type: &invalidType,
	}

	item, err := service.UpdateItem(ctx, userID, itemID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidItemType, err)
	assert.Nil(t, item)
}

func TestItemService_UpdateItem_WithoutData(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()
	newTitle := "Updated Title"
	req := &models.UpdateItemRequest{
		Title: &newTitle,
	}

	updatedItem := &models.Item{
		ID:     itemID,
		UserID: userID,
		Title:  newTitle,
	}

	mockItemRepo.On("Update", ctx, userID, itemID, req, (*models.EncryptedData)(nil)).Return(updatedItem, nil)

	item, err := service.UpdateItem(ctx, userID, itemID, req)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, newTitle, item.Title)

	mockItemRepo.AssertExpectations(t)
}

func TestErrInvalidItemType(t *testing.T) {
	assert.Equal(t, "invalid item type", ErrInvalidItemType.Error())
}

func TestItemService_CreateItem_WithEncryptedData(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("12345678901234567890123456789012") // exactly 32 bytes
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	dataBase64 := "dGVzdCBkYXRh" // base64 of "test data"

	req := &models.CreateItemRequest{
		Type:       models.ItemTypeText,
		Title:      "Test Item",
		Metadata:   "{}",
		DataBase64: dataBase64,
	}

	// Mock key repository - no existing key
	mockKeyRepo.On("Load", ctx, userID).Return([]byte{}, false, nil)
	mockKeyRepo.On("Save", ctx, userID, mock.Anything).Return(nil)

	mockItemRepo.On("Create", ctx, mock.AnythingOfType("*models.Item"), mock.AnythingOfType("*models.EncryptedData")).Return(nil)

	item, err := service.CreateItem(ctx, req, userID)

	require.NoError(t, err)
	assert.NotNil(t, item)
	mockKeyRepo.AssertExpectations(t)
	mockItemRepo.AssertExpectations(t)
}

func TestItemService_CreateItem_KeyLoadError(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	req := &models.CreateItemRequest{
		Type:       models.ItemTypeText,
		Title:      "Test",
		Metadata:   "{}",
		DataBase64: "dGVzdA==",
	}

	mockKeyRepo.On("Load", ctx, userID).Return([]byte{}, false, errors.New("db error"))

	item, err := service.CreateItem(ctx, req, userID)

	assert.Error(t, err)
	assert.Nil(t, item)
	mockKeyRepo.AssertExpectations(t)
}

func TestItemService_GetItem_WithoutEncryptedData(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()

	item := &models.Item{ID: itemID, UserID: userID, Title: "Test"}

	mockItemRepo.On("GetByID", ctx, userID, itemID).Return(item, nil, nil)

	gotItem, gotData, err := service.GetItem(ctx, userID, itemID)

	require.NoError(t, err)
	assert.Equal(t, item, gotItem)
	assert.Nil(t, gotData)
	mockItemRepo.AssertExpectations(t)
}

func TestItemService_GetItem_Error(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()

	mockItemRepo.On("GetByID", ctx, userID, itemID).
		Return(nil, nil, errors.New("database error"))

	item, data, err := service.GetItem(ctx, userID, itemID)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.Nil(t, data)
	mockItemRepo.AssertExpectations(t)
}

func TestItemService_UpdateItem_InvalidBase64(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()

	invalidData := "not-valid-base64!!!"
	req := &models.UpdateItemRequest{
		DataBase64: &invalidData,
	}

	item, err := service.UpdateItem(ctx, userID, itemID, req)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "failed to decode base64")
}

func TestItemService_UpdateItem_UpdateError(t *testing.T) {
	mockKeyRepo := new(MockKeyRepo)
	mockItemRepo := new(MockItemRepo)
	masterKey := []byte("master-key-32-bytes-for-aes256!")
	service := NewItemService(mockKeyRepo, mockItemRepo, masterKey)

	ctx := context.Background()
	userID := uuid.New()
	itemID := uuid.New()
	newTitle := "Updated"

	req := &models.UpdateItemRequest{
		Title: &newTitle,
	}

	mockItemRepo.On("Update", ctx, userID, itemID, req, (*models.EncryptedData)(nil)).
		Return(nil, errors.New("update failed"))

	item, err := service.UpdateItem(ctx, userID, itemID, req)

	assert.Error(t, err)
	assert.Nil(t, item)
	mockItemRepo.AssertExpectations(t)
}
