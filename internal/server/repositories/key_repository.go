package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// KeyRepository handles database operations for user encryption keys.
type KeyRepository struct {
	db *pgxpool.Pool
}

// NewKeyRepository creates a new key repository instance.
func NewKeyRepository(db *pgxpool.Pool) *KeyRepository {
	return &KeyRepository{db: db}
}

// Save stores or updates an encrypted user key in the database.
// Uses upsert to create or update the key for the specified user.
func (r *KeyRepository) Save(ctx context.Context, userID uuid.UUID, enc []byte) error {
	query := `
		INSERT INTO encryption_keys (user_id, key_encrypted) 
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET key_encrypted = EXCLUDED.key_encrypted
	`
	if _, err := r.db.Exec(ctx, query, userID, enc); err != nil {
		return fmt.Errorf("failed to save encryption key: %w", err)
	}
	return nil
}

// Load retrieves an encrypted user key from the database.
// Returns the encrypted key, true if found, and any error.
// Returns nil, false, nil if the key doesn't exist.
func (r *KeyRepository) Load(ctx context.Context, userID uuid.UUID) ([]byte, bool, error) {
	query := `
		SELECT key_encrypted 
		FROM encryption_keys 
		WHERE user_id = $1
	`
	var enc []byte
	if err := r.db.QueryRow(ctx, query, userID).Scan(&enc); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to load encryption key: %w", err)
	}
	return enc, true, nil
}
