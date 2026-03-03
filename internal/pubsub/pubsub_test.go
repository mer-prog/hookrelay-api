package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryPubSub_PublishSubscribe(t *testing.T) {
	ps := NewInMemory()
	defer ps.Close()
	ctx := context.Background()

	ch, err := ps.Subscribe(ctx, "events")
	require.NoError(t, err)

	msg := []byte(`{"type":"test"}`)
	err = ps.Publish(ctx, "events", msg)
	require.NoError(t, err)

	select {
	case received := <-ch:
		assert.Equal(t, msg, received)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestInMemoryPubSub_MultipleSubscribers(t *testing.T) {
	ps := NewInMemory()
	defer ps.Close()
	ctx := context.Background()

	ch1, err := ps.Subscribe(ctx, "events")
	require.NoError(t, err)
	ch2, err := ps.Subscribe(ctx, "events")
	require.NoError(t, err)

	msg := []byte("hello")
	err = ps.Publish(ctx, "events", msg)
	require.NoError(t, err)

	for _, ch := range []<-chan []byte{ch1, ch2} {
		select {
		case received := <-ch:
			assert.Equal(t, msg, received)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for message")
		}
	}
}

func TestInMemoryPubSub_NoSubscribers(t *testing.T) {
	ps := NewInMemory()
	defer ps.Close()
	ctx := context.Background()

	err := ps.Publish(ctx, "nobody", []byte("msg"))
	assert.NoError(t, err)
}

func TestInMemoryPubSub_CloseErrors(t *testing.T) {
	ps := NewInMemory()
	ctx := context.Background()

	require.NoError(t, ps.Close())

	err := ps.Publish(ctx, "events", []byte("msg"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	_, err = ps.Subscribe(ctx, "events")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestInMemoryPubSub_ChannelIsolation(t *testing.T) {
	ps := NewInMemory()
	defer ps.Close()
	ctx := context.Background()

	ch1, err := ps.Subscribe(ctx, "chan-a")
	require.NoError(t, err)
	ch2, err := ps.Subscribe(ctx, "chan-b")
	require.NoError(t, err)

	err = ps.Publish(ctx, "chan-a", []byte("for-a"))
	require.NoError(t, err)

	select {
	case received := <-ch1:
		assert.Equal(t, []byte("for-a"), received)
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}

	select {
	case <-ch2:
		t.Fatal("chan-b should not receive chan-a messages")
	case <-time.After(50 * time.Millisecond):
		// expected — no message
	}
}
