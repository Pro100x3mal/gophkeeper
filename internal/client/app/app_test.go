package app

import (
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Pro100x3mal/gophkeeper/internal/client/config"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockApiService is a mock implementation of ApiService interface
type MockApiService struct {
	mock.Mock
}

func (m *MockApiService) SetToken(token string) {
	m.Called(token)
}

func (m *MockApiService) Register(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockApiService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockApiService) CreateItem(req *models.CreateItemRequest) (*models.Item, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockApiService) UpdateItem(id uuid.UUID, req *models.UpdateItemRequest) (*models.Item, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockApiService) GetItem(id uuid.UUID) (*models.Item, *string, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*models.Item), args.Get(1).(*string), args.Error(2)
}

func (m *MockApiService) ListItems() ([]*models.Item, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func (m *MockApiService) DeleteItem(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockCacheRepository is a mock implementation of CacheRepository interface
type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) GetToken() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCacheRepository) SetToken(token string) {
	m.Called(token)
}

func (m *MockCacheRepository) ItemsList() map[string]models.Item {
	args := m.Called()
	return args.Get(0).(map[string]models.Item)
}

func (m *MockCacheRepository) Load() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheRepository) Save() error {
	args := m.Called()
	return args.Error(0)
}

// Helper function to create test app
func createTestApp() *App {
	logger, _ := zap.NewDevelopment()
	return &App{
		config: &config.Config{
			ServerAddr:  "http://localhost:8080",
			LogLevel:    "info",
			TLSInsecure: false,
		},
		logger: logger,
		api:    &MockApiService{},
		cache:  &MockCacheRepository{},
	}
}

// Test App.Close method
func TestApp_Close(t *testing.T) {
	app := createTestApp()
	mockCache := new(MockCacheRepository)
	app.cache = mockCache

	mockCache.On("Save").Return(nil)

	err := app.Close()
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestApp_Close_Error(t *testing.T) {
	app := createTestApp()
	mockCache := new(MockCacheRepository)
	app.cache = mockCache

	mockCache.On("Save").Return(errors.New("save failed"))

	err := app.Close()
	assert.Error(t, err)
	assert.Equal(t, "save failed", err.Error())
	mockCache.AssertExpectations(t)
}

// Test parseID helper function
func TestParseID_Valid(t *testing.T) {
	validUUID := uuid.New().String()
	id, err := parseID(validUUID)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)
}

func TestParseID_Invalid(t *testing.T) {
	_, err := parseID("invalid-uuid")
	assert.Error(t, err)
}

func TestParseID_Empty(t *testing.T) {
	_, err := parseID("")
	assert.Error(t, err)
}

// createTestAppWithMocks creates a test app with provided mocks
func createTestAppWithMocks(mockAPI *MockApiService, mockCache *MockCacheRepository) *App {
	logger, _ := zap.NewDevelopment()
	return &App{
		config: &config.Config{
			ServerAddr:   "http://localhost:8080",
			LogLevel:     "info",
			TLSInsecure:  false,
			BuildVersion: "test",
			BuildDate:    "test",
		},
		logger: logger,
		api:    mockAPI,
		cache:  mockCache,
	}
}

// Command tests

