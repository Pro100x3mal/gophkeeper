package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/Pro100x3mal/gophkeeper/pkg/crypto"
	"github.com/google/uuid"
)

// ErrInvalidItemType is returned when an invalid item type is provided.
var ErrInvalidItemType = fmt.Errorf("invalid item type")

// KeyRepoInterface defines the encryption key repository contract.
type KeyRepoInterface interface {
	Save(ctx context.Context, userID uuid.UUID, enc []byte) error
	Load(ctx context.Context, userID uuid.UUID) ([]byte, bool, error)
}

// ItemRepoInterface defines the item repository contract.
type ItemRepoInterface interface {
	Create(ctx context.Context, item *models.Item, encData *models.EncryptedData) error
	GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, *models.EncryptedData, error)
	DeleteByID(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Item, error)
	Update(
		ctx context.Context,
		userID, itemID uuid.UUID,
		req *models.UpdateItemRequest,
		encData *models.EncryptedData,
	) (*models.Item, error)
}

// ItemService handles encrypted item management with envelope encryption.
// Uses a master key to encrypt per-user keys, which in turn encrypt individual data keys.
type ItemService struct {
	keyRepo   KeyRepoInterface
	itemRepo  ItemRepoInterface
	masterKey []byte
}

// NewItemService creates a new item service instance with the specified master key.
func NewItemService(keyRepo KeyRepoInterface, itemRepo ItemRepoInterface, masterKey []byte) *ItemService {
	return &ItemService{
		keyRepo:   keyRepo,
		itemRepo:  itemRepo,
		masterKey: masterKey,
	}
}

// CreateItem creates a new encrypted item with the provided data.
// Uses envelope encryption: data is encrypted with a data key, which is encrypted with a user key.
func (s *ItemService) CreateItem(ctx context.Context, req *models.CreateItemRequest, userID uuid.UUID) (*models.Item, error) {
	if !isValidType(req.Type) {
		return nil, ErrInvalidItemType
	}

	payload, err := decodeBase64(req.DataBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	item := &models.Item{
		ID:       uuid.New(),
		UserID:   userID,
		Type:     req.Type,
		Title:    req.Title,
		Metadata: req.Metadata,
	}

	var encData *models.EncryptedData
	if len(payload) > 0 {
		userKey, err := s.loadOrCreateKey(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to load or create key: %w", err)
		}

		dataKey, err := crypto.KeyGen()
		if err != nil {
			return nil, fmt.Errorf("failed to generate data key: %w", err)
		}

		dataEncrypted, err := crypto.Encrypt(dataKey, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data: %w", err)
		}

		dataKeyEncrypted, err := crypto.Encrypt(userKey, dataKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data key: %w", err)
		}

		encData = &models.EncryptedData{
			ID:               uuid.New(),
			ItemID:           item.ID,
			DataEncrypted:    dataEncrypted,
			DataKeyEncrypted: dataKeyEncrypted,
		}
	}

	if err = s.itemRepo.Create(ctx, item, encData); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}

// UpdateItem updates an existing item's metadata and/or encrypted data.
// Only provided fields are updated. Uses envelope encryption for new data.
func (s *ItemService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, req *models.UpdateItemRequest) (*models.Item, error) {
	if req.Type != nil && !isValidType(*req.Type) {
		return nil, ErrInvalidItemType
	}

	var encData *models.EncryptedData

	if req.DataBase64 != nil && len(*req.DataBase64) > 0 {
		payload, err := decodeBase64(*req.DataBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 data: %w", err)
		}
		userKey, err := s.loadOrCreateKey(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to load or create key: %w", err)
		}

		dataKey, err := crypto.KeyGen()
		if err != nil {
			return nil, fmt.Errorf("failed to generate data key: %w", err)
		}

		dataEncrypted, err := crypto.Encrypt(dataKey, payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data: %w", err)
		}

		dataKeyEncrypted, err := crypto.Encrypt(userKey, dataKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data key: %w", err)
		}

		encData = &models.EncryptedData{
			ID:               uuid.New(),
			ItemID:           itemID,
			DataEncrypted:    dataEncrypted,
			DataKeyEncrypted: dataKeyEncrypted,
		}
	}

	item, err := s.itemRepo.Update(ctx, userID, itemID, req, encData)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	return item, nil
}

// ListItems retrieves all items for a user without decrypting their data.
func (s *ItemService) ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	items, err := s.itemRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	return items, nil
}

// GetItem retrieves an item and decrypts its data using envelope encryption.
// Returns the item metadata and decrypted data.
func (s *ItemService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, []byte, error) {
	item, encData, err := s.itemRepo.GetByID(ctx, userID, itemID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get item: %w", err)
	}

	var plainData []byte
	if encData != nil && len(encData.DataEncrypted) > 0 {
		userKey, err := s.loadOrCreateKey(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load or create key: %w", err)
		}
		dataKey, err := crypto.Decrypt(userKey, encData.DataKeyEncrypted)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decrypt data key: %w", err)
		}
		plainData, err = crypto.Decrypt(dataKey, encData.DataEncrypted)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decrypt data: %w", err)
		}
	}
	return item, plainData, nil
}

// DeleteItem removes an item and its encrypted data from the database.
func (s *ItemService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	if err := s.itemRepo.DeleteByID(ctx, userID, itemID); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

// loadOrCreateKey retrieves a user's encryption key or generates a new one if it doesn't exist.
// The user key is encrypted with the master key before storage.
func (s *ItemService) loadOrCreateKey(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	keyEncrypted, ok, err := s.keyRepo.Load(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %w", err)
	} else if ok && len(keyEncrypted) > 0 {
		return crypto.Decrypt(s.masterKey, keyEncrypted)
	}

	key, err := crypto.KeyGen()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	enc, err := crypto.Encrypt(s.masterKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}

	if err := s.keyRepo.Save(ctx, userID, enc); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	return key, nil
}

// isValidType checks if an item type is one of the supported types.
func isValidType(t models.ItemType) bool {
	switch t {
	case models.ItemTypeCredential, models.ItemTypeText, models.ItemTypeBinary, models.ItemTypeCard:
		return true
	default:
		return false
	}
}

// decodeBase64 decodes a base64-encoded string.
// Returns nil if the input is empty.
func decodeBase64(data string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}
	payload, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return payload, nil
}
