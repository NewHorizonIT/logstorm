package livetail_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/ingestion"
	"github.com/logstorm/api/internal/modules/livetail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

// fakeHub là fake implementation của Broadcaster, ghi lại tất cả event được broadcast.
type fakeHub struct {
	received []*ingestion.LogEvent
}

// Broadcast ghi lại event vào received để test có thể assert.
func (f *fakeHub) Broadcast(events []*ingestion.LogEvent) {
	f.received = append(f.received, events...)
}

// makeRecord chuyển LogEvent thành *kgo.Record (JSON encode vào Value).
func makeRecord(event *ingestion.LogEvent) *kgo.Record {
	data, _ := json.Marshal(event)
	return &kgo.Record{Value: data}
}

// makeFreshEvent tạo LogEvent với Timestamp = now, đảm bảo qua staleness filter.
func makeFreshEvent(projectID uuid.UUID) *ingestion.LogEvent {
	return &ingestion.LogEvent{
		ID:        uuid.New(),
		Timestamp: time.Now().UTC(),
		ProjectID: projectID,
		Level:     "info",
		Service:   "api",
		Message:   "hello",
	}
}

// TestConsumer_FreshEvent_HubReceivesIt kiểm tra luồng happy path:
// event fresh được deserialize đúng và hub nhận được đúng event đó.
func TestConsumer_FreshEvent_HubReceivesIt(t *testing.T) {
	hub := &fakeHub{}
	consumer := livetail.NewLiveTailConsumerForTest(hub)
	projectID := uuid.New()

	event := makeFreshEvent(projectID)
	consumer.ProcessRecords([]*kgo.Record{makeRecord(event)})

	require.Len(t, hub.received, 1, "hub phải nhận đúng 1 event")
	assert.Equal(t, event.ID, hub.received[0].ID)
	assert.Equal(t, event.Message, hub.received[0].Message)
	assert.Equal(t, event.ProjectID, hub.received[0].ProjectID)
}

// TestConsumer_StaleEvent_Dropped kiểm tra staleness filter:
// event có Timestamp cũ hơn StalenessThreshold bị bỏ qua, hub không nhận gì.
func TestConsumer_StaleEvent_Dropped(t *testing.T) {
	hub := &fakeHub{}
	consumer := livetail.NewLiveTailConsumerForTest(hub)
	projectID := uuid.New()

	staleEvent := makeFreshEvent(projectID)
	// Đặt timestamp cũ hơn ngưỡng 1 giây để chắc chắn bị drop
	staleEvent.Timestamp = time.Now().UTC().Add(-(livetail.StalenessThreshold + time.Second))

	consumer.ProcessRecords([]*kgo.Record{makeRecord(staleEvent)})

	assert.Empty(t, hub.received, "event stale không được broadcast tới hub")
}

// TestConsumer_DeserializationError_Skipped kiểm tra rằng record có JSON lỗi
// bị bỏ qua và không gây panic, hub không nhận gì.
func TestConsumer_DeserializationError_Skipped(t *testing.T) {
	hub := &fakeHub{}
	consumer := livetail.NewLiveTailConsumerForTest(hub)

	invalidRecord := &kgo.Record{Value: []byte(`not valid json {`)}

	require.NotPanics(t, func() {
		consumer.ProcessRecords([]*kgo.Record{invalidRecord})
	})
	assert.Empty(t, hub.received, "record lỗi JSON không được broadcast tới hub")
}

// TestConsumer_MultipleEvents_AllNonStaleForwarded kiểm tra xử lý batch hỗn hợp:
// event fresh được forward, event stale bị drop, chỉ fresh event xuất hiện trong hub.
func TestConsumer_MultipleEvents_AllNonStaleForwarded(t *testing.T) {
	hub := &fakeHub{}
	consumer := livetail.NewLiveTailConsumerForTest(hub)
	projectID := uuid.New()

	fresh1 := makeFreshEvent(projectID)
	fresh2 := makeFreshEvent(projectID)
	staleEvent := makeFreshEvent(projectID)
	staleEvent.Timestamp = time.Now().UTC().Add(-(livetail.StalenessThreshold + time.Second))

	consumer.ProcessRecords([]*kgo.Record{
		makeRecord(fresh1),
		makeRecord(staleEvent),
		makeRecord(fresh2),
	})

	require.Len(t, hub.received, 2, "chỉ 2 event fresh được broadcast, stale bị bỏ")
	ids := []uuid.UUID{hub.received[0].ID, hub.received[1].ID}
	assert.Contains(t, ids, fresh1.ID)
	assert.Contains(t, ids, fresh2.ID)
}

// TestConsumer_EmptyBatch_NoBroadcast kiểm tra rằng batch nil hoặc rỗng
// không gọi Broadcast (tránh call không cần thiết vào hub).
func TestConsumer_EmptyBatch_NoBroadcast(t *testing.T) {
	hub := &fakeHub{}
	consumer := livetail.NewLiveTailConsumerForTest(hub)

	consumer.ProcessRecords(nil)
	consumer.ProcessRecords([]*kgo.Record{})

	assert.Empty(t, hub.received, "batch rỗng không được gọi Broadcast")
}

// TestConsumer_AllStale_NoBroadcast kiểm tra rằng khi toàn bộ batch đều stale,
// Broadcast không được gọi (batch sau filter rỗng).
func TestConsumer_AllStale_NoBroadcast(t *testing.T) {
	hub := &fakeHub{}
	consumer := livetail.NewLiveTailConsumerForTest(hub)
	projectID := uuid.New()

	stale1 := makeFreshEvent(projectID)
	stale1.Timestamp = time.Now().UTC().Add(-(livetail.StalenessThreshold + time.Second))
	stale2 := makeFreshEvent(projectID)
	stale2.Timestamp = time.Now().UTC().Add(-(livetail.StalenessThreshold + time.Minute))

	consumer.ProcessRecords([]*kgo.Record{makeRecord(stale1), makeRecord(stale2)})

	assert.Empty(t, hub.received, "khi toàn bộ event stale, hub không nhận gì")
}

// TestNewLiveTailConsumer_RequiresLiveBroker là integration test, cần broker thật.
func TestNewLiveTailConsumer_RequiresLiveBroker(t *testing.T) {
	t.Skip("integration test — requires live Redpanda broker")
}
