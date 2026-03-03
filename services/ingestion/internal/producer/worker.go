package producer

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

func (p *KafkaProducer) worker(ctx context.Context) {
	defer p.wg.Done()

	batch := make([]kafka.Message, 0, p.batchSize)
	ticker := time.NewTicker(p.flushEvery)

	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		p.sendWithRetry(ctx, batch)
		batch = batch[:0]
	}

	for {
		select {
		// If ctx is done, flush messages and exit
		case <-ctx.Done():
			flush()
			return

		// If get a message, add to batch
		case event, ok := <-p.queue:
			// If channel is closed, flush remaining messages and exit
			if !ok {
				flush()
				return
			}
			data, err := p.encoder.Encode(event)
			if err != nil {
				continue
			}
			msg := kafka.Message{
				Key:   []byte(event.Service),
				Value: data,
				Time:  event.Timestamp,
			}

			batch = append(batch, msg)
			if len(batch) >= p.batchSize {
				flush()
			}

		// If ticker ticks, flush messages
		case <-ticker.C:
			flush()
		}
	}

}
