package pubsub

import "context"

// PubSub defines the interface for publish/subscribe messaging.
type PubSub interface {
	Publish(ctx context.Context, channel string, message []byte) error
	Subscribe(ctx context.Context, channel string) (<-chan []byte, error)
	Close() error
}
