package repositories

import (
	"context"

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
	return nil
}

func (r *ItemRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	return nil, nil
}

func (r *ItemRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *ItemRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	return nil, nil
}
