package incident

import "github.com/google/uuid"

// IncidentCreatedEvent dipublikasikan saat IncidentRecord baru berhasil dibuat.
// TenantID dan CompanyID wajib ada agar event handler bisa memproses dengan isolasi yang benar.
type IncidentCreatedEvent struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	CompanyID    uuid.UUID `json:"company_id"`
	ID           uuid.UUID `json:"id"`
	Severity     string    `json:"severity"`
	Title        string    `json:"title"`
}

func (e IncidentCreatedEvent) EventName() string   { return "incident.created" }
func (e IncidentCreatedEvent) AggregateID() string { return e.ID.String() }

// IncidentResolvedEvent dipublikasikan saat IncidentRecord berhasil di-resolve.
type IncidentResolvedEvent struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	CompanyID      uuid.UUID `json:"company_id"`
	ID             uuid.UUID `json:"id"`
	Severity       string    `json:"severity"`
	ResolutionTime int       `json:"resolution_time"`
}

func (e IncidentResolvedEvent) EventName() string   { return "incident.resolved" }
func (e IncidentResolvedEvent) AggregateID() string { return e.ID.String() }
