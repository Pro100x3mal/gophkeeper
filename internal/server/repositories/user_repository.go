// Package repositories provides data access layer for the GophKeeper server.
//
// This package implements repository pattern for database operations including
// user management, item storage, and encryption key management.
package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrUserAlreadyExists is returned when attempting to create a user with an existing username.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrUserNotFound is returned when a user cannot be found by ID or username.
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository handles database operations for user entities.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository instance.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the database.
// Returns ErrUserAlreadyExists if the username is already taken.
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`

	if err := r.db.QueryRow(ctx, query, user.ID, user.Username, user.PasswordHash).
		Scan(&user.CreatedAt, &user.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// isUniqueViolation checks if an error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// GetUserByID retrieves a user by their ID.
// Returns ErrUserNotFound if the user does not exist.
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user models.User
	if err := r.db.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username.
// Returns ErrUserNotFound if the user does not exist.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var user models.User
	if err := r.db.QueryRow(ctx, query, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}
