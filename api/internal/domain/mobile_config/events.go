package mobile_config

import "github.com/google/uuid"

// MobileConfigCreatedEvent dipublikasikan saat MobileAppConfig baru berhasil dibuat.
type MobileConfigCreatedEvent struct {
	TenantID   uuid.UUID  `json:"tenant_id"`
	CompanyID  uuid.UUID  `json:"company_id"`
	ID         uuid.UUID  `json:"id"`
	AppVariant AppVariant `json:"app_variant"`
	Platform   Platform   `json:"platform"`
}

func (e MobileConfigCreatedEvent) EventName() string   { return "mobile_config.created" }
func (e MobileConfigCreatedEvent) AggregateID() string { return e.ID.String() }

// MobileConfigUpdatedEvent dipublikasikan saat MobileAppConfig berhasil diupdate.
type MobileConfigUpdatedEvent struct {
	TenantID   uuid.UUID  `json:"tenant_id"`
	CompanyID  uuid.UUID  `json:"company_id"`
	ID         uuid.UUID  `json:"id"`
	AppVariant AppVariant `json:"app_variant"`
	Platform   Platform   `json:"platform"`
}

func (e MobileConfigUpdatedEvent) EventName() string   { return "mobile_config.updated" }
func (e MobileConfigUpdatedEvent) AggregateID() string { return e.ID.String() }
