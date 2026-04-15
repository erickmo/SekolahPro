// Package eventhandler berisi event handler lintas domain.
// Satu file per kombinasi source → target domain.
package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog"

	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/eventbus"
)

// ExampleEventHandler menangani side effect dari domain events Example.
// Side effect: logging + audit trail. Tambahkan notifikasi, cache invalidation, dll sesuai kebutuhan.
type ExampleEventHandler struct {
	log zerolog.Logger
}

// NewExampleEventHandler membuat instance ExampleEventHandler baru.
func NewExampleEventHandler(log zerolog.Logger) *ExampleEventHandler {
	return &ExampleEventHandler{log: log}
}

// RegisterHandlers mendaftarkan semua event handler ke event bus.
func (h *ExampleEventHandler) RegisterHandlers(eb eventbus.EventBus) error {
	if err := eb.Subscribe(context.Background(), "example.created", h.onExampleCreated); err != nil {
		return fmt.Errorf("subscribe example.created: %w", err)
	}
	if err := eb.Subscribe(context.Background(), "example.updated", h.onExampleUpdated); err != nil {
		return fmt.Errorf("subscribe example.updated: %w", err)
	}
	return nil
}

func (h *ExampleEventHandler) onExampleCreated(ctx context.Context, msg *message.Message) error {
	var event example.ExampleCreatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return fmt.Errorf("unmarshal ExampleCreatedEvent: %w", err)
	}

	h.log.Info().
		Str("event", event.EventName()).
		Str("tenant_id", event.TenantID.String()).
		Str("company_id", event.CompanyID.String()).
		Str("id", event.AggregateID()).
		Str("name", event.Name).
		Msg("example dibuat — side effects diproses")

	// Side effects yang bisa ditambahkan:
	// - Kirim notifikasi ke user (email/push)
	// - Invalidate cache list examples untuk tenant ini
	// - Update search index (Elasticsearch/Typesense)
	// - Kirim audit log ke audit service
	// - Trigger workflow eksternal (Zapier, webhook)

	return nil
}

func (h *ExampleEventHandler) onExampleUpdated(ctx context.Context, msg *message.Message) error {
	var event example.ExampleUpdatedEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return fmt.Errorf("unmarshal ExampleUpdatedEvent: %w", err)
	}

	h.log.Info().
		Str("event", event.EventName()).
		Str("tenant_id", event.TenantID.String()).
		Str("id", event.AggregateID()).
		Str("name", event.Name).
		Msg("example diupdate — side effects diproses")

	return nil
}
