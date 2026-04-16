// Package list_governance_positions menangani query untuk mengambil daftar GovernancePosition.
package list_governance_positions

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_governance_positions"

// Query berisi parameter untuk mengambil daftar GovernancePosition.
type Query struct {
	Params pagination.ListParams
	Filter governance.ListFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu GovernancePosition dalam hasil list.
type Item struct {
	ID                string `json:"id"`
	NasabahID         string `json:"nasabah_id"`
	PositionType      string `json:"position_type"`
	PositionLevel     string `json:"position_level"`
	TermStart         string `json:"term_start"`
	TermEnd           string `json:"term_end"`
	TermNumber        int    `json:"term_number"`
	Status            string `json:"status"`
	AppointedBy       string `json:"appointed_by"`
	MaxApprovalAmount int64  `json:"max_approval_amount"`
	CanDisburse       bool   `json:"can_disburse"`
	CanReverse        bool   `json:"can_reverse"`
	CanWaivePenalty   bool   `json:"can_waive_penalty"`
	CanWriteOff       bool   `json:"can_write_off"`
	CreatedAt         string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_governance_positions.
type Handler struct {
	repo governance.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo governance.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar GovernancePosition dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list governance positions: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:                e.ID.String(),
			NasabahID:         e.NasabahID.String(),
			PositionType:      e.PositionType,
			PositionLevel:     e.PositionLevel,
			TermStart:         e.TermStart.Format("2006-01-02T15:04:05Z"),
			TermEnd:           e.TermEnd.Format("2006-01-02T15:04:05Z"),
			TermNumber:        e.TermNumber,
			Status:            e.Status,
			AppointedBy:       e.AppointedBy,
			MaxApprovalAmount: e.MaxApprovalAmount,
			CanDisburse:       e.CanDisburse,
			CanReverse:        e.CanReverse,
			CanWaivePenalty:   e.CanWaivePenalty,
			CanWriteOff:       e.CanWriteOff,
			CreatedAt:         e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
