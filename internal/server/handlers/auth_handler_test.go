package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pro100x3mal/gophkeeper/internal/server/services"
	"github.com/Pro100x3mal/gophkeeper/internal/server/validators"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockAuthService is a mock implementation of AuthSvc
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.User), args.String(1), args.Error(2)
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.User), args.String(1), args.Error(2)
}

func TestNewAuthHandler(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()

	handler := NewAuthHandler(mockService, validator, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.authSvc)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	userID := uuid.New()
	token := "test-token"
	user := &models.User{ID: userID, Username: "testuser"}

	mockService.On("Register", mock.Anything, "testuser", "password123").
		Return(user, token, nil)

	reqBody := RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response AuthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, token, response.Token)
	assert.Equal(t, userID, response.UserID)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_MissingContentType(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	reqBody := RegisterRequest{Username: "testuser", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	// No Content-Type header set
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_EmptyUsername(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	reqBody := RegisterRequest{Username: "", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_EmptyPassword(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	reqBody := RegisterRequest{Username: "testuser", Password: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Register_UserAlreadyExists(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	mockService.On("Register", mock.Anything, "existinguser", "password123").
		Return(nil, "", models.ErrUserAlreadyExists)

	reqBody := RegisterRequest{Username: "existinguser", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	mockService.On("Register", mock.Anything, "testuser", "password123").
		Return(nil, "", errors.New("database error"))

	reqBody := RegisterRequest{Username: "testuser", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	userID := uuid.New()
	token := "login-token"
	user := &models.User{ID: userID, Username: "testuser"}

	mockService.On("Login", mock.Anything, "testuser", "password123").
		Return(user, token, nil)

	reqBody := LoginRequest{Username: "testuser", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response AuthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, token, response.Token)
	assert.Equal(t, userID, response.UserID)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	mockService.On("Login", mock.Anything, "testuser", "wrongpassword").
		Return(nil, "", services.ErrInvalidCredentials)

	reqBody := LoginRequest{Username: "testuser", Password: "wrongpassword"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_EmptyUsername(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	reqBody := LoginRequest{Username: "", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_EmptyPassword(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	reqBody := LoginRequest{Username: "testuser", Password: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_MissingContentType(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	reqBody := LoginRequest{Username: "testuser", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	// No Content-Type header
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	mockService := new(MockAuthService)
	validator := validators.NewAuthValidator()
	logger, _ := zap.NewDevelopment()
	handler := NewAuthHandler(mockService, validator, logger)

	mockService.On("Login", mock.Anything, "testuser", "password123").
		Return(nil, "", errors.New("database error"))

	reqBody := LoginRequest{Username: "testuser", Password: "password123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
