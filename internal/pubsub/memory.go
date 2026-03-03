package pubsub

import (
	"context"
	"fmt"
	"sync"
)

const channelBufferSize = 100

// InMemoryPubSub implements PubSub using in-process channels.
type InMemoryPubSub struct {
	mu     sync.RWMutex
	subs   map[string][]chan []byte
	closed bool
}

func NewInMemory() *InMemoryPubSub {
	return &InMemoryPubSub{
		subs: make(map[string][]chan []byte),
	}
}

func (ps *InMemoryPubSub) Publish(_ context.Context, channel string, message []byte) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return fmt.Errorf("pubsub is closed")
	}

	for _, ch := range ps.subs[channel] {
		select {
		case ch <- message:
		default:
			// Subscriber too slow — drop message to avoid blocking.
		}
	}
	return nil
}

func (ps *InMemoryPubSub) Subscribe(_ context.Context, channel string) (<-chan []byte, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil, fmt.Errorf("pubsub is closed")
	}

	ch := make(chan []byte, channelBufferSize)
	ps.subs[channel] = append(ps.subs[channel], ch)
	return ch, nil
}

func (ps *InMemoryPubSub) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil
	}
	ps.closed = true

	for _, subscribers := range ps.subs {
		for _, ch := range subscribers {
			close(ch)
		}
	}
	ps.subs = nil
	return nil
}
