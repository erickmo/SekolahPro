// Package vernon menyediakan Vernon denormalized read-cache architecture.
//
// Setiap domain table menyimpan data bisnis di dalam kolom JSONB _data,
// sehingga GET endpoint tidak membutuhkan JOIN — langsung baca _data.
// Konsistensi dijaga secara asinkron via SyncEngine (pkg/vernonsync).
//
// Pola Hybrid:
//   - Domain sederhana (sedikit relasi) → tetap pakai CQRS + sqlc
//   - Domain berat (banyak JOIN)        → Vernon _rels/_data pattern
package vernon

import (
	"time"

	"github.com/google/uuid"
)

// Sync status constants.
const (
	SyncStatusSynced  = "synced"
	SyncStatusPending = "pending"
	SyncStatusError   = "error"
)

// Relationship type constants.
const (
	RelBelongsTo   = "belongs_to"
	RelHasOne      = "has_one"
	RelHasMany     = "has_many"
	RelManyToMany  = "many_to_many"
)

// Action constants untuk SyncEvent.
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionPatch  = "patch"
	ActionDelete = "delete"
)

// RelDef mendefinisikan satu relasi dalam _rels.
// Disimpan di database sehingga SyncEngine bisa membacanya tanpa registry.
type RelDef struct {
	Domain      string   `json:"domain"`       // target table, e.g. "customers"
	Type        string   `json:"type"`         // belongs_to | has_one | has_many | many_to_many
	FK          string   `json:"fk"`           // foreign key field name
	LocalKey    string   `json:"local_key"`    // key pada entitas ini
	ForeignKey  string   `json:"foreign_key"`  // key pada entitas target
	IsAutoload  bool     `json:"is_autoload"`  // true = embed di _data saat write
	Fields      []string `json:"fields"`       // field yang di-denormalisasi
	MaxItems    int      `json:"max_items,omitempty"` // cap untuk array (has_many/m2m)

	// Many-to-many
	Junction           string `json:"junction,omitempty"`
	JunctionLocalKey   string `json:"junction_local_key,omitempty"`
	JunctionForeignKey string `json:"junction_foreign_key,omitempty"`
}

// BaseDomain adalah struct induk untuk semua domain Vernon.
// Business fields disimpan di dalam Data (kolom _data JSONB).
type BaseDomain struct {
	ID          uuid.UUID              `json:"id"           db:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"    db:"tenant_id"`
	CompanyID   uuid.UUID              `json:"company_id"   db:"company_id"`
	Rels        map[string]RelDef      `json:"_rels"        db:"_rels"`
	Data        map[string]any         `json:"_data"        db:"_data"`
	SyncStatus  string                 `json:"_sync_status" db:"_sync_status"`
	SyncVersion int64                  `json:"_sync_version" db:"_sync_version"`
	CreatedAt   time.Time              `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"   db:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SetDataField menyimpan satu field ke dalam _data.
func (b *BaseDomain) SetDataField(key string, value any) {
	if b.Data == nil {
		b.Data = make(map[string]any)
	}
	b.Data[key] = value
}

// GetDataField mengambil satu field dari _data.
func (b *BaseDomain) GetDataField(key string) (any, bool) {
	v, ok := b.Data[key]
	return v, ok
}

// DomainDescriptor mendefinisikan metadata sebuah domain Vernon.
// Setiap domain concrete hanya perlu mengimplementasi interface ini.
// CRUD, sync, dan query parsing diwariskan dari base layer.
type DomainDescriptor interface {
	TableName() string
	DefaultRels() map[string]RelDef
	Validate(data map[string]any) error
}

// SyncEvent dipublish setelah setiap write operation.
// SyncEngine (pkg/vernonsync) subscribe ke event ini dan propagate
// ke consumer domains yang punya autoload relationship.
//
// Mengimplementasi eventbus.DomainEvent secara implisit (duck typing).
type SyncEvent struct {
	Domain        string    `json:"domain"`
	EntityID      string    `json:"entity_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	CompanyID     uuid.UUID `json:"company_id"`
	Action        string    `json:"action"`
	ChangedFields []string  `json:"changed_fields,omitempty"`
	SyncVersion   int64     `json:"sync_version"`
}

// EventName mengimplementasi eventbus.DomainEvent.
// Format: "sync.{domain}" — e.g. "sync.customers".
func (e SyncEvent) EventName() string { return "sync." + e.Domain }

// AggregateID mengimplementasi eventbus.DomainEvent.
func (e SyncEvent) AggregateID() string { return e.EntityID }

// FindParams berisi parameter untuk operasi list/find.
type FindParams struct {
	Filters map[string]string // field -> value (mendukung dot notation)
	Sort    string            // prefix "-" untuk DESC, e.g. "-created_at"
	Page    int
	Limit   int
}
