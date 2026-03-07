package event

import "context"

type IPublisher interface {
	Publish(ctx context.Context, topic string, key []byte, message []byte) error
}
