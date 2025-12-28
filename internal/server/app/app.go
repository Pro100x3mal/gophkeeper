// Package app provides the main application setup and lifecycle management for the GophKeeper server.
//
// This package initializes all components including database connections, migrations,
// HTTP routes, and graceful shutdown handling.
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
	"github.com/Pro100x3mal/gophkeeper/internal/server/validators"
	"github.com/Pro100x3mal/gophkeeper/pkg/crypto"
	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"github.com/Pro100x3mal/gophkeeper/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// App represents the main application with all its dependencies.
type App struct {
	config       *config.Config
	logger       *zap.Logger
	db           *pgxpool.Pool
	server       *http.Server
	buildVersion string
	buildDate    string
}

// NewApp creates and initializes a new application instance.
// It sets up the database, runs migrations, initializes services and handlers,
// and configures the HTTP router.
//
// Parameters:
//   - cfg: application configuration
//   - buildVersion: version string for the build
//   - buildDate: build timestamp
//
// Returns the initialized App instance or an error if initialization fails.
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

	authValidator := validators.NewAuthValidator()
	itemValidator := validators.NewItemValidator()

	infoHandler := handlers.NewInfoHandler(buildVersion, buildDate)
	authHandler := handlers.NewAuthHandler(authService, authValidator, appLogger)
	itemHandler := handlers.NewItemHandler(itemService, itemValidator, appLogger)

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /api/v1/health", infoHandler.HealthCheck)
	mux.HandleFunc("GET /api/v1/version", infoHandler.Version)
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	// Protected endpoints
	authMiddleware := middleware.Auth(jwtGen, appLogger)
	mux.Handle("POST /api/v1/items/", authMiddleware(middleware.RequireUser(itemHandler.CreateItem)))
	mux.Handle("GET /api/v1/items/", authMiddleware(middleware.RequireUser(itemHandler.ListItems)))
	mux.Handle("GET /api/v1/items/{id}", authMiddleware(middleware.RequireUser(itemHandler.GetItem)))
	mux.Handle("PUT /api/v1/items/{id}", authMiddleware(middleware.RequireUser(itemHandler.UpdateItem)))
	mux.Handle("DELETE /api/v1/items/{id}", authMiddleware(middleware.RequireUser(itemHandler.DeleteItem)))

	// Wrap with Logger middleware
	handler := middleware.Logger(appLogger)(mux)

	server := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      handler,
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

// Run starts the HTTP server and handles graceful shutdown on system signals.
// It listens for SIGINT, SIGTERM, and SIGQUIT signals and performs cleanup
// when a signal is received or the server encounters an error.
//
// Returns an error if the server fails to start or encounters a fatal error.
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

// initDB initializes a PostgreSQL connection pool with configured parameters.
// Sets up connection pooling with health checks and connection lifecycle limits.
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

// runMigrations applies database migrations from embedded SQL files.
// Uses golang-migrate to manage schema changes.
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

// decodeMasterKey decodes and validates a base64-encoded master encryption key.
// Ensures the key has the correct length for AES-256 encryption.
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
