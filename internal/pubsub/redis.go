//go:build redis

package pubsub

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisPubSub implements PubSub using Redis Pub/Sub.
type RedisPubSub struct {
	client *redis.Client
}

func NewRedis(redisURL string) (*RedisPubSub, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	return &RedisPubSub{client: client}, nil
}

func (ps *RedisPubSub) Publish(ctx context.Context, channel string, message []byte) error {
	if err := ps.client.Publish(ctx, channel, message).Err(); err != nil {
		return fmt.Errorf("redis publish: %w", err)
	}
	return nil
}

func (ps *RedisPubSub) Subscribe(ctx context.Context, channel string) (<-chan []byte, error) {
	sub := ps.client.Subscribe(ctx, channel)
	if _, err := sub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("redis subscribe: %w", err)
	}

	ch := make(chan []byte, channelBufferSize)
	go func() {
		defer close(ch)
		for msg := range sub.Channel() {
			select {
			case ch <- []byte(msg.Payload):
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func (ps *RedisPubSub) Close() error {
	return ps.client.Close()
}
