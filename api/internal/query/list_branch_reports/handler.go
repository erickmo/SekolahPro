// Package list_branch_reports menangani query untuk mengambil daftar BranchFinancialSummary.
package list_branch_reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_branch_reports"

// Query berisi parameter untuk mengambil daftar BranchFinancialSummary.
type Query struct {
	Params     pagination.ListParams
	BranchID   *uuid.UUID
	PeriodType *branch_report.PeriodType
	StartDate  *time.Time
	EndDate    *time.Time
	ApprovedBy *uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu BranchFinancialSummary dalam hasil list.
type Item struct {
	ID                string   `json:"id"`
	BranchID          *string  `json:"branch_id"`
	PeriodType        string   `json:"period_type"`
	PeriodStart       string   `json:"period_start"`
	PeriodEnd         string   `json:"period_end"`
	TotalPendapatan   int64    `json:"total_pendapatan"`
	TotalBiaya        int64    `json:"total_biaya"`
	LabaRugiBersih    int64    `json:"laba_rugi_bersih"`
	TotalAset         int64    `json:"total_aset"`
	TotalKewajiban    int64    `json:"total_kewajiban"`
	ApprovedBy        *string  `json:"approved_by"`
	CreatedAt         string   `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_branch_reports.
type Handler struct {
	repo branch_report.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo branch_report.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar BranchFinancialSummary dengan pagination, sorting, dan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	filter := branch_report.ListFilter{
		BranchID:    qry.BranchID,
		PeriodType:  qry.PeriodType,
		PeriodStart: qry.StartDate,
		PeriodEnd:   qry.EndDate,
		ApprovedBy:  qry.ApprovedBy,
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list branch reports: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		item := &Item{
			ID:              e.ID.String(),
			PeriodType:      string(e.PeriodType),
			PeriodStart:     e.PeriodStart.Format("2006-01-02"),
			PeriodEnd:       e.PeriodEnd.Format("2006-01-02"),
			TotalPendapatan: e.TotalPendapatan,
			TotalBiaya:      e.TotalBiaya,
			LabaRugiBersih:  e.LabaRugiBersih,
			TotalAset:       e.TotalAset,
			TotalKewajiban:  e.TotalKewajiban,
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.BranchID != nil {
			bid := e.BranchID.String()
			item.BranchID = &bid
		}
		if e.ApprovedBy != nil {
			aid := e.ApprovedBy.String()
			item.ApprovedBy = &aid
		}
		items = append(items, item)
	}

	return &Result{Data: items, Total: total}, nil
}
