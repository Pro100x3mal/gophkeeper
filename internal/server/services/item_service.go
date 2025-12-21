package services

import (
	"context"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
)

type KeyRepoInterface interface {
	Save(ctx context.Context, userID uuid.UUID, enc []byte) error
	Load(ctx context.Context, userID uuid.UUID) ([]byte, error)
}

type ItemRepoInterface interface {
	Create(ctx context.Context, item *models.Item) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Item, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Item, error)
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

func (is *ItemService) CreateItem(ctx context.Context, itemReq *models.CreateItemRequest) (*models.Item, error) {
	return nil, nil
}

func (is *ItemService) ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error) {
	return nil, nil
}

func (is *ItemService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error) {
	return nil, nil
}

func (is *ItemService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	return nil
}

func (is *ItemService) LoadOrCreateKey(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	return nil, nil
}
