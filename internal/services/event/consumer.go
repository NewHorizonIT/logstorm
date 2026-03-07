package event

import "context"

// MessageHandler handles consumed messages
type MessageHandler func(ctx context.Context, key, value []byte) error

type IConsumer interface {
	// Consume starts consuming from topic and calls handler for each message
	Consume(ctx context.Context, topic string, handler MessageHandler) error
	Stop() error
}
