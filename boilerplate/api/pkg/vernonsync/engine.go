// Package vernonsync menyediakan SyncEngine untuk Vernon denormalized read-cache.
//
// Flow:
//  1. BaseService publish SyncEvent via EventBus setelah TX commit.
//  2. SyncEngine subscribe ke topic "sync.*" via Watermill.
//  3. Worker mencari consumer domains dari Registry.reverseMap.
//  4. Untuk setiap consumer, update _data dengan embedded data terbaru (CAS).
//
// SyncEngine dipisah dari pkg/vernon (core) agar dependency pgx/DB
// tidak mengkontaminasi domain yang tidak butuh sync.
package vernonsync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
	"github.com/yourorg/boilerplate/pkg/vernon"
)

// SyncEngine subscribe ke SyncEvent dari EventBus dan propagate perubahan
// ke semua consumer domain yang punya autoload relationship.
type SyncEngine struct {
	db       *sqlx.DB
	registry *vernon.Registry
	eb       eventbus.EventBus
	log      zerolog.Logger
}

// NewSyncEngine membuat SyncEngine baru.
func NewSyncEngine(db *sqlx.DB, registry *vernon.Registry, eb eventbus.EventBus, log zerolog.Logger) *SyncEngine {
	return &SyncEngine{db: db, registry: registry, eb: eb, log: log}
}

// Subscribe mendaftarkan handler untuk semua domain yang sudah registered.
// Harus dipanggil setelah semua domain di-register ke Registry,
// dan sebelum EventBus.StartRouter().
func (e *SyncEngine) Subscribe(ctx context.Context) error {
	// Collect semua domain names dari registry untuk di-subscribe.
	// Karena registry.domains bersifat private, kita iterate melalui consumers
	// dan juga subscribe ke semua domain yang MENJADI sumber perubahan.
	domains := e.collectSourceDomains()
	for _, domain := range domains {
		domainName := domain
		topic := "sync." + domainName
		if err := e.eb.Subscribe(ctx, topic, func(ctx context.Context, msg *message.Message) error {
			return e.handleSyncEvent(ctx, msg, domainName)
		}); err != nil {
			return fmt.Errorf("subscribe sync.%s: %w", domainName, err)
		}
		e.log.Debug().Str("topic", topic).Msg("vernonsync subscribed")
	}
	return nil
}

// handleSyncEvent memproses satu SyncEvent dari EventBus.
func (e *SyncEngine) handleSyncEvent(ctx context.Context, msg *message.Message, _ string) error {
	var event vernon.SyncEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		e.log.Error().Err(err).Msg("invalid sync event payload")
		msg.Ack()
		return nil
	}

	sc := scope.Scope{TenantID: event.TenantID, CompanyID: event.CompanyID}
	consumers := e.registry.GetConsumers(event.Domain)
	if len(consumers) == 0 {
		msg.Ack()
		return nil
	}

	visited := map[string]bool{event.Domain + ":" + event.EntityID: true}
	if err := e.propagate(ctx, event, sc, consumers, visited); err != nil {
		e.log.Error().Err(err).
			Str("domain", event.Domain).Str("id", event.EntityID).
			Msg("sync propagation failed")
		msg.Nack()
		return err
	}

	msg.Ack()
	return nil
}

// propagate memperbarui _data di semua consumer domains yang terpengaruh.
func (e *SyncEngine) propagate(ctx context.Context, event vernon.SyncEvent, sc scope.Scope, consumers []vernon.ConsumerRef, visited map[string]bool) error {
	sourceData, err := e.fetchSourceData(ctx, sc, event.Domain, event.EntityID)
	if err != nil {
		return fmt.Errorf("fetch source %s/%s: %w", event.Domain, event.EntityID, err)
	}

	for _, ref := range consumers {
		if err := e.propagateToConsumer(ctx, sc, ref, event, sourceData, visited); err != nil {
			e.log.Warn().Err(err).
				Str("consumer", ref.ConsumerDomain).Str("rel", ref.RelName).
				Msg("consumer propagation failed, continuing")
		}
	}
	return nil
}

