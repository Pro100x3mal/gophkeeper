// Package logger provides structured logging functionality using Zap logger.
//
// This package configures and creates logger instances with ISO8601 time encoding
// and configurable log levels for the GophKeeper application.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap logger instance with the specified log level.
// The logger uses development configuration with ISO8601 time encoding.
//
// Parameters:
//   - level: log level string (e.g., "debug", "info", "warn", "error")
//
// Returns a configured zap.Logger instance or an error if configuration fails.
func New(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	config := zap.NewDevelopmentConfig()
	config.Level = lvl
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger config: %w", err)
	}

	return logger, nil
}
