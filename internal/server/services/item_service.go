package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/Pro100x3mal/gophkeeper/pkg/crypto"
	"github.com/google/uuid"
)

var ErrInvalidItemType = fmt.Errorf("invalid item type")

type KeyRepoInterface interface {
	Save(ctx context.Context, userID uuid.UUID, enc []byte) error
	Load(ctx context.Context, userID uuid.UUID) ([]byte, bool, error)
}

type ItemRepoInterface interface {
	Create(ctx context.Context, item *models.Item, encData *models.EncryptedData) error
	GetByID(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, *models.EncryptedData, error)
	DeleteByID(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Item, error)
}

type ItemService struct {
	keyRepo   KeyRepoInterface
	itemRepo  ItemRepoInterface
	masterKey []byte
}

func NewItemService(keyRepo KeyRepoInterface, itemRepo ItemRepoInterface, masterKey []byte) *ItemService {
	return &ItemService{
		keyRepo:   keyRepo,
		itemRepo:  itemRepo,
		masterKey: masterKey,
	}
}

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

func (s *ItemService) ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	items, err := s.itemRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	return items, nil
}

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

func (s *ItemService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	if err := s.itemRepo.DeleteByID(ctx, userID, itemID); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

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

func isValidType(t models.ItemType) bool {
	switch t {
	case models.ItemTypeCredential, models.ItemTypeText, models.ItemTypeBinary, models.ItemTypeCard:
		return true
	default:
		return false
	}
}

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
