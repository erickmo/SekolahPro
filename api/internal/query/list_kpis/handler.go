// Package list_kpis menangani query untuk mengambil daftar KPIDefinition.
package list_kpis

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_kpis"

// Query berisi parameter untuk mengambil daftar KPIDefinition.
type Query struct {
	Params pagination.ListParams
	Filter kpi.KPIDefinitionFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu KPIDefinition dalam hasil list.
type Item struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Code              string  `json:"code"`
	Category          string  `json:"category"`
	Frequency         string  `json:"frequency"`
	Unit              string  `json:"unit"`
	TargetValue       float64 `json:"target_value"`
	WarningThreshold  float64 `json:"warning_threshold"`
	CriticalThreshold float64 `json:"critical_threshold"`
	IsActive          bool    `json:"is_active"`
	CreatedAt         string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_kpis.
type Handler struct {
	repo kpi.KPIDefinitionReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIDefinitionReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar KPIDefinition dengan pagination, sorting, dan filter.
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
		return nil, fmt.Errorf("list kpis: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:                e.ID.String(),
			Name:              e.Name,
			Code:              e.Code,
			Category:          string(e.Category),
			Frequency:         string(e.Frequency),
			Unit:              e.Unit,
			TargetValue:       e.TargetValue,
			WarningThreshold:  e.WarningThreshold,
			CriticalThreshold: e.CriticalThreshold,
			IsActive:          e.IsActive,
			CreatedAt:         e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
