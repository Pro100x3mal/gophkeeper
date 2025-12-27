// Package config provides configuration management for the GophKeeper client.
//
// Configuration is loaded from environment variables (via .env file) and command-line flags.
// Command-line flags take precedence over environment variables.
package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the client configuration parameters.
type Config struct {
	// ServerAddr is the address of the GophKeeper server.
	ServerAddr string
	// LogLevel specifies the logging verbosity.
	LogLevel string
	// TLSInsecure disables TLS certificate verification when true.
	TLSInsecure bool
	// CachePath is the path to the local cache file.
	CachePath string
	// TokenPath is the path to the authentication token file.
	TokenPath string
	// BuildVersion contains the version of the application.
	BuildVersion string
	// BuildDate contains the build timestamp.
	BuildDate string
}

// Load reads configuration from environment variables and command-line flags.
// It first attempts to load a .env file, then parses command-line flags.
// Returns the loaded configuration or an error if loading fails.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}

	execPath, err := os.Executable()
	if err != nil {
		execPath = os.Args[0]
	}
	execDir := filepath.Dir(execPath)

	defaultCache := filepath.Join(execDir, "cache.json")
	defaultToken := filepath.Join(execDir, "token")

	flag.StringVar(&cfg.ServerAddr, "a", getEnv("SERVER_ADDR", "https://localhost:8080"), "Server address")
	flag.StringVar(&cfg.LogLevel, "l", getEnv("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	flag.BoolVar(&cfg.TLSInsecure, "v", getBoolEnv("TLS_INSECURE", false), "Disable TLS certificate verification")
	flag.StringVar(&cfg.CachePath, "c", getEnv("CACHE_PATH", defaultCache), "Path to the local cache file")
	flag.StringVar(&cfg.TokenPath, "t", getEnv("TOKEN_PATH", defaultToken), "Path to the token file")

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

// getBoolEnv retrieves a boolean from an environment variable or returns a default value.
func getBoolEnv(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if v, err := strconv.ParseBool(value); err == nil {
			return v
		}
	}
	return defaultValue
}
