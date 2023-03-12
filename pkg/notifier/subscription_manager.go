package notifier

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sealbro/go-feed-me/pkg/graceful"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"go.uber.org/zap"
	"time"
)

type batchConfig struct {
	BatchSize int
	BatchTime time.Duration
}

type SubscriptionManager[T any] struct {
	*batchConfig

	subscribers map[string]chan []T
	output      chan []T
	input       chan T
	closed      bool
	logger      *logger.Logger
}

func NewSubscriptionManager[T any](logger *logger.Logger, shutdownCloser *graceful.ShutdownCloser) *SubscriptionManager[T] {
	manager := &SubscriptionManager[T]{
		batchConfig: &batchConfig{
			BatchSize: 10,
			BatchTime: 1 * time.Minute,
		},
		logger:      logger,
		subscribers: map[string]chan []T{},
		input:       make(chan T),
		closed:      false,
	}

	manager.output = batchProcess(manager.input, manager.BatchSize, manager.BatchTime)

	go func() {
		for events := range manager.output {

			logger.Info("Send events to subscribers", zap.Int("links", len(events)), zap.Int("subscribers", len(manager.subscribers)))
			for _, subscriber := range manager.subscribers {
				subscriber <- events
			}
		}
	}()

	shutdownCloser.Register(manager)

	return manager
}

func (manager *SubscriptionManager[T]) Notify(events ...T) {
	if manager.closed {
		return
	}

	for _, eventData := range events {
		manager.input <- eventData
	}
}

func (manager *SubscriptionManager[T]) AddSubscriber(ctx context.Context) (chan []T, error) {
	if manager.closed {
		return nil, fmt.Errorf("SubscriptionManager closed the connection")
	}

	key := uuid.New().String()
	ch := make(chan []T)
	manager.subscribers[key] = ch

	manager.logger.Ctx(ctx).Info("SubscriptionManager - Added new subscriber", zap.String("key", key))

	go func() {
		<-ctx.Done()
		manager.RemoveSubscriber(key)
		manager.logger.Ctx(ctx).Info("SubscriptionManager - Removed subscriber", zap.String("key", key))
	}()

	return ch, nil
}

func (manager *SubscriptionManager[T]) RemoveSubscriber(key string) {
	if ch, ok := manager.subscribers[key]; ok {
		if ch != nil {
			close(ch)
		}

		delete(manager.subscribers, key)
	}
}

func (manager *SubscriptionManager[T]) Close() error {
	manager.closed = true

	for key := range manager.subscribers {
		manager.RemoveSubscriber(key)
	}

	close(manager.input)

	return nil
}