package governance

import "github.com/google/uuid"

// GovernancePositionCreatedEvent dipublikasikan saat posisi governance baru berhasil dibuat.
type GovernancePositionCreatedEvent struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	CompanyID     uuid.UUID `json:"company_id"`
	ID            uuid.UUID `json:"id"`
	NasabahID     uuid.UUID `json:"nasabah_id"`
	PositionType  string    `json:"position_type"`
	PositionLevel string    `json:"position_level"`
}

func (e GovernancePositionCreatedEvent) EventName() string   { return "governance_position.created" }
func (e GovernancePositionCreatedEvent) AggregateID() string { return e.ID.String() }

// GovernancePositionUpdatedEvent dipublikasikan saat posisi governance berhasil diupdate.
type GovernancePositionUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
}

func (e GovernancePositionUpdatedEvent) EventName() string   { return "governance_position.updated" }
func (e GovernancePositionUpdatedEvent) AggregateID() string { return e.ID.String() }

// GovernanceStatusChangedEvent dipublikasikan saat status posisi governance berubah.
type GovernanceStatusChangedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
}

func (e GovernanceStatusChangedEvent) EventName() string   { return "governance_position.status_changed" }
func (e GovernanceStatusChangedEvent) AggregateID() string { return e.ID.String() }
