package handlers

import (
	"context"
	"net/http"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ItemService interface {
	CreateItem(ctx context.Context, item *models.Item) (*models.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error)
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, error)
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error
}

type ItemHandler struct {
	itemSvc ItemService
	logger  *zap.Logger
}

func NewItemHandler(itemSvc ItemService, logger *zap.Logger) *ItemHandler {
	return &ItemHandler{
		itemSvc: itemSvc,
		logger:  logger.Named("item_handler"),
	}
}

func (ih *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {

}

func (ih *ItemHandler) List(w http.ResponseWriter, r *http.Request) {

}

func (ih *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {

}

func (ih *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {

}
