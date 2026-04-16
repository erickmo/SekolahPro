// Package get_rat_meeting_by_id menangani query untuk mengambil RAT Meeting berdasarkan ID.
package get_rat_meeting_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/rat"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_rat_meeting_by_id"

// Query berisi parameter untuk mengambil RAT Meeting berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                  string     `json:"id"`
	MeetingType         string     `json:"meeting_type"`
	MeetingNumber       string     `json:"meeting_number"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	MeetingDate         string     `json:"meeting_date"`
	MeetingTime         string     `json:"meeting_time"`
	MeetingLocation     string     `json:"meeting_location"`
	MeetingMode         string     `json:"meeting_mode"`
	OnlineLink          *string    `json:"online_link,omitempty"`
	TotalEligibleMembers int       `json:"total_eligible_members"`
	QuorumRequired      int        `json:"quorum_required"`
	ActualAttendees     int        `json:"actual_attendees"`
	QuorumMet           bool       `json:"quorum_met"`
	Status              string     `json:"status"`
	CancelledReason     *string    `json:"cancelled_reason,omitempty"`
	RescheduledToID     *string    `json:"rescheduled_to_id,omitempty"`
	AgendaDocID         *string    `json:"agenda_doc_id,omitempty"`
	MinutesDocID        *string    `json:"minutes_doc_id,omitempty"`
	FinancialReportID   *string    `json:"financial_report_id,omitempty"`
	ShuProposalID       *string    `json:"shu_proposal_id,omitempty"`
	ConvenedBy          string     `json:"convened_by"`
	SecretaryID         string     `json:"secretary_id"`
	CreatedAt           string     `json:"created_at"`
	UpdatedAt           string     `json:"updated_at"`
}

// Handler menangani Query get_rat_meeting_by_id.
type Handler struct {
	repo rat.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo rat.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil RAT Meeting berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get rat meeting by id: %w", err)
	}

	result := &Result{
		ID:                  entity.ID.String(),
		MeetingType:         entity.MeetingType,
		MeetingNumber:       entity.MeetingNumber,
		Title:               entity.Title,
		Description:         entity.Description,
		MeetingDate:         entity.MeetingDate.Format("2006-01-02T15:04:05Z"),
		MeetingTime:         entity.MeetingTime,
		MeetingLocation:     entity.MeetingLocation,
		MeetingMode:         entity.MeetingMode,
		OnlineLink:          entity.OnlineLink,
		TotalEligibleMembers: entity.TotalEligibleMembers,
		QuorumRequired:      entity.QuorumRequired,
		ActualAttendees:     entity.ActualAttendees,
		QuorumMet:           entity.QuorumMet,
		Status:              entity.Status,
		CancelledReason:     entity.CancelledReason,
		ConvenedBy:          entity.ConvenedBy.String(),
		SecretaryID:         entity.SecretaryID.String(),
		CreatedAt:           entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if entity.RescheduledToID != nil {
		rid := entity.RescheduledToID.String()
		result.RescheduledToID = &rid
	}
	if entity.AgendaDocID != nil {
		aid := entity.AgendaDocID.String()
		result.AgendaDocID = &aid
	}
	if entity.MinutesDocID != nil {
		mid := entity.MinutesDocID.String()
		result.MinutesDocID = &mid
	}
	if entity.FinancialReportID != nil {
		fid := entity.FinancialReportID.String()
		result.FinancialReportID = &fid
	}
	if entity.ShuProposalID != nil {
		sid := entity.ShuProposalID.String()
		result.ShuProposalID = &sid
	}

	return result, nil
}
