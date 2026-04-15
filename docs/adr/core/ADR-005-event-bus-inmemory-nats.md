# ADR-005: Event Bus Abstraction — InMemory (Dev) / NATS JetStream (Prod)

**Status**: Accepted
**Date**: 2026-04-14
**Deciders**: Vernon Corp Engineering Team

## Context

Sistem menggunakan event-driven architecture untuk beberapa use case kritis:

1. **Vernon SyncEngine**: Ketika category di-update, `CategoryUpdatedEvent` harus men-trigger update `_data` di semua products yang mereferensikannya.
2. **Domain Events**: `OrderCreated` → inventory reservation, `PaymentReceived` → order fulfillment, `StockDepleted` → purchase order suggestion.
3. **Audit Trail**: Setiap command penting diterbitkan sebagai event untuk audit log.
4. **Cross-domain Communication**: Domain yang berbeda berkomunikasi via events, bukan direct method call (loose coupling).

### Masalah yang Perlu Diselesaikan

**Developer Experience**: Menjalankan NATS JetStream di local development membutuhkan Docker, konfigurasi tambahan, dan resources. Ini meningkatkan friction onboarding dan memperlambat feedback loop saat development.

**Test Isolation**: Integration test yang bergantung pada NATS nyata sulit untuk dijalankan di CI tanpa infrastruktur tambahan, dan rentan terhadap flakiness (timing issues).

**Production Requirements**: Di production, event bus harus:
- Durabel: event tidak hilang jika consumer restart
- At-least-once delivery dengan deduplikasi
- Ordered delivery per subject
- Replay capability untuk recovery
- Horizontal scaling consumer

**Abstraction Need**: Application code tidak boleh tightly coupled ke implementasi event bus spesifik — harus bisa switch tanpa mengubah domain atau use case layer.

## Decision

Mengadopsi **event bus abstraction** dengan dua implementasi yang dapat di-swap via konfigurasi:

### Interface Definition

```go
// domain/port/eventbus.go
type Event struct {
    ID          string
    Type        string
    AggregateID string
    OccurredAt  time.Time
    Payload     []byte  // JSON-encoded
}

type Handler func(ctx context.Context, event Event) error

type EventBus interface {
    // Publish menerbitkan event secara async
    Publish(ctx context.Context, subject string, event Event) error
    
    // Subscribe mendaftarkan handler untuk subject tertentu
    Subscribe(subject string, handler Handler) error
    
    // Close menutup koneksi dengan graceful
    Close() error
}
```

### Implementasi 1: InMemory (Development & Test)

```go
// infrastructure/messaging/inmemory/bus.go
type InMemoryBus struct {
    mu       sync.RWMutex
    handlers map[string][]Handler
    logger   zerolog.Logger
}

func (b *InMemoryBus) Publish(ctx context.Context, subject string, event Event) error {
    b.mu.RLock()
    handlers := b.handlers[subject]
    b.mu.RUnlock()
    
    for _, h := range handlers {
        // Dispatch synchronously dalam goroutine terpisah
        go func(handler Handler) {
            if err := handler(ctx, event); err != nil {
                b.logger.Error().Err(err).Str("subject", subject).Msg("event handler error")
            }
        }(h)
    }
    return nil
}
```

Karakteristik InMemory:
- In-process, zero network latency
- Events hilang jika process restart (acceptable di dev/test)
- Synchronous publish, async dispatch
- No configuration required — works out of the box

### Implementasi 2: NATS JetStream (Production)

```go
// infrastructure/messaging/nats/bus.go
type NATSBus struct {
    nc     *nats.Conn
    js     nats.JetStreamContext
    logger zerolog.Logger
}

func (b *NATSBus) Publish(ctx context.Context, subject string, event Event) error {
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal event: %w", err)
    }
    
    msg := &nats.Msg{
        Subject: subject,
        Data:    data,
        Header:  nats.Header{},
    }
    msg.Header.Set("Event-ID", event.ID)     // deduplikasi
    msg.Header.Set("Event-Type", event.Type)
    
    ack, err := b.js.PublishMsg(msg, nats.MsgId(event.ID))  // idempotency key
    if err != nil {
        return fmt.Errorf("publish to NATS: %w", err)
    }
    
    b.logger.Debug().
        Str("subject", subject).
        Uint64("sequence", ack.Sequence).
        Msg("event published")
    return nil
}

func (b *NATSBus) Subscribe(subject string, handler Handler) error {
    _, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
        var event Event
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            msg.Nak()
            return
        }
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := handler(ctx, event); err != nil {
            msg.NakWithDelay(5 * time.Second)  // retry dengan backoff
            return
        }
        
        msg.Ack()
    }, nats.Durable("boilerplate-"+subject), nats.AckExplicit())
    
    return err
}
```

### JetStream Stream Configuration

