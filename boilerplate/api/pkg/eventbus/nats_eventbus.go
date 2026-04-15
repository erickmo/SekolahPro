package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/jetstream"
	"github.com/ThreeDotsLabs/watermill/message"
	natsio "github.com/nats-io/nats.go"
)

// NATSConfig berisi konfigurasi koneksi NATS JetStream.
type NATSConfig struct {
	URL        string
	StreamName string // Nama JetStream stream, e.g. "APP_EVENTS"
}

type natsEventBus struct {
	publisher    message.Publisher
	subscriber   message.Subscriber
	router       *message.Router
	mu           sync.Mutex
	handlerCount map[string]int
}

// NewNATSEventBus membuat EventBus berbasis NATS JetStream.
// Digunakan di production saat USE_NATS=true.
// JetStream memberikan at-least-once delivery dengan persistent storage.
func NewNATSEventBus(cfg NATSConfig) (EventBus, error) {
	logger := watermill.NewStdLogger(false, false)

	nc, err := natsio.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to nats %s: %w", cfg.URL, err)
	}

	pub, err := jetstream.NewPublisher(jetstream.PublisherConfig{
		Conn:   nc,
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("create nats publisher: %w", err)
	}

	sub, err := jetstream.NewSubscriber(jetstream.SubscriberConfig{
		Conn:   nc,
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("create nats subscriber: %w", err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("create watermill router: %w", err)
	}

	return &natsEventBus{
		publisher:    pub,
		subscriber:   sub,
		router:       router,
		handlerCount: make(map[string]int),
	}, nil
}

func (b *natsEventBus) Publish(ctx context.Context, event DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event %s: %w", event.EventName(), err)
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.SetContext(ctx)
	return b.publisher.Publish(event.EventName(), msg)
}

func (b *natsEventBus) Subscribe(_ context.Context, eventName string, handler func(ctx context.Context, event *message.Message) error) error {
	b.mu.Lock()
	b.handlerCount[eventName]++
	count := b.handlerCount[eventName]
	b.mu.Unlock()

	handlerName := fmt.Sprintf("nats_handler_%s", eventName)
	if count > 1 {
		handlerName = fmt.Sprintf("nats_handler_%s_%d", eventName, count)
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

func (b *natsEventBus) StartRouter(ctx context.Context) error {
	return b.router.Run(ctx)
}

func (b *natsEventBus) StopRouter(ctx context.Context) error {
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

func (b *natsEventBus) RouterRunning() <-chan struct{} {
	return b.router.Running()
}
