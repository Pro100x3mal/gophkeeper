package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := resty.New()
	baseURL := "http://localhost:8080"

	apiClient := NewAPIClient(client, baseURL)

	assert.NotNil(t, apiClient)
	assert.NotNil(t, apiClient.client)
}

func TestAPIClient_SetToken(t *testing.T) {
	client := resty.New()
	apiClient := NewAPIClient(client, "http://localhost:8080")

	// Set token
	apiClient.SetToken("test-token-123")
	assert.Equal(t, "test-token-123", apiClient.client.Token)

	// Clear token
	apiClient.SetToken("")
	assert.Equal(t, "", apiClient.client.Token)
}

func TestAPIClient_Register_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/register", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "testuser", req["username"])
		assert.Equal(t, "testpass", req["password"])

		resp := authResponse{
			Token:  "test-token-123",
			UserID: uuid.New(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	token, err := apiClient.Register("testuser", "testpass")
	assert.NoError(t, err)
	assert.Equal(t, "test-token-123", token)
}

func TestAPIClient_Register_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := authResponse{
			Token:  "",
			UserID: uuid.New(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	_, err := apiClient.Register("testuser", "testpass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty token")
}

func TestAPIClient_Login_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/login", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "testuser", req["username"])
		assert.Equal(t, "testpass", req["password"])

		resp := authResponse{
			Token:  "login-token-456",
			UserID: uuid.New(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	token, err := apiClient.Login("testuser", "testpass")
	assert.NoError(t, err)
	assert.Equal(t, "login-token-456", token)
}

func TestAPIClient_Login_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := authResponse{
			Token:  "",
			UserID: uuid.New(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	_, err := apiClient.Login("testuser", "testpass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty token")
}

func TestAPIClient_CreateItem_Success(t *testing.T) {
	itemID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req models.CreateItemRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, models.ItemTypeText, req.Type)
		assert.Equal(t, "Test Item", req.Title)

		resp := struct {
			Item *models.Item `json:"item"`
		}{
			Item: &models.Item{
				ID:    itemID,
				Type:  req.Type,
				Title: req.Title,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	req := &models.CreateItemRequest{
		Type:       models.ItemTypeText,
		Title:      "Test Item",
		DataBase64: "dGVzdCBkYXRh",
	}

	item, err := apiClient.CreateItem(req)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, "Test Item", item.Title)
}

func TestAPIClient_CreateItem_NilRequest(t *testing.T) {
	client := resty.New()
	apiClient := NewAPIClient(client, "http://localhost:8080")

	_, err := apiClient.CreateItem(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestAPIClient_UpdateItem_Success(t *testing.T) {
	itemID := uuid.New()
	newTitle := "Updated Title"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items/"+itemID.String(), r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		var req models.UpdateItemRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, newTitle, *req.Title)

		resp := struct {
			Item *models.Item `json:"item"`
		}{
			Item: &models.Item{
				ID:    itemID,
				Type:  models.ItemTypeText,
				Title: newTitle,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	req := &models.UpdateItemRequest{
		Title: &newTitle,
	}

	item, err := apiClient.UpdateItem(itemID, req)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, newTitle, item.Title)
}

func TestAPIClient_UpdateItem_NilRequest(t *testing.T) {
	client := resty.New()
	apiClient := NewAPIClient(client, "http://localhost:8080")

	itemID := uuid.New()
	_, err := apiClient.UpdateItem(itemID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestAPIClient_GetItem_Success(t *testing.T) {
	itemID := uuid.New()
	dataBase64 := "dGVzdCBkYXRh"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items/"+itemID.String(), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		resp := struct {
			Item *models.Item `json:"item"`
			Data *string      `json:"data_base64,omitempty"`
		}{
			Item: &models.Item{
				ID:    itemID,
				Type:  models.ItemTypeText,
				Title: "Test Item",
			},
			Data: &dataBase64,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	item, data, err := apiClient.GetItem(itemID)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.NotNil(t, data)
	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, dataBase64, *data)
}

func TestAPIClient_ListItems_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		resp := struct {
			Items []*models.Item `json:"items"`
		}{
			Items: []*models.Item{
				{ID: uuid.New(), Type: models.ItemTypeText, Title: "Item 1"},
				{ID: uuid.New(), Type: models.ItemTypeCredential, Title: "Item 2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	items, err := apiClient.ListItems()
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Item 1", items[0].Title)
	assert.Equal(t, "Item 2", items[1].Title)
}

func TestAPIClient_ListItems_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Items []*models.Item `json:"items"`
		}{
			Items: []*models.Item{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	items, err := apiClient.ListItems()
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestAPIClient_DeleteItem_Success(t *testing.T) {
	itemID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items/"+itemID.String(), r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	err := apiClient.DeleteItem(itemID)
	assert.NoError(t, err)
}

func TestAPIClient_DeleteItem_NotFound(t *testing.T) {
	itemID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("item not found"))
	}))
	defer server.Close()

	client := resty.New()
	apiClient := NewAPIClient(client, server.URL)

	err := apiClient.DeleteItem(itemID)
	assert.Error(t, err)
}

// Test with different item types
func TestAPIClient_CreateItem_DifferentTypes(t *testing.T) {
	tests := []struct {
		name     string
		itemType models.ItemType
	}{
		{"credentials", models.ItemTypeCredential},
		{"text", models.ItemTypeText},
		{"binary", models.ItemTypeBinary},
		{"card", models.ItemTypeCard},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemID := uuid.New()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req models.CreateItemRequest
				json.NewDecoder(r.Body).Decode(&req)

				assert.Equal(t, tt.itemType, req.Type)

				resp := struct {
					Item *models.Item `json:"item"`
				}{
					Item: &models.Item{
						ID:    itemID,
						Type:  req.Type,
						Title: req.Title,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := resty.New()
			apiClient := NewAPIClient(client, server.URL)

			req := &models.CreateItemRequest{
				Type:       tt.itemType,
				Title:      "Test " + tt.name,
				DataBase64: "dGVzdA==",
			}

			item, err := apiClient.CreateItem(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.itemType, item.Type)
		})
	}
}
