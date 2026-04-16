package rat

import "github.com/google/uuid"

// MeetingCreatedEvent dipublikasikan saat RAT meeting baru berhasil dibuat.
type MeetingCreatedEvent struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	CompanyID    uuid.UUID `json:"company_id"`
	ID           uuid.UUID `json:"id"`
	MeetingType  string    `json:"meeting_type"`
	Title        string    `json:"title"`
	MeetingDate  string    `json:"meeting_date"`
}

func (e MeetingCreatedEvent) EventName() string   { return "rat_meeting.created" }
func (e MeetingCreatedEvent) AggregateID() string { return e.ID.String() }

// MeetingUpdatedEvent dipublikasikan saat RAT meeting berhasil diupdate.
type MeetingUpdatedEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	CompanyID uuid.UUID `json:"company_id"`
	ID        uuid.UUID `json:"id"`
}

func (e MeetingUpdatedEvent) EventName() string   { return "rat_meeting.updated" }
func (e MeetingUpdatedEvent) AggregateID() string { return e.ID.String() }

// MeetingCancelledEvent dipublikasikan saat RAT meeting dibatalkan.
type MeetingCancelledEvent struct {
	TenantID        uuid.UUID  `json:"tenant_id"`
	CompanyID       uuid.UUID  `json:"company_id"`
	ID              uuid.UUID  `json:"id"`
	CancelledReason *string    `json:"cancelled_reason,omitempty"`
}

func (e MeetingCancelledEvent) EventName() string   { return "rat_meeting.cancelled" }
func (e MeetingCancelledEvent) AggregateID() string { return e.ID.String() }
