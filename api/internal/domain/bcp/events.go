package bcp

import "github.com/google/uuid"

// BackupStrategyCreatedEvent dipublikasikan saat BackupStrategy baru berhasil dibuat.
// TenantID dan CompanyID wajib ada agar event handler bisa memproses dengan isolasi yang benar.
type BackupStrategyCreatedEvent struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	CompanyID     uuid.UUID `json:"company_id"`
	ID            uuid.UUID `json:"id"`
	StrategyType  string    `json:"strategy_type"`
	Frequency     string    `json:"frequency"`
}

func (e BackupStrategyCreatedEvent) EventName() string   { return "backup_strategy.created" }
func (e BackupStrategyCreatedEvent) AggregateID() string { return e.ID.String() }

// BackupStrategyUpdatedEvent dipublikasikan saat BackupStrategy berhasil diupdate.
type BackupStrategyUpdatedEvent struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	CompanyID     uuid.UUID `json:"company_id"`
	ID            uuid.UUID `json:"id"`
	StrategyType  string    `json:"strategy_type"`
	Frequency     string    `json:"frequency"`
}

func (e BackupStrategyUpdatedEvent) EventName() string   { return "backup_strategy.updated" }
func (e BackupStrategyUpdatedEvent) AggregateID() string { return e.ID.String() }
