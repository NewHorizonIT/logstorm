package livetail_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/ingestion"
	"github.com/logstorm/api/internal/modules/livetail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeEvent tạo một LogEvent tối giản để dùng trong test.
func makeEvent(projectID uuid.UUID, level, service string) *ingestion.LogEvent {
	return &ingestion.LogEvent{
		ID:        uuid.New(),
		Timestamp: time.Now().UTC(),
		ProjectID: projectID,
		Level:     level,
		Service:   service,
		Message:   "test message",
	}
}

// TestHub_SubscribeAndReceive kiểm tra luồng cơ bản:
// subscriber đăng ký → Broadcast gửi event → channel nhận được record đúng.
func TestHub_SubscribeAndReceive(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	ch, unsub := hub.Subscribe(projectID, livetail.Filter{})
	defer unsub()

	event := makeEvent(projectID, "info", "api")
	hub.Broadcast([]*ingestion.LogEvent{event})

	select {
	case record := <-ch:
		assert.Equal(t, event.Message, record.Message)
		assert.Equal(t, event.Level, record.Level)
		assert.Equal(t, event.ProjectID, record.ProjectID)
	case <-time.After(time.Second):
		t.Fatal("hết timeout nhưng không nhận được event")
	}
}

// TestHub_FilterByLevel kiểm tra rằng event không khớp level filter
// không được gửi vào channel của subscriber.
func TestHub_FilterByLevel(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	ch, unsub := hub.Subscribe(projectID, livetail.Filter{Level: "error"})
	defer unsub()

	// Gửi event level "info" — không khớp filter "error"
	hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})

	select {
	case <-ch:
		t.Fatal("event level=info không được phép qua filter level=error")
	default:
	}
}

// TestHub_FilterByService kiểm tra rằng event không khớp service filter
// không được gửi vào channel của subscriber.
func TestHub_FilterByService(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	ch, unsub := hub.Subscribe(projectID, livetail.Filter{Service: "worker"})
	defer unsub()

	// Gửi event service "api" — không khớp filter "worker"
	hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})

	select {
	case <-ch:
		t.Fatal("event service=api không được phép qua filter service=worker")
	default:
	}
}

// TestHub_UnsubscribeRemovesChannel kiểm tra rằng sau khi unsub,
// subscriber không còn nhận event nào nữa.
func TestHub_UnsubscribeRemovesChannel(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	ch, unsub := hub.Subscribe(projectID, livetail.Filter{})
	unsub() // hủy đăng ký ngay lập tức

	hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})

	select {
	case <-ch:
		t.Fatal("channel đã unsub không được nhận event")
	default:
	}
}

// TestHub_UnsubscribeIdempotent kiểm tra rằng gọi unsub nhiều lần không panic.
func TestHub_UnsubscribeIdempotent(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	_, unsub := hub.Subscribe(projectID, livetail.Filter{})

	require.NotPanics(t, func() {
		unsub()
		unsub() // lần hai phải an toàn
	})
}

// TestHub_SlowConsumer_DoesNotBlock kiểm tra rằng khi buffer của subscriber
// đầy, Broadcast không bị block và vẫn trả về ngay lập tức.
func TestHub_SlowConsumer_DoesNotBlock(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	_, unsub := hub.Subscribe(projectID, livetail.Filter{}) // subscriber không đọc channel
	defer unsub()

	done := make(chan struct{})
	go func() {
		// Gửi nhiều hơn buffer size để chắc chắn buffer tràn
		for i := 0; i < livetail.SubscriberBufferSize+10; i++ {
			hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Broadcast bị block do slow consumer")
	}
}

// TestHub_MultipleSubscribersSameProject kiểm tra rằng tất cả subscriber
// cùng project đều nhận được event khi Broadcast.
func TestHub_MultipleSubscribersSameProject(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	ch1, unsub1 := hub.Subscribe(projectID, livetail.Filter{})
	defer unsub1()
	ch2, unsub2 := hub.Subscribe(projectID, livetail.Filter{})
	defer unsub2()

	hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})

	select {
	case <-ch1:
	case <-time.After(time.Second):
		t.Fatal("subscriber 1 không nhận được event")
	}

	select {
	case <-ch2:
	case <-time.After(time.Second):
		t.Fatal("subscriber 2 không nhận được event")
	}
}

// TestHub_ConcurrentSubscribeBroadcast kiểm tra không có data race khi
// nhiều goroutine gọi Subscribe/Broadcast/Unsub đồng thời.
// Phải chạy với -race flag để phát hiện race condition.
func TestHub_ConcurrentSubscribeBroadcast(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch, unsub := hub.Subscribe(projectID, livetail.Filter{})
			hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})
			unsub()
			// drain channel để tránh goroutine leak
			select {
			case <-ch:
			default:
			}
		}()
	}
	wg.Wait()
}

// TestHub_ZeroSubscribers_NoPanic kiểm tra rằng Broadcast khi không có
// subscriber nào không gây panic.
func TestHub_ZeroSubscribers_NoPanic(t *testing.T) {
	hub := livetail.NewHub()
	projectID := uuid.New()

	require.NotPanics(t, func() {
		hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectID, "info", "api")})
	})
}

// TestHub_DifferentProjectsIsolated kiểm tra tenant isolation:
// subscriber của project A không nhận event của project B.
func TestHub_DifferentProjectsIsolated(t *testing.T) {
	hub := livetail.NewHub()
	projectA := uuid.New()
	projectB := uuid.New()

	chA, unsubA := hub.Subscribe(projectA, livetail.Filter{})
	defer unsubA()

	// Broadcast event cho project B — chA không được nhận
	hub.Broadcast([]*ingestion.LogEvent{makeEvent(projectB, "info", "api")})

	select {
	case <-chA:
		t.Fatal("subscriber project A không được nhận event của project B")
	default:
	}
}
