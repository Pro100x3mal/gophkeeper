// Package repositories provides data access layer for the GophKeeper client.
//
// This package implements local caching functionality for storing authentication
// tokens and item metadata.
package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Pro100x3mal/gophkeeper/models"
)

// Cache manages local storage of authentication tokens and item metadata.
type Cache struct {
	// Token is the authentication token for API requests.
	Token string `json:"token"`
	// Items is a map of item IDs to item metadata.
	Items map[string]models.Item `json:"items"`
	// Path is the file path for cache persistence.
	Path string `json:"-"`
}

// NewCache creates a new cache instance with the specified file path.
func NewCache(path string) *Cache {
	return &Cache{
		Items: make(map[string]models.Item),
		Path:  path,
	}
}

// GetToken retrieves the stored authentication token.
func (c *Cache) GetToken() string {
	return c.Token
}

// SetToken updates the stored authentication token.
func (c *Cache) SetToken(token string) {
	c.Token = token
}

// ItemsList returns the map of cached items.
func (c *Cache) ItemsList() map[string]models.Item {
	return c.Items
}

// Load reads the cache from disk and populates the cache structure.
// Returns nil if the cache file doesn't exist.
func (c *Cache) Load() error {
	if c.Path == "" {
		return errors.New("cache path cannot be empty")
	}

	data, err := os.ReadFile(c.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to load cache: %w", err)
	}

	if err = json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}
	if c.Items == nil {
		c.Items = make(map[string]models.Item)
	}

	return nil
}

// Save writes the cache to disk in JSON format with indentation.
// Creates the cache directory if it doesn't exist.
func (c *Cache) Save() error {
	if c.Path == "" {
		return errors.New("cache path cannot be empty")
	}

	if err := os.MkdirAll(filepath.Dir(c.Path), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err = os.WriteFile(c.Path, data, 0644); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return nil
}