func TestCmdRegister(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	token := "test-token-123"
	mockAPI.On("Register", "alice", "secret123").Return(token, nil)
	mockCache.On("SetToken", token).Return()
	mockAPI.On("SetToken", token).Return()

	cmd := app.cmdRegister()
	cmd.SetArgs([]string{"--username", "alice", "--password", "secret123"})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCmdRegister_Error(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	mockAPI.On("Register", "alice", "secret").Return("", assert.AnError)

	cmd := app.cmdRegister()
	cmd.SetArgs([]string{"--username", "alice", "--password", "secret"})

	err := cmd.Execute()
	assert.Error(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdLogin(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	token := "login-token-456"
	mockAPI.On("Login", "alice", "secret123").Return(token, nil)
	mockCache.On("SetToken", token).Return()
	mockAPI.On("SetToken", token).Return()

	cmd := app.cmdLogin()
	cmd.SetArgs([]string{"--username", "alice", "--password", "secret123"})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCmdCreate_WithData(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	testData := "test secret data"
	expectedBase64 := base64.StdEncoding.EncodeToString([]byte(testData))
	itemID := uuid.New()

	mockAPI.On("CreateItem", mock.MatchedBy(func(req *models.CreateItemRequest) bool {
		return req.Type == models.ItemTypeText &&
			req.Title == "Test Note" &&
			req.DataBase64 == expectedBase64
	})).Return(&models.Item{
		ID:    itemID,
		Type:  models.ItemTypeText,
		Title: "Test Note",
	}, nil)

	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdCreate()
	cmd.SetArgs([]string{
		"--type", "text",
		"--title", "Test Note",
		"--data", testData,
	})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdCreate_WithFile(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	testData := []byte("file content")
	_, err = tmpFile.Write(testData)
	require.NoError(t, err)
	tmpFile.Close()

	expectedBase64 := base64.StdEncoding.EncodeToString(testData)
	itemID := uuid.New()

	mockAPI.On("CreateItem", mock.MatchedBy(func(req *models.CreateItemRequest) bool {
		return req.Type == models.ItemTypeBinary &&
			req.Title == "Test File" &&
			req.DataBase64 == expectedBase64
	})).Return(&models.Item{
		ID:    itemID,
		Type:  models.ItemTypeBinary,
		Title: "Test File",
	}, nil)

	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdCreate()
	cmd.SetArgs([]string{
		"--type", "binary",
		"--title", "Test File",
		"--file", tmpFile.Name(),
	})

	err = cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdCreate_WithBothFileAndData(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	cmd := app.cmdCreate()
	cmd.SetArgs([]string{
		"--type", "text",
		"--title", "Test",
		"--file", "/tmp/test.txt",
		"--data", "some data",
	})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use both")
}

func TestCmdCreate_WithMeta(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	itemID := uuid.New()
	metadata := "important note"

	mockAPI.On("CreateItem", mock.MatchedBy(func(req *models.CreateItemRequest) bool {
		return req.Type == models.ItemTypeText &&
			req.Title == "Note" &&
			req.Metadata == metadata
	})).Return(&models.Item{
		ID:    itemID,
		Type:  models.ItemTypeText,
		Title: "Note",
	}, nil)

	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdCreate()
	cmd.SetArgs([]string{
		"--type", "text",
		"--title", "Note",
		"--meta", metadata,
	})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdUpdate_WithData(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	itemID := uuid.New()
	testData := "updated data"
	expectedBase64 := base64.StdEncoding.EncodeToString([]byte(testData))

	mockAPI.On("UpdateItem", itemID, mock.MatchedBy(func(req *models.UpdateItemRequest) bool {
		return req.DataBase64 != nil && *req.DataBase64 == expectedBase64
	})).Return(&models.Item{
		ID:    itemID,
		Type:  models.ItemTypeText,
		Title: "Updated",
	}, nil)

	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdUpdate()
	cmd.SetArgs([]string{
		"--id", itemID.String(),
		"--data", testData,
	})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdUpdate_WithFile(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	itemID := uuid.New()

	tmpFile, err := os.CreateTemp("", "update-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	testData := []byte("updated file content")
	_, err = tmpFile.Write(testData)
	require.NoError(t, err)
	tmpFile.Close()

	expectedBase64 := base64.StdEncoding.EncodeToString(testData)

	mockAPI.On("UpdateItem", itemID, mock.MatchedBy(func(req *models.UpdateItemRequest) bool {
		return req.DataBase64 != nil && *req.DataBase64 == expectedBase64
	})).Return(&models.Item{
		ID:    itemID,
		Type:  models.ItemTypeBinary,
		Title: "Updated",
	}, nil)

	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdUpdate()
	cmd.SetArgs([]string{
		"--id", itemID.String(),
		"--file", tmpFile.Name(),
	})

	err = cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdUpdate_WithBothFileAndData(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	itemID := uuid.New()

	cmd := app.cmdUpdate()
	cmd.SetArgs([]string{
		"--id", itemID.String(),
		"--file", "/tmp/test.txt",
		"--data", "some data",
	})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use both")
}

func TestCmdGet(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	itemID := uuid.New()
	dataBase64 := base64.StdEncoding.EncodeToString([]byte("secret data"))

	mockAPI.On("GetItem", itemID).Return(&models.Item{
		ID:    itemID,
		Type:  models.ItemTypeText,
		Title: "Test Item",
	}, &dataBase64, nil)

	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdGet()
	cmd.SetArgs([]string{"--id", itemID.String()})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdList(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	now := time.Now()
	items := []*models.Item{
		{ID: uuid.New(), Type: models.ItemTypeText, Title: "Item 1", UpdatedAt: now},
		{ID: uuid.New(), Type: models.ItemTypeCredential, Title: "Item 2", UpdatedAt: now.Add(-time.Hour)},
	}

	mockAPI.On("ListItems").Return(items, nil)
	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdList()

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdDelete(t *testing.T) {
	mockAPI := new(MockApiService)
	mockCache := new(MockCacheRepository)
	app := createTestAppWithMocks(mockAPI, mockCache)

	itemID := uuid.New()

	mockAPI.On("DeleteItem", itemID).Return(nil)
	mockCache.On("ItemsList").Return(make(map[string]models.Item))

	cmd := app.cmdDelete()
	cmd.SetArgs([]string{"--id", itemID.String()})

	err := cmd.Execute()
	assert.NoError(t, err)
	mockAPI.AssertExpectations(t)
}

func TestCmdVersion(t *testing.T) {
	app := createTestApp()
	app.config.BuildVersion = "1.0.0"
	app.config.BuildDate = "2024-01-01"

	cmd := app.cmdVersion()

	err := cmd.Execute()
	assert.NoError(t, err)
}
