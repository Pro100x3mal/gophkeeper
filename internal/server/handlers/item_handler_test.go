package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pro100x3mal/gophkeeper/internal/server/repositories"
	"github.com/Pro100x3mal/gophkeeper/internal/server/services"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockItemService is a mock implementation of ItemService
type MockItemService struct {
	mock.Mock
}

func (m *MockItemService) CreateItem(ctx context.Context, req *models.CreateItemRequest, userID uuid.UUID) (*models.Item, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemService) ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func (m *MockItemService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, []byte, error) {
	args := m.Called(ctx, userID, itemID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	var data []byte
	if args.Get(1) != nil {
		data = args.Get(1).([]byte)
	}
	return args.Get(0).(*models.Item), data, args.Error(2)
}

func (m *MockItemService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, req *models.UpdateItemRequest) (*models.Item, error) {
	args := m.Called(ctx, userID, itemID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	args := m.Called(ctx, userID, itemID)
	return args.Error(0)
}

func TestNewItemHandler(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()

	handler := NewItemHandler(mockService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.itemSvc)
}

func TestItemHandler_CreateItem_Success(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()
	item := &models.Item{
		ID:     itemID,
		UserID: userID,
		Type:   models.ItemTypeText,
		Title:  "Test Item",
	}

	mockService.On("CreateItem", mock.Anything, mock.AnythingOfType("*models.CreateItemRequest"), userID).
		Return(item, nil)

	reqBody := models.CreateItemRequest{
		Type:  models.ItemTypeText,
		Title: "Test Item",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateItem(w, req, userID)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_CreateItem_MissingType(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	reqBody := models.CreateItemRequest{
		Type:  "",
		Title: "Test Item",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateItem(w, req, userID)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_CreateItem_MissingTitle(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	reqBody := models.CreateItemRequest{
		Type:  models.ItemTypeText,
		Title: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateItem(w, req, userID)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_CreateItem_InvalidItemType(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	mockService.On("CreateItem", mock.Anything, mock.AnythingOfType("*models.CreateItemRequest"), userID).
		Return(nil, services.ErrInvalidItemType)

	reqBody := models.CreateItemRequest{
		Type:  "invalid",
		Title: "Test Item",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateItem(w, req, userID)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_ListItems_Success(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	items := []*models.Item{
		{ID: uuid.New(), UserID: userID, Type: models.ItemTypeText, Title: "Item 1"},
		{ID: uuid.New(), UserID: userID, Type: models.ItemTypeCredential, Title: "Item 2"},
	}

	mockService.On("ListItems", mock.Anything, userID).Return(items, nil)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	w := httptest.NewRecorder()

	handler.ListItems(w, req, userID)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_GetItem_Success(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()
	item := &models.Item{ID: itemID, UserID: userID, Type: models.ItemTypeText, Title: "Item"}
	data := []byte("test data")

	mockService.On("GetItem", mock.Anything, userID, itemID).Return(item, data, nil)

	r := chi.NewRouter()
	r.Get("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.GetItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/"+itemID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_GetItem_NotFound(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()

	mockService.On("GetItem", mock.Anything, userID, itemID).
		Return(nil, nil, repositories.ErrItemNotFound)

	r := chi.NewRouter()
	r.Get("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.GetItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/"+itemID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_UpdateItem_Success(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()
	newTitle := "Updated Title"
	item := &models.Item{ID: itemID, UserID: userID, Title: newTitle}

	mockService.On("UpdateItem", mock.Anything, userID, itemID, mock.AnythingOfType("*models.UpdateItemRequest")).
		Return(item, nil)

	reqBody := models.UpdateItemRequest{Title: &newTitle}
	body, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Put("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.UpdateItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodPut, "/items/"+itemID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_UpdateItem_NoFields(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()

	reqBody := models.UpdateItemRequest{}
	body, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Put("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.UpdateItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodPut, "/items/"+itemID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_DeleteItem_Success(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()

	mockService.On("DeleteItem", mock.Anything, userID, itemID).Return(nil)

	r := chi.NewRouter()
	r.Delete("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.DeleteItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodDelete, "/items/"+itemID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_DeleteItem_NotFound(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	itemID := uuid.New()

	mockService.On("DeleteItem", mock.Anything, userID, itemID).
		Return(repositories.ErrItemNotFound)

	r := chi.NewRouter()
	r.Delete("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.DeleteItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodDelete, "/items/"+itemID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_CreateItem_InternalError(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	mockService.On("CreateItem", mock.Anything, mock.AnythingOfType("*models.CreateItemRequest"), userID).
		Return(nil, errors.New("database error"))

	reqBody := models.CreateItemRequest{
		Type:  models.ItemTypeText,
		Title: "Test Item",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateItem(w, req, userID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_ListItems_Error(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	mockService.On("ListItems", mock.Anything, userID).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	w := httptest.NewRecorder()

	handler.ListItems(w, req, userID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestItemHandler_GetItem_InvalidUUID(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()

	r := chi.NewRouter()
	r.Get("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.GetItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/invalid-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_UpdateItem_InvalidUUID(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()
	newTitle := "Updated"
	reqBody := models.UpdateItemRequest{Title: &newTitle}
	body, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Put("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.UpdateItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodPut, "/items/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_DeleteItem_InvalidUUID(t *testing.T) {
	mockService := new(MockItemService)
	logger, _ := zap.NewDevelopment()
	handler := NewItemHandler(mockService, logger)

	userID := uuid.New()

	r := chi.NewRouter()
	r.Delete("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		handler.DeleteItem(w, req, userID)
	})

	req := httptest.NewRequest(http.MethodDelete, "/items/invalid-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
