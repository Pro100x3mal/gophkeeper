package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_Client(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{"Env set", "TEST_KEY_CLIENT", "default", "custom", true, "custom"},
		{"Env not set", "TEST_KEY_CLIENT_NOTSET", "default", "", false, "default"},
		{"Empty env", "TEST_KEY_CLIENT_EMPTY", "default", "", true, ""},
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

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		setEnv       bool
		expected     bool
	}{
		{"True string", "TEST_BOOL", false, "true", true, true},
		{"False string", "TEST_BOOL", true, "false", true, false},
		{"1 value", "TEST_BOOL", false, "1", true, true},
		{"0 value", "TEST_BOOL", true, "0", true, false},
		{"Invalid value", "TEST_BOOL", true, "invalid", true, true},
		{"Not set", "TEST_BOOL_NOTSET", false, "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getBoolEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClientConfig_Struct(t *testing.T) {
	cfg := &Config{
		ServerAddr:   "https://server.com",
		LogLevel:     "warn",
		TLSInsecure:  true,
		CachePath:    "/path/to/cache",
		TokenPath:    "/path/to/token",
		BuildVersion: "1.0.0",
		BuildDate:    "2024-01-01",
	}

	assert.Equal(t, "https://server.com", cfg.ServerAddr)
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.True(t, cfg.TLSInsecure)
	assert.Equal(t, "/path/to/cache", cfg.CachePath)
	assert.Equal(t, "/path/to/token", cfg.TokenPath)
	assert.Equal(t, "1.0.0", cfg.BuildVersion)
	assert.Equal(t, "2024-01-01", cfg.BuildDate)
}
