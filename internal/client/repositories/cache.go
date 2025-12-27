package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Pro100x3mal/gophkeeper/models"
)

type Cache struct {
	Token string                 `json:"token"`
	Items map[string]models.Item `json:"items"`
	Path  string                 `json:"-"`
}

func NewCache(path string) *Cache {
	return &Cache{
		Items: make(map[string]models.Item),
		Path:  path,
	}
}

func (c *Cache) GetToken() string {
	return c.Token
}

func (c *Cache) SetToken(token string) {
	c.Token = token
}

func (c *Cache) ItemsList() map[string]models.Item {
	return c.Items
}

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
