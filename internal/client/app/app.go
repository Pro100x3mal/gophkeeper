// Package app provides the main application logic for the GophKeeper client.
//
// This package implements a CLI interface for interacting with the GophKeeper server,
// including authentication, item management, and local caching functionality.
package app

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Pro100x3mal/gophkeeper/internal/client/config"
	"github.com/Pro100x3mal/gophkeeper/internal/client/repositories"
	"github.com/Pro100x3mal/gophkeeper/internal/client/services"
	"github.com/Pro100x3mal/gophkeeper/models"
	"github.com/Pro100x3mal/gophkeeper/pkg/logger"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// CacheRepository defines the local cache repository contract.
type CacheRepository interface {
	GetToken() string
	SetToken(token string)
	ItemsList() map[string]models.Item
	Load() error
	Save() error
}

// ApiService defines the API client service contract.
type ApiService interface {
	SetToken(token string)
	Register(username, password string) (string, error)
	Login(username, password string) (string, error)
	CreateItem(req *models.CreateItemRequest) (*models.Item, error)
	UpdateItem(id uuid.UUID, req *models.UpdateItemRequest) (*models.Item, error)
	GetItem(id uuid.UUID) (*models.Item, *string, error)
	ListItems() ([]*models.Item, error)
	DeleteItem(id uuid.UUID) error
}

// App represents the main client application with its dependencies.
type App struct {
	config *config.Config
	logger *zap.Logger
	api    ApiService
	cache  CacheRepository
}

// NewApp creates and initializes a new client application instance.
// Loads the cache, configures the HTTP client, and sets up the API service.
func NewApp(cfg *config.Config) (*App, error) {
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	cache := repositories.NewCache(cfg.CachePath)
	if err = cache.Load(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %w", err)
	}

	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: cfg.TLSInsecure})
	client.SetTimeout(10 * time.Second)
	api := services.NewAPIClient(client, cfg.ServerAddr)
	api.SetToken(cache.GetToken())

	return &App{
		config: cfg,
		logger: log.Named("client"),
		api:    api,
		cache:  cache,
	}, nil
}

// Close performs cleanup operations before application shutdown.
// Saves the cache and syncs the logger.
func (a *App) Close() error {
	defer a.logger.Sync()
	return a.cache.Save()
}

// Run starts the CLI application and processes commands.
// Initializes the command tree and executes the root command.
func (a *App) Run() error {
	a.logger.Info("Starting client", zap.String("server_addr", a.config.ServerAddr))
	defer a.logger.Info("Stopping client")

	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Gophkeeper client",
	}

	root.AddCommand(a.cmdVersion())
	root.AddCommand(a.cmdRegister())
	root.AddCommand(a.cmdLogin())

	return root.Execute()
}

func (a *App) cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show build info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s, Build: %s\n", a.config.BuildVersion, a.config.BuildDate)
		},
	}
}

