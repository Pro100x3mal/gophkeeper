// Package services provides business logic layer for the GophKeeper server.
//
// This package implements services for authentication and item management,
// handling encryption, key management, and user authentication workflows.
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when username or password is incorrect.
var ErrInvalidCredentials = errors.New("invalid credentials")

// UserRepo defines the user repository contract.
type UserRepo interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

// AuthService handles user authentication and registration operations.
type AuthService struct {
	userRepo UserRepo
	jwtGen   *jwt.Generator
}

// NewAuthService creates a new authentication service instance.
func NewAuthService(userRepo UserRepo, jwtGen *jwt.Generator) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtGen:   jwtGen,
	}
}

// Register creates a new user account with hashed password and returns a JWT token.
func (as *AuthService) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err = as.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, models.ErrUserAlreadyExists) {
			return nil, "", fmt.Errorf("user with username %s already exists: %w", username, err)
		}
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err := as.jwtGen.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return user, token, nil
}

// Login authenticates a user by username and password, returning a JWT token.
// Returns ErrInvalidCredentials if username or password is incorrect.
func (as *AuthService) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	user, err := as.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	if err = comparePasswordHash(user.PasswordHash, password); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := as.jwtGen.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return user, token, nil
}

// hashPassword hashes a password using bcrypt with default cost.
func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// comparePasswordHash verifies a password against its bcrypt hash.
func comparePasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
