package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/internal/server/repositories"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserRepoInterface interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}
type AuthService struct {
	userRepo UserRepoInterface
	jwtGen   *jwt.Generator
}

func NewAuthService(userRepo UserRepoInterface, jwtGen *jwt.Generator) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtGen:   jwtGen,
	}
}

func (as *AuthService) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err = as.userRepo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
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

func (as *AuthService) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	user, err := as.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
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

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func comparePasswordHash(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
