package logger

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWithLogger_ShouldStoreLoggerInContext(
	t *testing.T,
) {
	t.Parallel()

	// Arrange
	expectedLogger := zerolog.New(nil)

	ctx := context.Background()

	// Act
	ctx = WithLogger(
		ctx,
		expectedLogger,
	)

	actualLogger := FromContext(ctx)

	// Assert
	assert.Equal(
		t,
		expectedLogger.GetLevel(),
		actualLogger.GetLevel(),
	)
}

func TestFromContext_WhenLoggerDoesNotExist(
	t *testing.T,
) {
	t.Parallel()

	// Arrange
	ctx := context.Background()

	// Act
	actualLogger := FromContext(ctx)

	// Assert
	assert.Equal(
		t,
		zerolog.Disabled,
		actualLogger.GetLevel(),
	)
}
