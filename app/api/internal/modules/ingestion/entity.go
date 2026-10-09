package ingestion

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation error")

	validLevels = map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
)

const maxMessageSize = 10 * 1024 * 1024 // 10MB

type LogEvent struct {
	ID          uuid.UUID
	Timestamp   time.Time
	ProjectID   uuid.UUID
	Environment string
	Level       string
	Service     string
	Message     string
	TraceID     string
	SpanID      string
	Source      string
	Attributes  map[string]string
}

type IngestLogInput struct {
	Timestamp   *time.Time        `json:"timestamp"`
	Environment string            `json:"environment"`
	Level       string            `json:"level"`
	Service     string            `json:"service"`
	Message     string            `json:"message"`
	TraceID     string            `json:"trace_id"`
	SpanID      string            `json:"span_id"`
	Attributes  map[string]string `json:"attributes"`
}

func (input *IngestLogInput) ToLogEvent(projectID uuid.UUID, source string) (*LogEvent, error) {
	if input.Level == "" {
		return nil, fmt.Errorf("%w: level is required", ErrValidation)
	}
	if !validLevels[strings.ToLower(input.Level)] {
		return nil, fmt.Errorf("%w: level must be one of debug, info, warn, error, fatal", ErrValidation)
	}
	if strings.TrimSpace(input.Message) == "" {
		return nil, fmt.Errorf("%w: message is required", ErrValidation)
	}
	if source == "" {
		return nil, fmt.Errorf("%w: source is required ", ErrValidation)
	}
	if len(input.Message) > maxMessageSize {
		return nil, fmt.Errorf("%w: message exceeds maximum size of 10MB", ErrValidation)
	}

	ts := time.Now().UTC()
	if input.Timestamp != nil {
		ts = input.Timestamp.UTC()
	}

	env := input.Environment
	if strings.TrimSpace(env) == "" {
		env = "production"
	}

	attrs := input.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}

	return &LogEvent{
		ID:          uuid.New(),
		Timestamp:   ts,
		ProjectID:   projectID,
		Environment: env,
		Level:       strings.ToLower(input.Level),
		Service:     input.Service,
		Message:     input.Message,
		TraceID:     input.TraceID,
		SpanID:      input.SpanID,
		Source:      source,
		Attributes:  attrs,
	}, nil
}
