package notifier_test

import (
	"context"
	"github.com/sealbro/go-feed-me/pkg/graceful"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"github.com/sealbro/go-feed-me/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestSubscriptionManagerSubscribeAndClose(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	newLogger, err := logger.NewLogger(&logger.Config{LogLevel: "warn"})
	assert.NoError(t, err, "error should be nil")

	closer := graceful.NewShutdownCloser()
	manager := notifier.NewSubscriptionManager[int](newLogger, closer)
	ctx := context.Background()
	subscriber, err := manager.AddSubscriber(ctx, "subscriber-1")

	assert.NotNil(t, subscriber, "subscriber should not be nil")
	assert.NoError(t, err, "error should be nil")

	go func() { manager.Notify(items...) }()
	values := <-subscriber

	assert.Equal(t, items, values, "values should be equal")

	err = manager.Close()
	assert.NoError(t, err, "error should be nil")

	go func() { manager.Notify(items...) }()
	values = <-subscriber

	assert.Empty(t, values, "values should be empty after close")

	subscriber, err = manager.AddSubscriber(ctx, "subscriber-2")

	assert.Nil(t, subscriber, "subscriber should be nil after close")
	assert.ErrorAs(t, err, &notifier.ErrSubscriptionManagerClosed, "error should be ErrSubscriptionManagerClosed")
}

func TestSubscriptionManagerGetAllEventsAfterClose(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8}

	newLogger, err := logger.NewLogger(&logger.Config{LogLevel: "warn"})
	assert.NoError(t, err, "error should be nil")

	closer := graceful.NewShutdownCloser()
	manager := notifier.NewSubscriptionManager[int](newLogger, closer)
	ctx := context.Background()
	subscriber, err := manager.AddSubscriber(ctx, "subscriber-1")

	assert.NotNil(t, subscriber, "subscriber should not be nil")
	assert.NoError(t, err, "error should be nil")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		manager.Notify(items...)
		wg.Done()
	}()
	wg.Wait()

	err = manager.Close()
	assert.NoError(t, err, "error should be nil")

	values := <-subscriber

	assert.Empty(t, values, "values should be empty after close")
	//assert.Equal(t, items, values, "values should be equal")
}
