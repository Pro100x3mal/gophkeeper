// Package config provides configuration management for the GophKeeper server.
//
// Configuration is loaded from environment variables (via .env file) and command-line flags.
// Command-line flags take precedence over environment variables.
package config

import (
	"flag"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the server configuration parameters.
type Config struct {
	// LogLevel specifies the logging verbosity (debug, info, warn, error).
	LogLevel string
	// ServerAddr is the address to bind the HTTP/HTTPS server to.
	ServerAddr string
	// DatabaseDSN is the PostgreSQL connection string.
	DatabaseDSN string
	// JWTSecret is the secret key used for signing JWT tokens.
	JWTSecret string
	// JWTExpiration is the duration for which JWT tokens remain valid.
	JWTExpiration time.Duration
	// TLSCertFile is the path to the TLS certificate file (optional).
	TLSCertFile string
	// TLSKeyFile is the path to the TLS private key file (optional).
	TLSKeyFile string
	// MasterKey is the base64-encoded master encryption key.
	MasterKey string
}

// Load reads configuration from environment variables and command-line flags.
// It first attempts to load a .env file, then parses command-line flags.
// Command-line flags take precedence over environment variables.
//
// Returns the loaded configuration or an error if loading fails.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}

	flag.StringVar(&cfg.LogLevel, "l", getEnv("LOG_LEVEL", "info"), "Log level")
	flag.StringVar(&cfg.ServerAddr, "a", getEnv("SERVER_ADDR", "localhost:8080"), "Server address")
	flag.StringVar(&cfg.DatabaseDSN, "d", getEnv("DATABASE_DSN", ""), "Database DSN")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", getEnv("JWT_SECRET", ""), "JWT secret key")
	flag.StringVar(&cfg.TLSCertFile, "tls-cert", getEnv("TLS_CERT_FILE", ""), "TLS certificate file")
	flag.StringVar(&cfg.TLSKeyFile, "tls-key", getEnv("TLS_KEY_FILE", ""), "TLS key file")
	flag.StringVar(&cfg.MasterKey, "master-key", getEnv("MASTER_KEY", ""), "Master encryption key in base64 format")
	flag.DurationVar(&cfg.JWTExpiration, "jwt-exp", getEnvDuration("JWT_EXPIRATION", 24*time.Hour), "JWT expiration time")

	flag.Parse()

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value if not set.
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// getEnvDuration retrieves a duration from an environment variable or returns a default value.
// The environment variable should contain a valid duration string (e.g., "24h", "30m").
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}
