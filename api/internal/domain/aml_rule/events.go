package aml_rule

import "github.com/google/uuid"

// RuleCreatedEvent dipublikasikan saat monitoring rule baru berhasil dibuat.
type RuleCreatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	RuleCode  string    `json:"rule_code"`
	RuleType  string    `json:"rule_type"`
}

func (e RuleCreatedEvent) EventName() string   { return "aml_rule.created" }
func (e RuleCreatedEvent) AggregateID() string { return e.ID.String() }

// RuleUpdatedEvent dipublikasikan saat monitoring rule berhasil diupdate.
type RuleUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
}

func (e RuleUpdatedEvent) EventName() string   { return "aml_rule.updated" }
func (e RuleUpdatedEvent) AggregateID() string { return e.ID.String() }

// RuleToggledEvent dipublikasikan saat monitoring rule diaktifkan/dinonaktifkan.
type RuleToggledEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
}

func (e RuleToggledEvent) EventName() string   { return "aml_rule.toggled" }
func (e RuleToggledEvent) AggregateID() string { return e.ID.String() }
