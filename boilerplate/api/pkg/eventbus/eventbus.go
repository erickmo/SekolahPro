package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// DomainEvent adalah interface yang harus diimplementasi oleh semua domain event.
type DomainEvent interface {
	EventName() string
	AggregateID() string
}

// EventBus mendefinisikan kontrak untuk publish dan subscribe event.
type EventBus interface {
	Publish(ctx context.Context, event DomainEvent) error
	Subscribe(ctx context.Context, eventName string, handler func(ctx context.Context, event *message.Message) error) error
}

// RouterRunner diimplementasi oleh EventBus yang butuh lifecycle management eksplisit.
type RouterRunner interface {
	StartRouter(ctx context.Context) error
	StopRouter(ctx context.Context) error
	RouterRunning() <-chan struct{}
}

type watermillEventBus struct {
	publisher    message.Publisher
	subscriber   message.Subscriber
	router       *message.Router
	mu           sync.Mutex
	handlerCount map[string]int
}

// NewInMemoryEventBus membuat in-memory event bus (untuk development / testing).
func NewInMemoryEventBus() (EventBus, error) {
	logger := watermill.NewStdLogger(false, false)
	ch := gochannel.NewGoChannel(gochannel.Config{
		OutputChannelBuffer:            100,
		Persistent:                     false,
		BlockPublishUntilSubscriberAck: false,
	}, logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("create watermill router: %w", err)
	}

	return &watermillEventBus{
		publisher:    ch,
		subscriber:   ch,
		router:       router,
		handlerCount: make(map[string]int),
	}, nil
}

func (b *watermillEventBus) Publish(ctx context.Context, event DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event %s: %w", event.EventName(), err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)

	return b.publisher.Publish(event.EventName(), msg)
}

func (b *watermillEventBus) Subscribe(ctx context.Context, eventName string, handler func(ctx context.Context, event *message.Message) error) error {
	b.mu.Lock()
	b.handlerCount[eventName]++
	count := b.handlerCount[eventName]
	b.mu.Unlock()

	handlerName := fmt.Sprintf("handler_%s", eventName)
	if count > 1 {
		handlerName = fmt.Sprintf("handler_%s_%d", eventName, count)
	}

	b.router.AddNoPublisherHandler(
		handlerName,
		eventName,
		b.subscriber,
		func(msg *message.Message) error {
			return handler(msg.Context(), msg)
		},
	)
	return nil
}

// StartRouter memulai watermill router. Harus dipanggil setelah semua Subscribe.
func (b *watermillEventBus) StartRouter(ctx context.Context) error {
	return b.router.Run(ctx)
}

// StopRouter menutup watermill router secara graceful.
func (b *watermillEventBus) StopRouter(ctx context.Context) error {
	done := make(chan struct{}, 1)
	go func() {
		b.router.Close() //nolint:errcheck
		done <- struct{}{}
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return nil
	}
}

// RouterRunning mengembalikan channel yang ditutup saat router siap memproses pesan.
func (b *watermillEventBus) RouterRunning() <-chan struct{} {
	return b.router.Running()
}
