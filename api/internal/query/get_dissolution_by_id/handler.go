// Package get_dissolution_by_id menangani query untuk mengambil proses pembubaran berdasarkan ID.
package get_dissolution_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_dissolution_by_id"

// Query berisi parameter untuk mengambil proses pembubaran berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                    string   `json:"id"`
	DissolutionType       string   `json:"dissolution_type"`
	RatMeetingID          *string  `json:"rat_meeting_id,omitempty"`
	Reason                string   `json:"reason"`
	EffectiveDate         string   `json:"effective_date"`
	LiquidatorIDs         []string `json:"liquidator_ids"`
	SupervisorID          *string  `json:"supervisor_id,omitempty"`
	Stage                 string   `json:"stage"`
	ClaimDeadline         *string  `json:"claim_deadline,omitempty"`
	TotalAssets           *int64   `json:"total_assets,omitempty"`
	TotalLiabilities      *int64   `json:"total_liabilities,omitempty"`
	NetEquity             *int64   `json:"net_equity,omitempty"`
	DistributionPerMember *int64   `json:"distribution_per_member,omitempty"`
	Status                string   `json:"status"`
	FinalReportDocID      *string  `json:"final_report_doc_id,omitempty"`
	ClosedAt              *string  `json:"closed_at,omitempty"`
	InitiatedAt           string   `json:"initiated_at"`
	InitiatedBy           string   `json:"initiated_by"`
	CreatedAt             string   `json:"created_at"`
}

// Handler menangani Query get_dissolution_by_id.
type Handler struct {
	repo dissolution.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo dissolution.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil proses pembubaran berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get dissolution by id: %w", err)
	}

	return toResult(entity), nil
}

func toResult(e *dissolution.DissolutionProcess) *Result {
	r := &Result{
		ID:              e.ID.String(),
		DissolutionType: string(e.DissolutionType),
		Reason:          e.Reason,
		EffectiveDate:   e.EffectiveDate.Format("2006-01-02"),
		Stage:           string(e.Stage),
		Status:          string(e.Status),
		InitiatedAt:     e.InitiatedAt.Format("2006-01-02T15:04:05Z"),
		InitiatedBy:     e.InitiatedBy.String(),
		CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if e.RatMeetingID != nil {
		s := e.RatMeetingID.String()
		r.RatMeetingID = &s
	}
	if e.SupervisorID != nil {
		s := e.SupervisorID.String()
		r.SupervisorID = &s
	}
	if e.ClaimDeadline != nil {
		s := e.ClaimDeadline.Format("2006-01-02")
		r.ClaimDeadline = &s
	}
	if e.TotalAssets != nil {
		r.TotalAssets = e.TotalAssets
	}
	if e.TotalLiabilities != nil {
		r.TotalLiabilities = e.TotalLiabilities
	}
	if e.NetEquity != nil {
		r.NetEquity = e.NetEquity
	}
	if e.DistributionPerMember != nil {
		r.DistributionPerMember = e.DistributionPerMember
	}
	if e.FinalReportDocID != nil {
		s := e.FinalReportDocID.String()
		r.FinalReportDocID = &s
	}
	if e.ClosedAt != nil {
		s := e.ClosedAt.Format("2006-01-02T15:04:05Z")
		r.ClosedAt = &s
	}

	r.LiquidatorIDs = make([]string, len(e.LiquidatorIDs))
	for i, id := range e.LiquidatorIDs {
		r.LiquidatorIDs[i] = id.String()
	}

	return r
}
