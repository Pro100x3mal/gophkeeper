package services

import (
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
