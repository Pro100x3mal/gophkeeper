package repositories

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	path := "/tmp/test_cache.json"
	cache := NewCache(path)

	assert.NotNil(t, cache)
	assert.Equal(t, path, cache.Path)
	assert.NotNil(t, cache.Items)
	assert.Len(t, cache.Items, 0)
}

func TestCache_GetToken(t *testing.T) {
	cache := NewCache("/tmp/cache.json")
	cache.Token = "test-token"

	token := cache.GetToken()
	assert.Equal(t, "test-token", token)
}

func TestCache_SetToken(t *testing.T) {
	cache := NewCache("/tmp/cache.json")
	cache.SetToken("new-token")

	assert.Equal(t, "new-token", cache.Token)
}

func TestCache_ItemsList(t *testing.T) {
	cache := NewCache("/tmp/cache.json")
	itemID := uuid.New().String()
	item := models.Item{
		ID:    uuid.MustParse(itemID),
		Title: "Test Item",
	}
	cache.Items[itemID] = item

	items := cache.ItemsList()
	assert.Len(t, items, 1)
	assert.Equal(t, item, items[itemID])
}

func TestCache_Save_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	cache := NewCache(cachePath)
	cache.SetToken("test-token")
	itemID := uuid.New().String()
	cache.Items[itemID] = models.Item{
		ID:    uuid.MustParse(itemID),
		Title: "Test Item",
		Type:  models.ItemTypeText,
	}

	err := cache.Save()
	require.NoError(t, err)

	// Verify file exists and contains correct data
	data, err := os.ReadFile(cachePath)
	require.NoError(t, err)

	var loaded Cache
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)
	assert.Equal(t, "test-token", loaded.Token)
	assert.Len(t, loaded.Items, 1)
}

func TestCache_Save_EmptyPath(t *testing.T) {
	cache := NewCache("")
	err := cache.Save()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache path cannot be empty")
}

func TestCache_Load_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Create cache file
	itemID := uuid.New().String()
	cacheData := Cache{
		Token: "saved-token",
		Items: map[string]models.Item{
			itemID: {
				ID:    uuid.MustParse(itemID),
				Title: "Saved Item",
				Type:  models.ItemTypeCredential,
			},
		},
	}
	data, _ := json.MarshalIndent(cacheData, "", "  ")
	err := os.WriteFile(cachePath, data, 0644)
	require.NoError(t, err)

	// Load cache
	cache := NewCache(cachePath)
	err = cache.Load()
	require.NoError(t, err)

	assert.Equal(t, "saved-token", cache.Token)
	assert.Len(t, cache.Items, 1)
	assert.Equal(t, "Saved Item", cache.Items[itemID].Title)
}

func TestCache_Load_EmptyPath(t *testing.T) {
	cache := NewCache("")
	err := cache.Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache path cannot be empty")
}

func TestCache_Load_FileNotExists(t *testing.T) {
	cache := NewCache("/tmp/nonexistent_cache.json")
	err := cache.Load()

	assert.NoError(t, err) // Should return nil if file doesn't exist
}

func TestCache_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "invalid_cache.json")

	// Create invalid JSON file
	err := os.WriteFile(cachePath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	cache := NewCache(cachePath)
	err = cache.Load()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal cache")
}

func TestCache_SaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "roundtrip_cache.json")

	// Save cache
	cache1 := NewCache(cachePath)
	cache1.SetToken("roundtrip-token")
	itemID := uuid.New().String()
	cache1.Items[itemID] = models.Item{
		ID:       uuid.MustParse(itemID),
		Title:    "Roundtrip Item",
		Type:     models.ItemTypeBinary,
		Metadata: `{"key": "value"}`,
	}
	err := cache1.Save()
	require.NoError(t, err)

	// Load cache
	cache2 := NewCache(cachePath)
	err = cache2.Load()
	require.NoError(t, err)

	assert.Equal(t, cache1.Token, cache2.Token)
	assert.Equal(t, len(cache1.Items), len(cache2.Items))
	assert.Equal(t, cache1.Items[itemID].Title, cache2.Items[itemID].Title)
	assert.Equal(t, cache1.Items[itemID].Type, cache2.Items[itemID].Type)
	assert.Equal(t, cache1.Items[itemID].Metadata, cache2.Items[itemID].Metadata)
}

func TestCache_Load_NullItems(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "null_items.json")

	// Create cache with null items
	data := []byte(`{"token": "test", "items": null}`)
	err := os.WriteFile(cachePath, data, 0644)
	require.NoError(t, err)

	cache := NewCache(cachePath)
	err = cache.Load()
	require.NoError(t, err)

	assert.NotNil(t, cache.Items) // Should initialize empty map
	assert.Len(t, cache.Items, 0)
}

func TestCache_MultipleItems(t *testing.T) {
	cache := NewCache("/tmp/test.json")

	for i := 0; i < 5; i++ {
		itemID := uuid.New().String()
		cache.Items[itemID] = models.Item{
			ID:    uuid.MustParse(itemID),
			Title: "Item " + itemID,
			Type:  models.ItemTypeText,
		}
	}

	assert.Len(t, cache.Items, 5)
}
