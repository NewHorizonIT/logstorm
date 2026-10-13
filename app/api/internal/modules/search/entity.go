package search

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LogRecord struct {
	ID          uuid.UUID         `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	ProjectID   uuid.UUID         `json:"project_id"`
	Environment string            `json:"environment"`
	Level       string            `json:"level"`
	Service     string            `json:"service"`
	Message     string            `json:"message"`
	TraceID     string            `json:"trace_id"`
	SpanID      string            `json:"span_id"`
	Source      string            `json:"source"`
	Attributes  map[string]string `json:"attributes"`
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
	return nil
}
