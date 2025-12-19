package config

import (
	"flag"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel      string
	ServerAddr    string
	DatabaseDSN   string
	JWTSecret     string
	JWTExpiration time.Duration
	TLSCertFile   string
	TLSKeyFile    string
	MasterKey     string
}

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

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}
	return defaultValue
}
