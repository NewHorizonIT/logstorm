package ingestion_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/ingestion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToLogEvent_ValidInput(t *testing.T) {
	projectID := uuid.New()
	input := &ingestion.IngestLogInput{
		Level:   "info",
		Message: "hello logstorm",
		Service: "api",
	}

	event, err := input.ToLogEvent(projectID, "sdk")

	require.NoError(t, err)
	assert.Equal(t, projectID, event.ProjectID)
	assert.Equal(t, "info", event.Level)
	assert.Equal(t, "hello logstorm", event.Message)
	assert.Equal(t, "api", event.Service)
	assert.Equal(t, "sdk", event.Source)
	assert.Equal(t, "production", event.Environment)
	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestToLogEvent_LevelNormalisedToLower(t *testing.T) {
	input := &ingestion.IngestLogInput{Level: "INFO", Message: "msg"}
	event, err := input.ToLogEvent(uuid.New(), "sdk")
	require.NoError(t, err)
	assert.Equal(t, "info", event.Level)
}

func TestToLogEvent_AllValidLevels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "fatal"} {
		t.Run(level, func(t *testing.T) {
			input := &ingestion.IngestLogInput{Level: level, Message: "msg"}
			_, err := input.ToLogEvent(uuid.New(), "sdk")
			assert.NoError(t, err)
		})
	}
}

func TestToLogEvent_EmptyLevel(t *testing.T) {
	input := &ingestion.IngestLogInput{Message: "msg"}
	_, err := input.ToLogEvent(uuid.New(), "sdk")
	assert.ErrorIs(t, err, ingestion.ErrValidation)
}

func TestToLogEvent_InvalidLevel(t *testing.T) {
	input := &ingestion.IngestLogInput{Level: "critical", Message: "msg"}
	_, err := input.ToLogEvent(uuid.New(), "sdk")
	assert.ErrorIs(t, err, ingestion.ErrValidation)
}

func TestToLogEvent_EmptyMessage(t *testing.T) {
	input := &ingestion.IngestLogInput{Level: "info", Message: "   "}
	_, err := input.ToLogEvent(uuid.New(), "sdk")
	assert.ErrorIs(t, err, ingestion.ErrValidation)
}

func TestToLogEvent_EmptySource(t *testing.T) {
	input := &ingestion.IngestLogInput{Level: "info", Message: "msg"}
	_, err := input.ToLogEvent(uuid.New(), "")

	assert.ErrorIs(t, err, ingestion.ErrValidation)
}

func TestToLogEvent_MessageTooLarge(t *testing.T) {
	input := &ingestion.IngestLogInput{
		Level:   "info",
		Message: strings.Repeat("a", 10*1024*1024+1),
	}
	_, err := input.ToLogEvent(uuid.New(), "sdk")
	assert.ErrorIs(t, err, ingestion.ErrValidation)
}

func TestToLogEvent_EnvironmentDefaultsToProduction(t *testing.T) {
	for _, env := range []string{"", "   "} {
		input := &ingestion.IngestLogInput{Level: "info", Message: "msg", Environment: env}
		event, err := input.ToLogEvent(uuid.New(), "sdk")
		require.NoError(t, err)
		assert.Equal(t, "production", event.Environment)
	}
}

func TestToLogEvent_ExplicitEnvironment(t *testing.T) {
	input := &ingestion.IngestLogInput{Level: "info", Message: "msg", Environment: "staging"}
	event, err := input.ToLogEvent(uuid.New(), "sdk")
	require.NoError(t, err)
	assert.Equal(t, "staging", event.Environment)
}

func TestToLogEvent_TimestampFromInput(t *testing.T) {
	ts := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	input := &ingestion.IngestLogInput{Level: "info", Message: "msg", Timestamp: &ts}
	event, err := input.ToLogEvent(uuid.New(), "sdk")
	require.NoError(t, err)
	assert.Equal(t, ts, event.Timestamp)
}

func TestToLogEvent_NilTimestampDefaultsToNow(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	input := &ingestion.IngestLogInput{Level: "info", Message: "msg"}
	event, err := input.ToLogEvent(uuid.New(), "sdk")
	require.NoError(t, err)
	assert.True(t, event.Timestamp.After(before))
}

func TestToLogEvent_NilAttributesDefaultsToEmptyMap(t *testing.T) {
	input := &ingestion.IngestLogInput{Level: "info", Message: "msg"}
	event, err := input.ToLogEvent(uuid.New(), "sdk")
	require.NoError(t, err)
	assert.NotNil(t, event.Attributes)
	assert.Empty(t, event.Attributes)
}

func TestToLogEvent_ProjectIDInjected(t *testing.T) {
	projectID := uuid.New()
	input := &ingestion.IngestLogInput{Level: "info", Message: "msg"}
	event, err := input.ToLogEvent(projectID, "sdk")
	require.NoError(t, err)
	assert.Equal(t, projectID, event.ProjectID)
}

func TestToLogEvent_SourceInjected(t *testing.T) {
	for _, source := range []string{"sdk", "ui", "agent"} {
		t.Run(source, func(t *testing.T) {
			input := &ingestion.IngestLogInput{Level: "info", Message: "msg"}
			event, err := input.ToLogEvent(uuid.New(), source)
			require.NoError(t, err)
			assert.Equal(t, source, event.Source)
		})
	}
}