```go
// infrastructure/messaging/nats/setup.go
func SetupStreams(js nats.JetStreamContext) error {
    streams := []nats.StreamConfig{
        {
            Name:       "DOMAIN_EVENTS",
            Subjects:   []string{"events.>"},  // wildcard — semua domain events
            Storage:    nats.FileStorage,       // persistent
            Retention:  nats.LimitsPolicy,
            MaxAge:     7 * 24 * time.Hour,    // retain 7 hari
            MaxMsgs:    10_000_000,
            Replicas:   3,                      // HA di production
            Duplicates: 1 * time.Minute,        // deduplikasi window
        },
        {
            Name:      "SYNC_ENGINE",
            Subjects:  []string{"sync.>"},     // Vernon SyncEngine events
            Storage:   nats.FileStorage,
            MaxAge:    24 * time.Hour,
            Replicas:  3,
        },
    }
    
    for _, cfg := range streams {
        if _, err := js.AddStream(&cfg); err != nil {
            // UpdateStream jika sudah ada
            if _, err := js.UpdateStream(&cfg); err != nil {
                return fmt.Errorf("setup stream %s: %w", cfg.Name, err)
            }
        }
    }
    return nil
}
```

### Configuration Switch

```go
// infrastructure/fx/messaging_module.go
var MessagingModule = fx.Module("messaging",
    fx.Provide(func(cfg *Config, logger zerolog.Logger) (EventBus, error) {
        if cfg.UseNATS {
            return nats.NewBus(cfg.NATSUrl, logger)
        }
        return inmemory.NewBus(logger), nil
    }),
)
```

Environment variables:
```bash
# .env.development
USE_NATS=false

# .env.production
USE_NATS=true
NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
```

### Subject Naming Convention

```
events.{domain}.{event_type}

Contoh:
events.products.product_created
events.products.product_updated
events.categories.category_updated
events.orders.order_placed
events.inventory.stock_depleted

sync.products.category_updated  (← SyncEngine subjects)
sync.products.supplier_updated
sync.inventory.warehouse_updated
```

## Consequences

### Positive

- **Zero-friction dev setup**: `USE_NATS=false` → tidak perlu Docker NATS untuk development. `go run ./cmd/api` langsung jalan.
- **Fast tests**: Unit dan integration test menggunakan InMemory — tidak ada network call, tidak ada timing issues.
- **Production-grade delivery**: NATS JetStream memberikan at-least-once delivery, persistence, dan replay yang dibutuhkan production.
- **Interface stability**: Application code hanya bergantung pada `EventBus` interface — pergantian implementasi tidak menyentuh domain atau use case layer.
- **Deduplikasi built-in**: `nats.MsgId(event.ID)` di JetStream memastikan event idempoten meskipun di-publish ulang karena retry.
- **Replay capability**: Jika SyncEngine crash, bisa replay events dari NATS untuk re-populate `_data` yang stale.

### Negative / Trade-offs

- **Behavioral difference**: InMemory dispatch berjalan dalam goroutine dengan no guarantee ordering; NATS JetStream memberikan ordered delivery per subject. Test yang bergantung pada ordering harus dijalankan dengan NATS nyata.
- **In-memory tidak durabel**: Jika process crash saat development dan ada event yang belum diproses, event tersebut hilang. Acceptable di dev, tidak acceptable di production.
- **NATS operasional overhead**: Mengelola NATS cluster di production membutuhkan keahlian operasional — monitoring, backup, upgrade.
- **Eventual consistency tetap ada**: Bahkan dengan NATS JetStream, consumer bisa lagging — SyncEngine tidak menjamin immediate consistency.
- **Subject naming harus konsisten**: Typo di subject name menyebabkan event tidak terdeliver ke consumer yang tepat — perlu constants, bukan string literal.

## Alternatives Considered

### 1. Kafka
- Industry standard, sangat scalable, strong ordering guarantees
- Ditolak karena operational overhead jauh lebih tinggi dari NATS (ZooKeeper/KRaft dependency, JVM-based), dan fitur yang dibutuhkan sepenuhnya tersedia di NATS JetStream dengan overhead lebih rendah.

### 2. RabbitMQ
- Mature, flexible routing via exchanges
- Ditolak karena tidak memiliki built-in persistence/streaming yang sekuat NATS JetStream, dan AMQP protocol lebih kompleks dari NATS protocol.

### 3. Redis Streams
- Sudah ada di stack (Redis untuk caching)
- Ditolak karena Redis Streams tidak didesain sebagai primary message broker — kurang reliable untuk high-throughput, dan menambah beban ke Redis instance yang sama dengan caching.

### 4. Google Cloud Pub/Sub / AWS SNS+SQS
- Managed, zero operational overhead
- Ditolak karena cloud vendor lock-in, tidak cocok untuk self-hosted deployment, dan menambah cost untuk development/staging.

### 5. Direct HTTP calls antar service (synchronous)
- Paling sederhana, tidak perlu message broker
- Ditolak karena tight coupling, cascade failure risk (jika consumer down maka publisher harus handle error/retry), dan tidak scalable untuk event fan-out.
