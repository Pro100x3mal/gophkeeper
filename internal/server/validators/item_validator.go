package validators

import (
	"errors"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
)

var (
	// ErrEmptyType is returned when item type field is empty during creation.
	ErrEmptyType = errors.New("item type cannot be empty")

	// ErrEmptyTitle is returned when item title field is empty during creation.
	ErrEmptyTitle = errors.New("item title cannot be empty")

	// ErrInvalidUUID is returned when provided UUID string cannot be parsed.
	ErrInvalidUUID = errors.New("invalid UUID format")

	// ErrNoFieldsToUpdate is returned when update request contains no fields to update.
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)

// ItemValidator handles validation of item management requests.
type ItemValidator struct{}

// NewItemValidator creates a new instance of ItemValidator.
func NewItemValidator() *ItemValidator {
	return &ItemValidator{}
}

// ValidateCreateItemRequest validates item creation request.
// Ensures that type and title fields are non-empty.
// Returns ErrEmptyType or ErrEmptyTitle if validation fails.
func (v *ItemValidator) ValidateCreateItemRequest(req *models.CreateItemRequest) error {
	if req.Type == "" {
		return ErrEmptyType
	}

	if req.Title == "" {
		return ErrEmptyTitle
	}

	return nil
}

// ValidateUpdateItemRequest validates item update request.
// Ensures that at least one field is provided for update.
// Returns ErrNoFieldsToUpdate if no fields are provided.
func (v *ItemValidator) ValidateUpdateItemRequest(req *models.UpdateItemRequest) error {
	if req.Type == nil && req.Title == nil && req.Metadata == nil && req.DataBase64 == nil {
		return ErrNoFieldsToUpdate
	}
	return nil
}

// ValidateUUID validates and parses UUID string.
// Returns parsed UUID or ErrInvalidUUID if parsing fails.
func (v *ItemValidator) ValidateUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}
	return parsed, nil
}
