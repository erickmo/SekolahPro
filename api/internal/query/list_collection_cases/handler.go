// Package list_collection_cases menangani query untuk mengambil daftar CollectionCase.
package list_collection_cases

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_collection_cases"

// Query berisi parameter untuk mengambil daftar CollectionCase.
type Query struct {
	Params pagination.ListParams
	Status *collection.CaseStatus
	AgingBucket *collection.AgingBucket
	CollectorID *uuid.UUID
	NasabahID   *uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu CollectionCase dalam hasil list.
type Item struct {
	ID                       string  `json:"id"`
	CaseNumber               string  `json:"case_number"`
	CurrentDPD               int     `json:"current_dpd"`
	AgingBucket              string  `json:"aging_bucket"`
	TotalOverdueAmount       int64   `json:"total_overdue_amount"`
	TotalOverdueInstallments int     `json:"total_overdue_installments"`
	PinjamanID               string  `json:"pinjaman_id"`
	NasabahID                string  `json:"nasabah_id"`
	AssignedCollectorID      *string `json:"assigned_collector_id"`
	EscalationLevel          int     `json:"escalation_level"`
	Status                   string  `json:"status"`
	OpenedAt                 string  `json:"opened_at"`
	CreatedAt                string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_collection_cases.
type Handler struct {
	repo collection.CaseReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo collection.CaseReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar CollectionCase dengan pagination, sorting, dan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	filter := collection.CaseFilter{
		Status:      qry.Status,
		AgingBucket: qry.AgingBucket,
		CollectorID: qry.CollectorID,
		NasabahID:   qry.NasabahID,
	}

	entities, total, err := h.repo.ListCases(ctx, s, filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list collection cases: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		item := &Item{
			ID:                       e.ID.String(),
			CaseNumber:               e.CaseNumber,
			CurrentDPD:               e.CurrentDPD,
			AgingBucket:              string(e.AgingBucket),
			TotalOverdueAmount:       e.TotalOverdueAmount,
			TotalOverdueInstallments: e.TotalOverdueInstallments,
			PinjamanID:               e.PinjamanID.String(),
			NasabahID:                e.NasabahID.String(),
			EscalationLevel:          e.EscalationLevel,
			Status:                   string(e.Status),
			OpenedAt:                 e.OpenedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:                e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.AssignedCollectorID != nil {
			s := e.AssignedCollectorID.String()
			item.AssignedCollectorID = &s
		}
		items = append(items, item)
	}

	return &Result{Data: items, Total: total}, nil
}
