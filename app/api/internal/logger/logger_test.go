package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/logstorm/api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ShouldWriteLogToFile(t *testing.T) {
	t.Parallel()

	// Arrange
	tempDir := t.TempDir()

	logPath := filepath.Join(
		tempDir,
		"logs",
		"app.log",
	)

	appLogger, err := New(
		config.LoggingConfig{
			Environment:    "test",
			Level:          "debug",
			ConsoleEnabled: false,
			FileEnabled:    true,
			FilePath:       logPath,

			RotationMaxSize:    100,
			RotationMaxBackups: 3,
			RotationMaxAge:     7,
			CompressionEnabled: false,
		},
	)

	require.NoError(t, err)

	// Act
	appLogger.Zerolog.Info().
		Str("test_key", "test_value").
		Msg("test_message")

	err = appLogger.Close()

	// Assert
	require.NoError(t, err)

	data, err := os.ReadFile(logPath)

	require.NoError(t, err)

	content := string(data)

	assert.Contains(
		t,
		content,
		"test_message",
	)

	assert.Contains(
		t,
		content,
		"test_value",
	)
}

func TestNew_ShouldCreateLogDirectory(t *testing.T) {
	t.Parallel()

	// Arrange
	tempDir := t.TempDir()

	logPath := filepath.Join(
		tempDir,
		"nested",
		"logs",
		"app.log",
	)

	// Act
	appLogger, err := New(
		config.LoggingConfig{
			Environment:    "test",
			Level:          "info",
			ConsoleEnabled: false,

			FileEnabled: true,
			FilePath:    logPath,

			RotationMaxSize:    100,
			RotationMaxBackups: 3,
			RotationMaxAge:     7,
			CompressionEnabled: false,
		},
	)

	// Assert
	require.NoError(t, err)

	defer func() {
		require.NoError(
			t,
			appLogger.Close(),
		)
	}()

	info, err := os.Stat(
		filepath.Dir(logPath),
	)

	require.NoError(t, err)

	assert.True(
		t,
		info.IsDir(),
	)
}
