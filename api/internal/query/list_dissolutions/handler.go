// Package list_dissolutions menangani query untuk mengambil daftar proses pembubaran.
package list_dissolutions

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_dissolutions"

// Query berisi parameter untuk mengambil daftar proses pembubaran.
type Query struct {
	Params pagination.ListParams
	Filter dissolution.ListFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu proses pembubaran dalam hasil list.
type Item struct {
	ID              string `json:"id"`
	DissolutionType string `json:"dissolution_type"`
	Reason          string `json:"reason"`
	EffectiveDate   string `json:"effective_date"`
	Stage           string `json:"stage"`
	Status          string `json:"status"`
	InitiatedAt     string `json:"initiated_at"`
	InitiatedBy     string `json:"initiated_by"`
	CreatedAt       string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_dissolutions.
type Handler struct {
	repo dissolution.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo dissolution.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar proses pembubaran dengan pagination, sorting, filter, dan scope.
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
		return nil, fmt.Errorf("list dissolutions: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:              e.ID.String(),
			DissolutionType: string(e.DissolutionType),
			Reason:          e.Reason,
			EffectiveDate:   e.EffectiveDate.Format("2006-01-02"),
			Stage:           string(e.Stage),
			Status:          string(e.Status),
			InitiatedAt:     e.InitiatedAt.Format("2006-01-02T15:04:05Z"),
			InitiatedBy:     e.InitiatedBy.String(),
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
