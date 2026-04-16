package consent

import "github.com/google/uuid"

// ConsentCreatedEvent dipublikasikan saat consent record baru berhasil dibuat.
type ConsentCreatedEvent struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	CompanyID    uuid.UUID `json:"company_id"`
	ID           uuid.UUID `json:"id"`
	NasabahID    uuid.UUID `json:"nasabah_id"`
	ConsentType  string    `json:"consent_type"`
	ConsentGiven bool      `json:"consent_given"`
}

func (e ConsentCreatedEvent) EventName() string   { return "consent_record.created" }
func (e ConsentCreatedEvent) AggregateID() string { return e.ID.String() }

// ConsentWithdrawnEvent dipublikasikan saat consent dicabut.
type ConsentWithdrawnEvent struct {
	TenantID uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID       uuid.UUID `json:"id"`
}

func (e ConsentWithdrawnEvent) EventName() string   { return "consent_record.withdrawn" }
func (e ConsentWithdrawnEvent) AggregateID() string { return e.ID.String() }
