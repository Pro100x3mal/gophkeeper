package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Pro100x3mal/gophkeeper/internal/server/services"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthSvc defines the authentication service contract.
type AuthSvc interface {
	Register(ctx context.Context, username, password string) (*models.User, string, error)
	Login(ctx context.Context, username, password string) (*models.User, string, error)
}

// AuthValidator defines the contract for validating authentication credentials.
type AuthValidator interface {
	ValidateCredentials(login, password string) error
}

// AuthHandler handles HTTP requests for user authentication.
type AuthHandler struct {
	authSvc   AuthSvc
	validator AuthValidator
	logger    *zap.Logger
}

// NewAuthHandler creates a new authentication handler instance.
func NewAuthHandler(authSvc AuthSvc, validator AuthValidator, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		validator: validator,
		logger:    logger.Named("auth_handler"),
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response with token and user ID.
type AuthResponse struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
}

// Register handles user registration requests.
// Creates a new user account and returns an authentication token.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.validator.ValidateCredentials(req.Username, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, token, err := h.authSvc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrUserAlreadyExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		h.logger.Error("failed to register user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Token:  token,
		UserID: user.ID,
	})
}

// Login handles user login requests.
// Authenticates the user and returns an authentication token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !isJSON(r) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.validator.ValidateCredentials(req.Username, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, token, err := h.authSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.logger.Error("failed to login user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token:  token,
		UserID: user.ID,
	})
}
