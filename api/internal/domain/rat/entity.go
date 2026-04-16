// Package rat adalah domain CQRS untuk Rapat Anggota Tahunan management (ADR-K026).
//
// Mengelola RAT meetings, agenda items, attendance, voting, dan elections.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package rat

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// Meeting adalah entity utama untuk RAT meeting.
type Meeting struct {
	ID                  uuid.UUID  `db:"id"                    json:"id"`
	MeetingType         string     `db:"meeting_type"          json:"meeting_type"`
	MeetingNumber       string     `db:"meeting_number"        json:"meeting_number"`
	Title               string     `db:"title"                 json:"title"`
	Description         string     `db:"description"           json:"description,omitempty"`
	MeetingDate         time.Time  `db:"meeting_date"          json:"meeting_date"`
	MeetingTime         string     `db:"meeting_time"          json:"meeting_time"`
	MeetingLocation     string     `db:"meeting_location"      json:"meeting_location"`
	MeetingMode         string     `db:"meeting_mode"          json:"meeting_mode"`
	OnlineLink          *string    `db:"online_link"           json:"online_link,omitempty"`
	TotalEligibleMembers int       `db:"total_eligible_members" json:"total_eligible_members"`
	QuorumRequired      int        `db:"quorum_required"       json:"quorum_required"`
	ActualAttendees     int        `db:"actual_attendees"      json:"actual_attendees"`
	QuorumMet           bool       `db:"quorum_met"            json:"quorum_met"`
	Status              string     `db:"status"                json:"status"`
	CancelledReason     *string    `db:"cancelled_reason"      json:"cancelled_reason,omitempty"`
	RescheduledToID     *uuid.UUID `db:"rescheduled_to_id"     json:"rescheduled_to_id,omitempty"`
	AgendaDocID         *uuid.UUID `db:"agenda_doc_id"         json:"agenda_doc_id,omitempty"`
	MinutesDocID        *uuid.UUID `db:"minutes_doc_id"        json:"minutes_doc_id,omitempty"`
	FinancialReportID   *uuid.UUID `db:"financial_report_id"   json:"financial_report_id,omitempty"`
	ShuProposalID       *uuid.UUID `db:"shu_proposal_id"       json:"shu_proposal_id,omitempty"`
	ConvenedBy          uuid.UUID  `db:"convened_by"           json:"convened_by"`
	SecretaryID         uuid.UUID  `db:"secretary_id"          json:"secretary_id"`
	CreatedAt           time.Time  `db:"created_at"            json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"            json:"updated_at"`
	DeletedAt           *time.Time `db:"deleted_at"            json:"-"`
}

// Meeting type constants.
const (
	MeetingTypeAnnual     = "annual"
	MeetingTypeSpecial    = "special"
	MeetingTypeJoint      = "joint"
	MeetingTypeDissolution = "dissolution"
)

// Meeting status constants.
const (
	MeetingStatusDraft          = "draft"
	MeetingStatusInvitationSent = "invitation_sent"
	MeetingStatusInProgress     = "in_progress"
	MeetingStatusCompleted      = "completed"
	MeetingStatusCancelled      = "cancelled"
	MeetingStatusRescheduled    = "rescheduled"
)

// Meeting mode constants.
const (
	MeetingModeOffline = "offline"
	MeetingModeOnline  = "online"
	MeetingModeHybrid  = "hybrid"
)

// ListFilter menyimpan filter opsional untuk List query.
type ListFilter struct {
	MeetingType *string
	Status      *string
	MeetingDate *time.Time
}

// WriteRepository mendefinisikan operasi write untuk RAT meetings.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *Meeting) error
	Update(ctx context.Context, s scope.Scope, e *Meeting) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk RAT meetings.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*Meeting, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*Meeting, int, error)
}
