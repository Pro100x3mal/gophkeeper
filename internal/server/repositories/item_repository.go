package repositories

import (
	"context"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO items (id, user_id, type, title, metadata, data_encrypted, data_key_encrypted) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	if err := r.db.QueryRow(ctx, query,
		item.ID, item.UserID, item.Type, item.Title, item.Metadata, item.DataEncrypted, item.DataKeyEncrypted).
		Scan(&item.CreatedAt, &item.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

func (r *ItemRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Item, error) {
	return nil, nil
}

func (r *ItemRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *ItemRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	return nil, nil
}
