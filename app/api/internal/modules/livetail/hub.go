package livetail

import (
	"sync"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/ingestion"
	"github.com/logstorm/api/internal/modules/search"
)

const SubscriberBufferSize = 64

type Filter struct {
	Level   string // rỗng = nhận mọi level
	Service string // rỗng = nhận mọi service
}

type subscriber struct {
	ch     chan *search.LogRecord
	filter Filter
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[*subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uuid.UUID]map[*subscriber]struct{}),
	}
}

func (h *Hub) Subscribe(projectID uuid.UUID, filter Filter) (<-chan *search.LogRecord, func()) {
	sub := &subscriber{
		ch:     make(chan *search.LogRecord, SubscriberBufferSize),
		filter: filter,
	}

	h.mu.Lock()
	if h.subscribers[projectID] == nil {
		h.subscribers[projectID] = make(map[*subscriber]struct{})
	}
	h.subscribers[projectID][sub] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			delete(h.subscribers[projectID], sub)
			if len(h.subscribers[projectID]) == 0 {
				delete(h.subscribers, projectID)
			}
		})
	}

	return sub.ch, unsub
}

func (h *Hub) Broadcast(events []*ingestion.LogEvent) {
	if len(events) == 0 {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, event := range events {
		subs := h.subscribers[event.ProjectID]
		if len(subs) == 0 {
			continue
		}
		record := logEventToRecord(event)
		for sub := range subs {
			if sub.filter.Level != "" && sub.filter.Level != record.Level {
				continue
			}
			if sub.filter.Service != "" && sub.filter.Service != record.Service {
				continue
			}
			select {
			case sub.ch <- record:
			default: // buffer đầy — bỏ qua event cho subscriber này
			}
		}
	}
}

func logEventToRecord(e *ingestion.LogEvent) *search.LogRecord {
	return &search.LogRecord{
		ID:          e.ID,
		Timestamp:   e.Timestamp,
		ProjectID:   e.ProjectID,
		Environment: e.Environment,
		Level:       e.Level,
		Service:     e.Service,
		Message:     e.Message,
		TraceID:     e.TraceID,
		SpanID:      e.SpanID,
		Source:      e.Source,
		Attributes:  e.Attributes,
	}
}
