package app

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Pro100x3mal/gophkeeper/internal/client/config"
	"github.com/Pro100x3mal/gophkeeper/internal/client/repositories"
	"github.com/Pro100x3mal/gophkeeper/internal/client/services"
	"github.com/Pro100x3mal/gophkeeper/pkg/logger"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type App struct {
	config *config.Config
	logger *zap.Logger
	api    *services.APIClient
	cache  *repositories.Cache
}

func NewApp(cfg *config.Config) (*App, error) {
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Sync()

	appLogger := log.Named("client")

	cache := repositories.NewCache(cfg.CachePath)
	if err = cache.Load(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %w", err)
	}

	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: cfg.TLSInsecure})
	client.SetTimeout(10 * time.Second)
	api := services.NewAPIClient(client, cfg.ServerAddr)
	api.SetToken(cache.Token)

	return &App{
		config: cfg,
		logger: appLogger,
		api:    api,
		cache:  cache,
	}, nil
}

func (a *App) Close() error {
	return a.cache.Save()
}

func (a *App) Run() error {
	a.logger.Info("Starting client", zap.String("server_addr", a.config.ServerAddr))
	defer a.logger.Info("Stopping client")

	root := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Gophkeeper client",
	}

	root.AddCommand(a.cmdVersion())

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
