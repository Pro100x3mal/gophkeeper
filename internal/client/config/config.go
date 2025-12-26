package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr   string
	LogLevel     string
	TLSInsecure  bool
	CachePath    string
	TokenPath    string
	BuildVersion string
	BuildDate    string
}

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

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if v, err := strconv.ParseBool(value); err == nil {
			return v
		}
	}
	return defaultValue
}
