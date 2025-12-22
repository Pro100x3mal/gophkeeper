package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Pro100x3mal/gophkeeper/internal/server/repositories"
	"github.com/Pro100x3mal/gophkeeper/internal/server/services"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ItemService interface {
	CreateItem(ctx context.Context, req *models.CreateItemRequest, userID uuid.UUID) (*models.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error)
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, []byte, error)
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

type itemResponse struct {
	Item *models.Item `json:"item"`
	Data string       `json:"data_base64,omitempty"`
}

type listItemsResponse struct {
	Items []*models.Item `json:"items"`
}

func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	if !isJSON(r) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req models.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if req.Type == "" || req.Title == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := h.itemSvc.CreateItem(r.Context(), &req, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvalidItemType) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		h.logger.Error("failed to create item", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, itemResponse{Item: item})
}

func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	items, err := h.itemSvc.ListItems(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list items", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, data, err := h.itemSvc.GetItem(r.Context(), userID, itemID)
	if err != nil {
		if errors.Is(err, repositories.ErrItemNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get item", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := itemResponse{Item: item}
	if len(data) > 0 {
		resp.Data = base64.StdEncoding.EncodeToString(data)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = h.itemSvc.DeleteItem(r.Context(), userID, itemID); err != nil {
		if errors.Is(err, repositories.ErrItemNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete item", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
