package ingestion_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/logstorm/api/internal/database"
	"github.com/logstorm/api/internal/modules/ingestion"
)

// --- fake driver.Batch ---

type fakeBatch struct {
	appendCalls int
	sendErr     error
	appendErr   error
	sent        bool
}

func (b *fakeBatch) Append(v ...any) error {
	if b.appendErr != nil {
		return b.appendErr
	}
	b.appendCalls++
	return nil
}
func (b *fakeBatch) AppendStruct(v any) error      { return nil }
func (b *fakeBatch) Column(int) driver.BatchColumn { return nil }
func (b *fakeBatch) Flush() error                  { return nil }
func (b *fakeBatch) Abort() error                  { return nil }
func (b *fakeBatch) IsSent() bool                  { return b.sent }
func (b *fakeBatch) Rows() int                     { return b.appendCalls }
func (b *fakeBatch) Columns() []column.Interface   { return nil }
func (b *fakeBatch) Close() error                  { return nil }
func (b *fakeBatch) Send() error {
	if b.sendErr != nil {
		return b.sendErr
	}
	b.sent = true
	return nil
}

var _ driver.Batch = (*fakeBatch)(nil)

// --- fake ClickHouseClient ---

type fakeClickHouse struct {
	prepareBatchFn func(ctx context.Context, query string) (driver.Batch, error)
}

func (f *fakeClickHouse) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	return f.prepareBatchFn(ctx, query)
}
func (f *fakeClickHouse) Exec(ctx context.Context, query string, args ...interface{}) error {
	return nil
}
func (f *fakeClickHouse) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return nil, nil
}
func (f *fakeClickHouse) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return nil
}
func (f *fakeClickHouse) Ping() error        { return nil }
func (f *fakeClickHouse) HealthCheck() error { return nil }
func (f *fakeClickHouse) Close() error       { return nil }

var _ database.ClickHouseClient = (*fakeClickHouse)(nil)

// --- fake DLQPublisher ---

type fakeDLQ struct {
	calls [][]byte
}

func (f *fakeDLQ) PublishDLQ(_ context.Context, key string, data []byte) error {
	f.calls = append(f.calls, data)
	return nil
}

// --- helpers ---

func nopLogger() zerolog.Logger { return zerolog.Nop() }

func sampleEvent() *ingestion.LogEvent {
	return &ingestion.LogEvent{
		ID:          uuid.New(),
		Timestamp:   time.Now().UTC(),
		ProjectID:   uuid.New(),
		Environment: "production",
		Level:       "info",
		Service:     "api",
		Message:     "hello world",
		Source:      "sdk",
		Attributes:  map[string]string{},
	}
}

func newProc(ch database.ClickHouseClient) *ingestion.LogProcessorForTest {
	return ingestion.NewLogProcessorForTest(ch, &fakeDLQ{}, nopLogger())
}

func newProcWithDLQ(ch database.ClickHouseClient, dlq *fakeDLQ) *ingestion.LogProcessorForTest {
	return ingestion.NewLogProcessorForTest(ch, dlq, nopLogger())
}

// --- tests ---

func TestFlushBatch_EmptyBatch_SkipsPrepareBatch(t *testing.T) {
	called := false
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			called = true
			return &fakeBatch{}, nil
		},
	}

	newProc(ch).FlushBatch(context.Background(), nil)
	assert.False(t, called, "PrepareBatch must not be called for empty batch")

	newProc(ch).FlushBatch(context.Background(), []*ingestion.LogEvent{})
	assert.False(t, called, "PrepareBatch must not be called for empty slice")
}

func TestFlushBatch_Success_AppendsAndSends(t *testing.T) {
	batch := &fakeBatch{}
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return batch, nil
		},
	}

	events := []*ingestion.LogEvent{sampleEvent(), sampleEvent(), sampleEvent()}
	newProc(ch).FlushBatch(context.Background(), events)

	assert.Equal(t, 3, batch.appendCalls, "must append one row per event")
	assert.True(t, batch.sent, "Send must be called after all appends")
}

func TestFlushBatch_PrepareBatchError_DoesNotPanic(t *testing.T) {
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return nil, errors.New("connection refused")
		},
	}

	require.NotPanics(t, func() {
		newProc(ch).FlushBatch(context.Background(), []*ingestion.LogEvent{sampleEvent()})
	})
}

func TestFlushBatch_SendError_DoesNotPanic(t *testing.T) {
	batch := &fakeBatch{sendErr: errors.New("broker unavailable")}
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return batch, nil
		},
	}

	require.NotPanics(t, func() {
		newProc(ch).FlushBatch(context.Background(), []*ingestion.LogEvent{sampleEvent()})
	})
	assert.Equal(t, 1, batch.appendCalls, "Append must still be called before Send fails")
}

func TestFlushBatch_AppendError_StillCallsSend(t *testing.T) {
	batch := &fakeBatch{appendErr: errors.New("bad field type")}
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return batch, nil
		},
	}

	require.NotPanics(t, func() {
		newProc(ch).FlushBatch(context.Background(), []*ingestion.LogEvent{sampleEvent(), sampleEvent()})
	})
	assert.True(t, batch.sent, "Send must be attempted even when Append errors")
}

func TestFlushBatch_LargeBatch_AllRowsAppended(t *testing.T) {
	batch := &fakeBatch{}
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return batch, nil
		},
	}

	const n = 100
	events := make([]*ingestion.LogEvent, n)
	for i := range events {
		events[i] = sampleEvent()
	}

	newProc(ch).FlushBatch(context.Background(), events)
	assert.Equal(t, n, batch.appendCalls)
	assert.True(t, batch.sent)
}

func TestFlushBatch_PrepareBatchError_SendsBatchToDLQ(t *testing.T) {
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return nil, errors.New("connection refused")
		},
	}
	dlq := &fakeDLQ{}
	events := []*ingestion.LogEvent{sampleEvent(), sampleEvent()}

	newProcWithDLQ(ch, dlq).FlushBatch(context.Background(), events)

	assert.Len(t, dlq.calls, 2, "all events must be sent to DLQ when PrepareBatch fails")
}

func TestFlushBatch_SendError_SendsBatchToDLQ(t *testing.T) {
	batch := &fakeBatch{sendErr: errors.New("broker unavailable")}
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return batch, nil
		},
	}
	dlq := &fakeDLQ{}
	events := []*ingestion.LogEvent{sampleEvent(), sampleEvent(), sampleEvent()}

	newProcWithDLQ(ch, dlq).FlushBatch(context.Background(), events)

	assert.Len(t, dlq.calls, 3, "all events must be sent to DLQ when Send fails")
}

func TestFlushBatch_Success_NoDLQ(t *testing.T) {
	batch := &fakeBatch{}
	ch := &fakeClickHouse{
		prepareBatchFn: func(_ context.Context, _ string) (driver.Batch, error) {
			return batch, nil
		},
	}
	dlq := &fakeDLQ{}

	newProcWithDLQ(ch, dlq).FlushBatch(context.Background(), []*ingestion.LogEvent{sampleEvent()})

	assert.Empty(t, dlq.calls, "DLQ must not be called on success")
}

func TestNewLogProcessor_RequiresLiveBroker(t *testing.T) {
	t.Skip("integration test — requires live Redpanda broker")
}
