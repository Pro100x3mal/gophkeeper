package app

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pro100x3mal/gophkeeper/internal/server/config"
	"github.com/Pro100x3mal/gophkeeper/internal/server/handlers"
	"github.com/Pro100x3mal/gophkeeper/internal/server/middleware"
	"github.com/Pro100x3mal/gophkeeper/internal/server/repositories"
	"github.com/Pro100x3mal/gophkeeper/internal/server/services"
	"github.com/Pro100x3mal/gophkeeper/pkg/crypto"
	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"github.com/Pro100x3mal/gophkeeper/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	config       *config.Config
	logger       *zap.Logger
	db           *pgxpool.Pool
	server       *http.Server
	buildVersion string
	buildDate    string
}

func NewApp(cfg *config.Config, buildVersion, buildDate string) (*App, error) {
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Sync()

	appLogger := log.Named("app")

	if cfg.DatabaseDSN == "" {
		return nil, errors.New("database DSN is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("JWT secret is required")
	}
	if cfg.MasterKey == "" {
		return nil, errors.New("master encryption key is required")
	}
	if (cfg.TLSCertFile == "") != (cfg.TLSKeyFile == "") {
		return nil, errors.New("both TLS certificate and key files must be specified or none of them")
	}

	masterKey, err := decodeMasterKey(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	ctx := context.Background()
	db, err := initDB(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	appLogger.Info("Database initialized successfully")

	if err = runMigrations(cfg.DatabaseDSN); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	appLogger.Info("Migrations successfully applied")

	jwtGen := jwt.NewGenerator(cfg.JWTSecret, cfg.JWTExpiration)

	userRepo := repositories.NewUserRepository(db)
	itemRepo := repositories.NewItemRepository(db)
	keyRepo := repositories.NewKeyRepository(db)

	authService := services.NewAuthService(userRepo, jwtGen)
	itemService := services.NewItemService(keyRepo, itemRepo, masterKey)

	infoHandler := handlers.NewInfoHandler(buildVersion, buildDate)
	authHandler := handlers.NewAuthHandler(authService, appLogger)
	itemHandler := handlers.NewItemHandler(itemService, appLogger)

	r := chi.NewRouter()
	r.Use(middleware.Logger(appLogger))
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", infoHandler.HealthCheck)
		r.Get("/version", infoHandler.Version)

		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtGen, appLogger))

			r.Route("/items", func(r chi.Router) {
				r.Post("/", middleware.RequireUser(itemHandler.CreateItem))
				r.Get("/", middleware.RequireUser(itemHandler.ListItems))
				r.Get("/{id}", middleware.RequireUser(itemHandler.GetItem))
				r.Delete("/{id}", middleware.RequireUser(itemHandler.DeleteItem))
			})
		})
	})

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	return &App{
		config:       cfg,
		logger:       appLogger,
		db:           db,
		server:       server,
		buildVersion: buildVersion,
		buildDate:    buildDate,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	serverErrCh := make(chan error, 1)
	go func() {
		if a.config.TLSCertFile != "" && a.config.TLSKeyFile != "" {
			a.logger.Info("Starting HTTPS server", zap.String("address", a.config.ServerAddr))
			serverErrCh <- a.server.ListenAndServeTLS(a.config.TLSCertFile, a.config.TLSKeyFile)
			return
		}

		a.logger.Info("Starting HTTP server", zap.String("address", a.config.ServerAddr))
		serverErrCh <- a.server.ListenAndServe()
	}()

	var serverErr error
	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("Server failed unexpectedly", zap.Error(err))
			serverErr = fmt.Errorf("server failed: %w", err)
		}
	case <-ctx.Done():
		a.logger.Info("Shutting down HTTP server...", zap.String("address", a.config.ServerAddr))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("Failed to shutdown server gracefully", zap.Error(err))
			serverErr = fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	a.logger.Info("Closing database connections...")
	a.db.Close()

	if serverErr == nil {
		a.logger.Info("Server shut down gracefully")
	}

	_ = a.logger.Sync()

	return serverErr
}

func initDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}
	poolConfig.MaxConns = 50
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	if err = pool.Ping(ctxWithTimeout); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

func runMigrations(dsn string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func decodeMasterKey(masterKey string) ([]byte, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	if len(decodedKey) != crypto.KeySize {
		return nil, errors.New("invalid master key length")
	}
	return decodedKey, nil
}
