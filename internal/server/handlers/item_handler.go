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
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ItemService defines the item management service contract.
type ItemService interface {
	CreateItem(ctx context.Context, req *models.CreateItemRequest, userID uuid.UUID) (*models.Item, error)
	ListItems(ctx context.Context, userID uuid.UUID) ([]*models.Item, error)
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (*models.Item, []byte, error)
	UpdateItem(ctx context.Context, userID, itemID uuid.UUID, req *models.UpdateItemRequest) (*models.Item, error)
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error
}

// ItemValidator defines the contract for validating item management requests.
type ItemValidator interface {
	ValidateCreateItemRequest(req *models.CreateItemRequest) error
	ValidateUpdateItemRequest(req *models.UpdateItemRequest) error
	ValidateUUID(id string) (uuid.UUID, error)
}

// ItemHandler handles HTTP requests for item management operations.
type ItemHandler struct {
	itemSvc   ItemService
	validator ItemValidator
	logger    *zap.Logger
}

// NewItemHandler creates a new item handler instance.
func NewItemHandler(itemSvc ItemService, validator ItemValidator, logger *zap.Logger) *ItemHandler {
	return &ItemHandler{
		itemSvc:   itemSvc,
		validator: validator,
		logger:    logger.Named("item_handler"),
	}
}

// itemResponse represents an item response with optional base64-encoded data.
type itemResponse struct {
	Item *models.Item `json:"item"`
	Data string       `json:"data_base64,omitempty"`
}

// CreateItem handles item creation requests.
// Creates a new encrypted item for the authenticated user.
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

	if err := h.validator.ValidateCreateItemRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

// UpdateItem handles item update requests.
// Updates an existing item's metadata and/or encrypted data.
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	if !isJSON(r) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	itemID, err := h.validator.ValidateUUID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req models.UpdateItemRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err = h.validator.ValidateUpdateItemRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.itemSvc.UpdateItem(r.Context(), userID, itemID, &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidItemType) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if errors.Is(err, repositories.ErrItemNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		h.logger.Error("failed to update item", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, itemResponse{Item: item})
}

// ListItems handles requests to list all items for the authenticated user.
// Returns item metadata without encrypted data.
func (h *ItemHandler) ListItems(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	items, err := h.itemSvc.ListItems(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list items", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// GetItem handles requests to retrieve a specific item with its decrypted data.
// Returns both item metadata and base64-encoded decrypted data.
func (h *ItemHandler) GetItem(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	itemID, err := h.validator.ValidateUUID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

// DeleteItem handles requests to delete a specific item.
// Permanently removes the item and its encrypted data from the database.
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	itemID, err := h.validator.ValidateUUID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
