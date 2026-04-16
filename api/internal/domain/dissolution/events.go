package dissolution

import "github.com/google/uuid"

// DissolutionInitiatedEvent dipublikasikan saat proses pembubaran baru berhasil dibuat.
// TenantID dan CompanyID wajib ada agar event handler bisa memproses dengan isolasi yang benar.
type DissolutionInitiatedEvent struct {
	TenantID        uuid.UUID       `json:"tenant_id"`
	CompanyID       uuid.UUID       `json:"company_id"`
	ID              uuid.UUID       `json:"id"`
	DissolutionType DissolutionType `json:"dissolution_type"`
	Stage           Stage           `json:"stage"`
	Status          Status          `json:"status"`
}

func (e DissolutionInitiatedEvent) EventName() string   { return "dissolution.initiated" }
func (e DissolutionInitiatedEvent) AggregateID() string { return e.ID.String() }

// DissolutionStageChangedEvent dipublikasikan saat tahapan pembubaran berubah.
type DissolutionStageChangedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	OldStage  Stage     `json:"old_stage"`
	NewStage  Stage     `json:"new_stage"`
}

func (e DissolutionStageChangedEvent) EventName() string   { return "dissolution.stage_changed" }
func (e DissolutionStageChangedEvent) AggregateID() string { return e.ID.String() }

// DissolutionCompletedEvent dipublikasikan saat proses pembubaran selesai.
type DissolutionCompletedEvent struct {
	TenantID              uuid.UUID `json:"tenant_id"`
	CompanyID             uuid.UUID `json:"company_id"`
	ID                    uuid.UUID `json:"id"`
	NetEquity             int64     `json:"net_equity"`
	DistributionPerMember int64     `json:"distribution_per_member"`
}

func (e DissolutionCompletedEvent) EventName() string   { return "dissolution.completed" }
func (e DissolutionCompletedEvent) AggregateID() string { return e.ID.String() }
