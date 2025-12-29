package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{"Env set", "TEST_KEY", "default", "custom", true, "custom"},
		{"Env not set", "TEST_KEY_NOTSET", "default", "", false, "default"},
		{"Empty env", "TEST_KEY_EMPTY", "default", "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		setEnv       bool
		expected     time.Duration
	}{
		{"Valid duration", "TEST_DURATION", time.Hour, "30m", true, 30 * time.Minute},
		{"Invalid duration", "TEST_DURATION_INVALID", time.Hour, "invalid", true, time.Hour},
		{"Not set", "TEST_DURATION_NOTSET", 2 * time.Hour, "", false, 2 * time.Hour},
		{"Hours duration", "TEST_DURATION_HOURS", time.Hour, "24h", true, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvDuration(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		LogLevel:      "warn",
		ServerAddr:    "example.com:8080",
		DatabaseDSN:   "postgres://db",
		JWTSecret:     "secret",
		JWTExpiration: 12 * time.Hour,
		TLSCertFile:   "/path/to/cert",
		TLSKeyFile:    "/path/to/key",
		MasterKey:     "master-key",
	}

	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, "example.com:8080", cfg.ServerAddr)
	assert.Equal(t, "postgres://db", cfg.DatabaseDSN)
	assert.Equal(t, "secret", cfg.JWTSecret)
	assert.Equal(t, 12*time.Hour, cfg.JWTExpiration)
	assert.Equal(t, "/path/to/cert", cfg.TLSCertFile)
	assert.Equal(t, "/path/to/key", cfg.TLSKeyFile)
	assert.Equal(t, "master-key", cfg.MasterKey)
}
