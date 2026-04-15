package example

import "github.com/google/uuid"

// ExampleCreatedEvent dipublikasikan saat Example baru berhasil dibuat.
// TenantID dan CompanyID wajib ada agar event handler bisa memproses dengan isolasi yang benar.
type ExampleCreatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
}

func (e ExampleCreatedEvent) EventName() string   { return "example.created" }
func (e ExampleCreatedEvent) AggregateID() string { return e.ID.String() }

// ExampleUpdatedEvent dipublikasikan saat Example berhasil diupdate.
type ExampleUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
}

func (e ExampleUpdatedEvent) EventName() string   { return "example.updated" }
func (e ExampleUpdatedEvent) AggregateID() string { return e.ID.String() }
