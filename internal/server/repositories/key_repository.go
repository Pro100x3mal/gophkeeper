package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrKeyNotFound = fmt.Errorf("encryption key not found")

type KeyRepository struct {
	db *pgxpool.Pool
}

func NewKeyRepository(db *pgxpool.Pool) *KeyRepository {
	return &KeyRepository{db: db}
}

func (r *KeyRepository) Save(ctx context.Context, userID uuid.UUID, enc []byte) error {
	query := `
		INSERT INTO encryption_keys (user_id, key_encrypted) 
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET key_encrypted = EXCLUDED.key_encrypted, updated_at = NOW()
	`
	if _, err := r.db.Exec(ctx, query, userID, enc); err != nil {
		return fmt.Errorf("failed to save encryption key: %w", err)
	}
	return nil
}

func (r *KeyRepository) Load(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	query := `
		SELECT key_encrypted 
		FROM encryption_keys 
		WHERE user_id = $1
	`
	var enc []byte
	if err := r.db.QueryRow(ctx, query, userID).Scan(&enc); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to load encryption key: %w", err)
	}
	return enc, nil
}
