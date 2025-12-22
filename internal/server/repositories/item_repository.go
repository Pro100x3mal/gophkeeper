package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrItemNotFound = fmt.Errorf("item not found")

type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(ctx context.Context, item *models.Item, encData *models.EncryptedData) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	itemQuery := `
		INSERT INTO items (id, user_id, type, title, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	if err = tx.QueryRow(ctx, itemQuery,
		item.ID, item.UserID, item.Type, item.Title, item.Metadata).
		Scan(&item.CreatedAt, &item.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	if encData != nil {
		encData.ItemID = item.ID

		dataQuery := `
			INSERT INTO encrypted_data (id, item_id, data_encrypted, data_key_encrypted)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`

		if err = tx.QueryRow(ctx, dataQuery,
			encData.ID, encData.ItemID, encData.DataEncrypted, encData.DataKeyEncrypted).
			Scan(&encData.ID); err != nil {
			return fmt.Errorf("failed to create encrypted-data: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ItemRepository) GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, *models.EncryptedData, error) {
	itemQuery := `
		SELECT id, user_id, type, title, metadata, created_at, updated_at
		FROM items
		WHERE id = $1 AND user_id = $2
	`
	var item models.Item
	if err := r.db.QueryRow(ctx, itemQuery, itemID, userID).
		Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Metadata, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrItemNotFound
		}
		return nil, nil, fmt.Errorf("failed to get item: %w", err)
	}

	dataQuery := `
		SELECT id, item_id, data_encrypted, data_key_encrypted
		FROM encrypted_data
		WHERE item_id = $1
	`
	var data models.EncryptedData
	if err := r.db.QueryRow(ctx, dataQuery, itemID).
		Scan(&data.ID, &data.ItemID, &data.DataEncrypted, &data.DataKeyEncrypted); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, fmt.Errorf("failed to get encrypted-data: %w", err)
		}
		return &item, nil, nil
	}

	return &item, &data, nil
}

func (r *ItemRepository) DeleteByID(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	query := `DELETE FROM items WHERE id = $1 AND user_id = $2`
	t, err := r.db.Exec(ctx, query, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	if t.RowsAffected() == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *ItemRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	query := `
		SELECT id, user_id, type, title, metadata, created_at, updated_at
		FROM items
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		var item models.Item
		if err = rows.Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Metadata, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over items: %w", err)
	}

	return items, nil
}
