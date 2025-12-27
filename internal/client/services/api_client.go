// Package services provides business logic layer for the GophKeeper client.
//
// This package implements API client functionality for communicating with
// the GophKeeper server over HTTP/HTTPS.
package services

import (
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// APIClient handles HTTP communication with the GophKeeper server.
type APIClient struct {
	client *resty.Client
}

// NewAPIClient creates a new API client instance with the specified base URL.
func NewAPIClient(client *resty.Client, baseURL string) *APIClient {
	client.SetBaseURL(baseURL)
	return &APIClient{client: client}
}

// authResponse represents the authentication response from the server.
type authResponse struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
}

// SetToken sets the authentication token for API requests.
// Clears the token if an empty string is provided.
func (c *APIClient) SetToken(token string) {
	if token == "" {
		c.client.SetAuthToken("")
		return
	}
	c.client.SetAuthToken(token)
}

// Register creates a new user account on the server.
// Returns the authentication token for the new user.
func (c *APIClient) Register(username, password string) (string, error) {
	var resp authResponse
	_, err := c.client.R().
		SetBody(map[string]string{"username": username, "password": password}).
		SetResult(&resp).
		Post("/api/v1/register")
	if err != nil {
		return "", fmt.Errorf("failed to register user %q: %w", username, err)
	}
	if resp.Token == "" {
		return "", fmt.Errorf("empty token in response")
	}
	return resp.Token, nil
}

// Login authenticates an existing user on the server.
// Returns the authentication token for the user.
func (c *APIClient) Login(username, password string) (string, error) {
	var resp authResponse
	_, err := c.client.R().
		SetBody(map[string]string{"username": username, "password": password}).
		SetResult(&resp).
		Post("/api/v1/login")
	if err != nil {
		return "", fmt.Errorf("failed to login user %q: %w", username, err)
	}
	if resp.Token == "" {
		return "", fmt.Errorf("empty token in response")
	}
	return resp.Token, nil
}

// CreateItem creates a new item on the server.
// Returns the created item metadata.
func (c *APIClient) CreateItem(req *models.CreateItemRequest) (*models.Item, error) {
	if req == nil {
		return nil, fmt.Errorf("create item request cannot be nil")
	}
	var resp struct {
		Item *models.Item `json:"item"`
	}
	_, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Post("/api/v1/items")
	if err != nil {
		return nil, fmt.Errorf("failed to create item %q: %w", req.Title, err)
	}
	return resp.Item, nil
}

// UpdateItem updates an existing item on the server.
// Returns the updated item metadata.
func (c *APIClient) UpdateItem(id uuid.UUID, req *models.UpdateItemRequest) (*models.Item, error) {
	if req == nil {
		return nil, fmt.Errorf("update item request cannot be nil")
	}
	var resp struct {
		Item *models.Item `json:"item"`
	}
	_, err := c.client.R().
		SetBody(req).
		SetResult(&resp).
		Put(fmt.Sprintf("/api/v1/items/%s", id))
	if err != nil {
		return nil, fmt.Errorf("failed to update item %s: %w", id, err)
	}
	return resp.Item, nil
}

// GetItem retrieves an item and its data from the server.
// Returns the item metadata and base64-encoded data.
func (c *APIClient) GetItem(id uuid.UUID) (*models.Item, *string, error) {
	var resp struct {
		Item *models.Item `json:"item"`
		Data *string      `json:"data_base64,omitempty"`
	}
	_, err := c.client.R().
		SetResult(&resp).
		Get(fmt.Sprintf("/api/v1/items/%s", id))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get item %s: %w", id, err)
	}
	return resp.Item, resp.Data, nil
}

// ListItems retrieves all items for the authenticated user from the server.
// Returns a list of item metadata.
func (c *APIClient) ListItems() ([]*models.Item, error) {
	var resp struct {
		Items []*models.Item `json:"items"`
	}
	_, err := c.client.R().
		SetResult(&resp).
		Get("/api/v1/items")
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	return resp.Items, nil
}

// DeleteItem removes an item from the server.
func (c *APIClient) DeleteItem(id uuid.UUID) error {
	_, err := c.client.R().
		Delete(fmt.Sprintf("/api/v1/items/%s", id))
	if err != nil {
		return fmt.Errorf("failed to delete item %s: %w", id, err)
	}
	return nil
}
