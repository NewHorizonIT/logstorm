package search

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LogRecord struct {
	ID          uuid.UUID         `ch:"id"          json:"id"`
	Timestamp   time.Time         `ch:"timestamp"   json:"timestamp"`
	ProjectID   uuid.UUID         `ch:"project_id"  json:"project_id"`
	Environment string            `ch:"environment" json:"environment"`
	Level       string            `ch:"level"       json:"level"`
	Service     string            `ch:"service"     json:"service"`
	Message     string            `ch:"message"     json:"message"`
	TraceID     string            `ch:"trace_id"    json:"trace_id"`
	SpanID      string            `ch:"span_id"     json:"span_id"`
	Source      string            `ch:"source"      json:"source"`
	Attributes  map[string]string `ch:"attributes"  json:"attributes"`
}

type SearchQuery struct {
	ProjectID uuid.UUID
	Level     string
	Service   string
	Text      string
	From      time.Time
	To        time.Time
	Limit     int
	Offset    int
}

type SearchResult struct {
	Logs  []*LogRecord
	Total int64
}

var ErrValidation = errors.New("validation error")

func (q *SearchQuery) Validate() error {
	if q.Limit <= 0 {
		q.Limit = 100
	}
	if q.Limit > 1000 {
		return fmt.Errorf("%w: limit must not exceed 1000", ErrValidation)
	}
	if q.From.IsZero() {
		q.From = time.Now().UTC().Add(-time.Hour)
	}
	if q.To.IsZero() {
		q.To = time.Now().UTC()
	}
	if q.From.After(q.To) {
		return fmt.Errorf("%w: from must be before to", ErrValidation)
	}
	if q.Offset < 0 {
		return fmt.Errorf("%w: offset must be non-negative", ErrValidation)
	}
	return nil
}
