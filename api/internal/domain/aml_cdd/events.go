package aml_cdd

import "github.com/google/uuid"

// CDDCreatedEvent dipublikasikan saat CDD record baru berhasil dibuat.
type CDDCreatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	NasabahID uuid.UUID `json:"nasabah_id"`
	RiskLevel string    `json:"risk_level"`
	CDDLevel  string    `json:"cdd_level"`
}

func (e CDDCreatedEvent) EventName() string   { return "aml_cdd.created" }
func (e CDDCreatedEvent) AggregateID() string { return e.ID.String() }

// CDDUpdatedEvent dipublikasikan saat CDD record berhasil diupdate.
type CDDUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
}

func (e CDDUpdatedEvent) EventName() string   { return "aml_cdd.updated" }
func (e CDDUpdatedEvent) AggregateID() string { return e.ID.String() }

// CDDRiskUpdatedEvent dipublikasikan saat risk assessment CDD berubah.
type CDDRiskUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	OldRisk   string    `json:"old_risk"`
	NewRisk   string    `json:"new_risk"`
}

func (e CDDRiskUpdatedEvent) EventName() string   { return "aml_cdd.risk_updated" }
func (e CDDRiskUpdatedEvent) AggregateID() string { return e.ID.String() }
