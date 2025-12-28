package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepo is a mock implementation of UserRepo
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestNewAuthService(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)

	service := NewAuthService(mockRepo, jwtGen)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.userRepo)
	assert.Equal(t, jwtGen, service.jwtGen)
}

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "testuser"
	password := "password123"

	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	user, token, err := service.Register(ctx, username, password)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.NotEmpty(t, token)
	assert.NotEqual(t, password, user.PasswordHash)

	// Verify password was hashed correctly
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "existinguser"
	password := "password123"

	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).
		Return(models.ErrUserAlreadyExists)

	user, token, err := service.Register(ctx, username, password)

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "already exists")

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_CreateUserError(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "testuser"
	password := "password123"

	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).
		Return(errors.New("database error"))

	user, token, err := service.Register(ctx, username, password)

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to create user")

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "testuser"
	password := "password123"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockRepo.On("GetUserByUsername", ctx, username).Return(existingUser, nil)

	user, token, err := service.Login(ctx, username, password)

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.NotEmpty(t, token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "nonexistent"
	password := "password123"

	mockRepo.On("GetUserByUsername", ctx, username).Return(nil, models.ErrUserNotFound)

	user, token, err := service.Login(ctx, username, password)

	require.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, user)
	assert.Empty(t, token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "testuser"
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockRepo.On("GetUserByUsername", ctx, username).Return(existingUser, nil)

	user, token, err := service.Login(ctx, username, wrongPassword)

	require.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, user)
	assert.Empty(t, token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_GetUserError(t *testing.T) {
	mockRepo := new(MockUserRepo)
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	service := NewAuthService(mockRepo, jwtGen)

	ctx := context.Background()
	username := "testuser"
	password := "password123"

	mockRepo.On("GetUserByUsername", ctx, username).Return(nil, errors.New("database error"))

	user, token, err := service.Login(ctx, username, password)

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to get user")

	mockRepo.AssertExpectations(t)
}

func TestHashPassword(t *testing.T) {
	password := "testpassword"

	hashed, err := hashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.NotEqual(t, password, string(hashed))

	// Verify hash is valid
	err = bcrypt.CompareHashAndPassword(hashed, []byte(password))
	assert.NoError(t, err)
}

func TestComparePasswordHash_Success(t *testing.T) {
	password := "testpassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	err := comparePasswordHash(string(hashed), password)

	assert.NoError(t, err)
}

func TestComparePasswordHash_Failure(t *testing.T) {
	password := "testpassword"
	wrongPassword := "wrongpassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	err := comparePasswordHash(string(hashed), wrongPassword)

	assert.Error(t, err)
}

func TestErrInvalidCredentials(t *testing.T) {
	assert.Equal(t, "invalid credentials", ErrInvalidCredentials.Error())
}