func (a *App) cmdRegister() *cobra.Command {
	var username, password string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := a.api.Register(username, password)
			if err != nil {
				return fmt.Errorf("failed to register user: %w", err)
			}
			a.cache.SetToken(token)
			a.api.SetToken(token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func (a *App) cmdLogin() *cobra.Command {
	var username, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := a.api.Login(username, password)
			if err != nil {
				return fmt.Errorf("failed to login: %w", err)
			}
			a.cache.SetToken(token)
			a.api.SetToken(token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func (a *App) cmdCreate() *cobra.Command {
	var typ, title, meta, filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create new item",
		RunE: func(cmd *cobra.Command, args []string) error {
			var dataBase64 string
			if filePath != "" {
				rawData, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				dataBase64 = base64.StdEncoding.EncodeToString(rawData)
			}

			if typ == "" {
				return errors.New("type is required")
			}
			switch models.ItemType(typ) {
			case models.ItemTypeCredential, models.ItemTypeText, models.ItemTypeBinary, models.ItemTypeCard:
			default:
				return fmt.Errorf("unsupported type: %s", typ)
			}

			req := &models.CreateItemRequest{
				Type:       models.ItemType(typ),
				Title:      title,
				Metadata:   meta,
				DataBase64: dataBase64,
			}
			item, err := a.api.CreateItem(req)
			if err != nil {
				return fmt.Errorf("failed to create item: %w", err)
			}
			fmt.Printf("Item created: %s\n", item.ID.String())
			a.cache.ItemsList()[item.ID.String()] = *item
			return nil
		},
	}

	cmd.Flags().StringVar(&typ, "type", "", "Item type (credential|text|binary|card)")
	cmd.Flags().StringVar(&title, "title", "", "Item title")
	cmd.Flags().StringVar(&meta, "meta", "", "Item metadata (plain text)")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to file with item data")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func (a *App) cmdUpdate() *cobra.Command {
	var rawID, typ, title, meta, filePath string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update existing item",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(rawID)
			if err != nil {
				return fmt.Errorf("failed to parse item ID: %w", err)
			}

			req := &models.UpdateItemRequest{}

			if typ != "" {
				switch models.ItemType(typ) {
				case models.ItemTypeCredential, models.ItemTypeText, models.ItemTypeBinary, models.ItemTypeCard:
					t := models.ItemType(typ)
					req.Type = &t
				default:
					return fmt.Errorf("unsupported type: %s", typ)
				}
			}

			if title != "" {
				req.Title = &title
			}

			if meta != "" {
				req.Metadata = &meta
			}

			if filePath != "" {
				rawData, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				dataBase64 := base64.StdEncoding.EncodeToString(rawData)
				req.DataBase64 = &dataBase64
			}

			if req.Type == nil && req.Title == nil && req.Metadata == nil && req.DataBase64 == nil {
				return errors.New("nothing to update")
			}

			item, err := a.api.UpdateItem(id, req)
			if err != nil {
				return fmt.Errorf("failed to update item: %w", err)
			}
			fmt.Printf("Item updated: %s\n", item.ID)
			a.cache.ItemsList()[item.ID.String()] = *item
			return nil
		},
	}

	cmd.Flags().StringVar(&rawID, "id", "", "Item ID")
	cmd.Flags().StringVar(&typ, "type", "", "Item type (credential|text|binary|card)")
	cmd.Flags().StringVar(&title, "title", "", "Item title")
	cmd.Flags().StringVar(&meta, "meta", "", "Item metadata (plain text)")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to file with item data")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (a *App) cmdGet() *cobra.Command {
	var rawID, outPath string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get item by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(rawID)
			if err != nil {
				return fmt.Errorf("failed to parse item ID: %w", err)
			}

			item, data, err := a.api.GetItem(id)
			if err != nil {
				a.logger.Warn("Failed to get item from server, using cache", zap.Error(err))
				if cachedItem, ok := a.cache.ItemsList()[id.String()]; ok {
					fmt.Printf("%+v\nData: <not cached>\n", cachedItem)
					return nil
				}
				return fmt.Errorf("failed to get item: %w", err)
			}
			fmt.Printf("%+v\n", *item)

			if data != nil && *data != "" {
				rawData, err := base64.StdEncoding.DecodeString(*data)
				if err != nil {
					return fmt.Errorf("failed to decode base64 data: %w", err)
				}

				if outPath != "" {
					if err = os.WriteFile(outPath, rawData, 0644); err != nil {
						return fmt.Errorf("failed to write data to file: %w", err)
					}
					fmt.Printf("Data saved to file: %s\n", outPath)
				} else {
					fmt.Printf("Data:\n%s\n", rawData)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&rawID, "id", "", "Item ID")
	cmd.Flags().StringVar(&outPath, "out", "", "Path to save item data")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (a *App) cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all items",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := a.api.ListItems()
			if err != nil {
				a.logger.Warn("Failed to get items from server, using cache", zap.Error(err))
				cachedItems := make([]models.Item, 0, len(a.cache.ItemsList()))
				for _, item := range a.cache.ItemsList() {
					cachedItems = append(cachedItems, item)
				}
				sort.Slice(cachedItems, func(i, j int) bool {
					return cachedItems[i].UpdatedAt.After(cachedItems[j].UpdatedAt)
				})
				for _, cached := range cachedItems {
					fmt.Printf("%s\t%s\t%s\n", cached.ID, cached.Type, cached.Title)
				}
				return nil
			}

			sort.Slice(items, func(i, j int) bool {
				if items[i] == nil || items[j] == nil {
					return false
				}
				return items[i].UpdatedAt.After(items[j].UpdatedAt)
			})

			for _, item := range items {
				fmt.Printf("%s\t%s\t%s\n", item.ID, item.Type, item.Title)
			}
			return nil
		},
	}
}

func (a *App) cmdDelete() *cobra.Command {
	var rawID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete item by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(rawID)
			if err != nil {
				return fmt.Errorf("invalid ID format: %w", err)
			}
			if err = a.api.DeleteItem(id); err != nil {
				return fmt.Errorf("failed to delete item: %w", err)
			}
			delete(a.cache.ItemsList(), id.String())
			return nil
		},
	}

	cmd.Flags().StringVar(&rawID, "id", "", "Item ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func parseID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
	}
	return id, nil
}
