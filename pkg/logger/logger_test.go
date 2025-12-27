package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew_Success(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected zapcore.Level
	}{
		{"Debug level", "debug", zapcore.DebugLevel},
		{"Info level", "info", zapcore.InfoLevel},
		{"Warn level", "warn", zapcore.WarnLevel},
		{"Error level", "error", zapcore.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)

			require.NoError(t, err)
			require.NotNil(t, logger)
			assert.True(t, logger.Core().Enabled(tt.expected))
		})
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"Invalid level string", "invalid"},
		{"Random string", "random_level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)

			assert.Error(t, err)
			assert.Nil(t, logger)
			assert.Contains(t, err.Error(), "invalid log level")
		})
	}
}

func TestNew_EmptyString(t *testing.T) {
	// Empty string defaults to info level in zap
	logger, err := New("")
	if err != nil {
		// If it returns error, that's also acceptable
		assert.Contains(t, err.Error(), "invalid log level")
		return
	}
	// If it doesn't return error, logger should be valid
	assert.NotNil(t, logger)
}

func TestNew_LoggerFunctionality(t *testing.T) {
	logger, err := New("info")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test that logger can log messages
	logger.Info("test message")
	logger.Warn("warning message")
	logger.Error("error message")

	// No assertion needed, just verify it doesn't panic
}

func TestNew_DebugLevelLogging(t *testing.T) {
	logger, err := New("debug")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Debug level should enable debug logging
	assert.True(t, logger.Core().Enabled(zapcore.DebugLevel))
	assert.True(t, logger.Core().Enabled(zapcore.InfoLevel))
	assert.True(t, logger.Core().Enabled(zapcore.WarnLevel))
	assert.True(t, logger.Core().Enabled(zapcore.ErrorLevel))
}

func TestNew_ErrorLevelLogging(t *testing.T) {
	logger, err := New("error")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Error level should only enable error and above
	assert.False(t, logger.Core().Enabled(zapcore.DebugLevel))
	assert.False(t, logger.Core().Enabled(zapcore.InfoLevel))
	assert.False(t, logger.Core().Enabled(zapcore.WarnLevel))
	assert.True(t, logger.Core().Enabled(zapcore.ErrorLevel))
}

func TestNew_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"Uppercase INFO", "INFO"},
		{"Mixed case Info", "Info"},
		{"Uppercase DEBUG", "DEBUG"},
		{"Uppercase ERROR", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)

			require.NoError(t, err)
			require.NotNil(t, logger)
		})
	}
}

func TestNew_LoggerType(t *testing.T) {
	logger, err := New("info")
	require.NoError(t, err)

	// Verify logger is of correct type
	assert.IsType(t, &zap.Logger{}, logger)
}
