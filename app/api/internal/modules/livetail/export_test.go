package livetail

import (
	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/logstorm/api/internal/modules/project"
)

// LiveTailConsumerForTest bọc LiveTailConsumer để expose processRecords
// cho black-box unit test mà không cần kết nối Redpanda thật.
type LiveTailConsumerForTest struct {
	c *LiveTailConsumer
}

// NewLiveTailConsumerForTest tạo consumer với fake broadcaster, bỏ qua kgo.Client.
//
// Đầu vào: b — bất kỳ type nào implement Broadcaster (thường là fakeHub trong test).
// Đầu ra:  *LiveTailConsumerForTest sẵn sàng gọi ProcessRecords.
func NewLiveTailConsumerForTest(b Broadcaster) *LiveTailConsumerForTest {
	return &LiveTailConsumerForTest{
		c: &LiveTailConsumer{
			broadcaster: b,
			log:         zerolog.Nop(),
		},
	}
}

// ProcessRecords gọi trực tiếp processRecords — dùng trong test để kiểm tra
// logic staleness filter và deserialization mà không cần broker thật.
func (t *LiveTailConsumerForTest) ProcessRecords(records []*kgo.Record) {
	t.c.processRecords(records)
}

// NewHandlerForTest tạo Handler với Subscriber interface (không cần *Hub thật),
// dùng trong handler_test để inject fake Subscriber và fake ProjectRepository.
//
// Đầu vào:
//   - hub: bất kỳ type nào implement Subscriber (thường là fakeSubscriber trong test).
//   - projectRepo: bất kỳ type nào implement ProjectRepository.
func NewHandlerForTest(hub Subscriber, projectRepo project.ProjectRepository) *Handler {
	return &Handler{hub: hub, projectRepo: projectRepo}
}
