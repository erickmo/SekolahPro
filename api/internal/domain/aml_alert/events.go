package aml_alert

import "github.com/google/uuid"

// AlertCreatedEvent dipublikasikan saat transaction alert baru berhasil dibuat.
type AlertCreatedEvent struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	CompanyID  uuid.UUID `json:"company_id"`
	ID         uuid.UUID `json:"id"`
	NasabahID  uuid.UUID `json:"nasabah_id"`
	RuleID     uuid.UUID `json:"rule_id"`
	AlertType  string    `json:"alert_type"`
	Severity   string    `json:"severity"`
}

func (e AlertCreatedEvent) EventName() string   { return "aml_alert.created" }
func (e AlertCreatedEvent) AggregateID() string { return e.ID.String() }

// AlertInvestigatedEvent dipublikasikan saat alert mulai diinvestigasi.
type AlertInvestigatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
}

func (e AlertInvestigatedEvent) EventName() string   { return "aml_alert.investigated" }
func (e AlertInvestigatedEvent) AggregateID() string { return e.ID.String() }

// AlertResolvedEvent dipublikasikan saat alert berhasil diresolusi.
type AlertResolvedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	Outcome   string    `json:"outcome"`
}

func (e AlertResolvedEvent) EventName() string   { return "aml_alert.resolved" }
func (e AlertResolvedEvent) AggregateID() string { return e.ID.String() }
