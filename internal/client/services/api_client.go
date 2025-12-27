package services

import (
	"fmt"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type APIClient struct {
	client *resty.Client
}

func NewAPIClient(client *resty.Client, baseURL string) *APIClient {
	client.SetBaseURL(baseURL)
	return &APIClient{client: client}
}

type authResponse struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
}

func (c *APIClient) SetToken(token string) {
	if token == "" {
		c.client.SetAuthToken("")
		return
	}
	c.client.SetAuthToken(token)
}

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

func (c *APIClient) DeleteItem(id uuid.UUID) error {
	_, err := c.client.R().
		Delete(fmt.Sprintf("/api/v1/items/%s", id))
	if err != nil {
		return fmt.Errorf("failed to delete item %s: %w", id, err)
	}
	return nil
}
