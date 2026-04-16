package collection

import "github.com/google/uuid"

// CollectionCaseCreatedEvent dipublikasikan saat CollectionCase baru berhasil dibuat.
type CollectionCaseCreatedEvent struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	CompanyID   uuid.UUID `json:"company_id"`
	ID          uuid.UUID `json:"id"`
	CaseNumber  string    `json:"case_number"`
	PinjamanID  uuid.UUID `json:"pinjaman_id"`
	NasabahID   uuid.UUID `json:"nasabah_id"`
	AgingBucket string    `json:"aging_bucket"`
}

func (e CollectionCaseCreatedEvent) EventName() string   { return "collection_case.created" }
func (e CollectionCaseCreatedEvent) AggregateID() string { return e.ID.String() }

// CollectionCaseResolvedEvent dipublikasikan saat CollectionCase berhasil diresolusi.
type CollectionCaseResolvedEvent struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	CompanyID      uuid.UUID `json:"company_id"`
	ID             uuid.UUID `json:"id"`
	ResolutionType string    `json:"resolution_type"`
}

func (e CollectionCaseResolvedEvent) EventName() string   { return "collection_case.resolved" }
func (e CollectionCaseResolvedEvent) AggregateID() string { return e.ID.String() }

// AgingBucketChangedEvent dipublikasikan saat aging bucket berubah pada CollectionCase.
type AgingBucketChangedEvent struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	CompanyID   uuid.UUID `json:"company_id"`
	ID          uuid.UUID `json:"id"`
	OldBucket   string    `json:"old_bucket"`
	NewBucket   string    `json:"new_bucket"`
	CurrentDPD  int       `json:"current_dpd"`
}

func (e AgingBucketChangedEvent) EventName() string   { return "collection_case.aging_bucket_changed" }
func (e AgingBucketChangedEvent) AggregateID() string { return e.ID.String() }
