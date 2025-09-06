package pubsub

import "context"

type PubSub[T any] interface {
	Publish(ctx context.Context, msg T) error
	Subscribe(ctx context.Context, handler PubSubHandler[T]) error
	Close() error
}

type PubSubHandler[T any] func(messages []T) error
