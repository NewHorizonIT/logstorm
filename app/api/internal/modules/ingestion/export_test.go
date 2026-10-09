package ingestion

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/logstorm/api/internal/database"
)

// LogProcessorForTest exposes flush for black-box unit tests.
type LogProcessorForTest struct {
	p *LogProcessor
}

func NewLogProcessorForTest(ch database.ClickHouseClient, dlq DLQPublisher, log zerolog.Logger) *LogProcessorForTest {
	return &LogProcessorForTest{
		p: &LogProcessor{ch: ch, dlq: dlq, log: log},
	}
}

func (t *LogProcessorForTest) FlushBatch(ctx context.Context, batch []*LogEvent) {
	t.p.flush(ctx, batch)
}
