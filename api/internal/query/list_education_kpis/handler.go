// Package list_education_kpis menangani query untuk mengambil daftar EducationKPI.
package list_education_kpis

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_education_kpis"

// Query berisi parameter untuk mengambil daftar EducationKPI.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu EducationKPI dalam hasil list.
type Item struct {
	ID          string  `json:"id"`
	MetricType  string  `json:"metric_type"`
	PeriodMonth string  `json:"period_month"`
	Value       float64 `json:"value"`
	CreatedAt   string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_education_kpis.
type Handler struct {
	repo enrollment.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo enrollment.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar EducationKPI dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.ListKPIs(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list education kpis: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:          e.ID.String(),
			MetricType:  string(e.MetricType),
			PeriodMonth: e.PeriodMonth,
			Value:       e.Value,
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