// propagateToConsumer memperbarui _data pada satu consumer domain.
func (e *SyncEngine) propagateToConsumer(ctx context.Context, sc scope.Scope, ref vernon.ConsumerRef, event vernon.SyncEvent, sourceData map[string]any, visited map[string]bool) error {
	embedded := extractFields(sourceData, ref.RelDef.Fields)

	switch ref.RelDef.Type {
	case vernon.RelBelongsTo:
		return e.propagateBelongsTo(ctx, sc, ref, event.EntityID, embedded, visited)
	case vernon.RelHasMany:
		return e.propagateHasMany(ctx, sc, ref, event.EntityID, sourceData, visited)
	default:
		// many_to_many: TODO — on-demand only, skip autoload propagation
		return nil
	}
}

// propagateBelongsTo update consumer dimana _data->>'fk' = entityID.
func (e *SyncEngine) propagateBelongsTo(ctx context.Context, sc scope.Scope, ref vernon.ConsumerRef, entityID string, embedded map[string]any, visited map[string]bool) error {
	consumerRepo := vernon.NewBaseRepository(e.db, ref.ConsumerDomain)
	consumers, err := consumerRepo.FindByDataField(ctx, sc, ref.RelDef.FK, entityID)
	if err != nil {
		return err
	}
	for _, c := range consumers {
		key := ref.ConsumerDomain + ":" + c.ID.String()
		if visited[key] {
			continue
		}
		visited[key] = true
		if casErr := consumerRepo.CASUpdate(ctx, sc, c.ID, c.SyncVersion, ref.RelName, embedded); casErr != nil {
			e.log.Warn().Err(casErr).Str("id", c.ID.String()).Msg("CAS conflict, skipping")
		}
	}
	return nil
}

// propagateHasMany update consumer dimana consumer.id = _data->>'parentFK'.
// Contoh: order_item berubah → cari order_id dari item → update orders.items array.
func (e *SyncEngine) propagateHasMany(ctx context.Context, sc scope.Scope, ref vernon.ConsumerRef, _ string, sourceData map[string]any, visited map[string]bool) error {
	parentIDVal, ok := sourceData[ref.RelDef.FK]
	if !ok {
		return nil
	}
	parentID, err := uuid.Parse(fmt.Sprintf("%v", parentIDVal))
	if err != nil {
		return nil
	}

	consumerRepo := vernon.NewBaseRepository(e.db, ref.ConsumerDomain)
	parent, err := consumerRepo.FindByID(ctx, sc, parentID)
	if err != nil {
		return err
	}

	childRepo := vernon.NewBaseRepository(e.db, ref.RelDef.Domain)
	children, _, err := childRepo.FindAll(ctx, sc, vernon.FindParams{
		Filters: map[string]string{ref.RelDef.FK: parentID.String()},
		Limit:   ref.RelDef.MaxItems,
		Page:    1,
	})
	if err != nil {
		return err
	}

	items := make([]map[string]any, 0, len(children))
	for _, child := range children {
		items = append(items, extractFields(child.Data, ref.RelDef.Fields))
	}

	key := ref.ConsumerDomain + ":" + parentID.String()
	if visited[key] {
		return nil
	}
	visited[key] = true

	return consumerRepo.CASUpdate(ctx, sc, parent.ID, parent.SyncVersion, ref.RelName, items)
}

// fetchSourceData mengambil _data dari entitas sumber.
func (e *SyncEngine) fetchSourceData(ctx context.Context, sc scope.Scope, domain, entityID string) (map[string]any, error) {
	id, err := uuid.Parse(entityID)
	if err != nil {
		return nil, fmt.Errorf("parse entity ID: %w", err)
	}
	repo := vernon.NewBaseRepository(e.db, domain)
	entity, err := repo.FindByID(ctx, sc, id)
	if err != nil {
		return nil, err
	}
	return entity.Data, nil
}

// collectSourceDomains mengumpulkan nama semua domain yang menjadi sumber sync.
// Ini adalah semua domain yang ada di reverseMap (sebagai target dari relasi autoload).
func (e *SyncEngine) collectSourceDomains() []string {
	// Kita subscribe ke topic "sync.{domain}" untuk semua domain yang punya consumers.
	// SyncEngine hanya perlu subscribe ke domain yang ADA di reverseMap.
	// Untuk mengakses reverseMap, kita ambil melalui GetConsumers — tapi kita butuh
	// semua source domain names.
	//
	// Solusi pragmatis: gunakan getSourceDomains helper yang di-expose Registry.
	return e.registry.GetSourceDomains()
}

func extractFields(data map[string]any, fields []string) map[string]any {
	if len(fields) == 0 {
		return data
	}
	result := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := data[f]; ok {
			result[f] = v
		}
	}
	return result
}
