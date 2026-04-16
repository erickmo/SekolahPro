// Package list_kpi_measurements menangani query untuk mengambil daftar KPIMeasurement.
package list_kpi_measurements

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_kpi_measurements"

// Query berisi parameter untuk mengambil daftar KPIMeasurement.
type Query struct {
	Params pagination.ListParams
	Filter kpi.KPIMeasurementFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu KPIMeasurement dalam hasil list.
type Item struct {
	ID            string  `json:"id"`
	DefinitionID  string  `json:"definition_id"`
	MeasuredValue float64 `json:"measured_value"`
	Status        string  `json:"status"`
	MeasuredAt    string  `json:"measured_at"`
	MeasuredBy    string  `json:"measured_by"`
	Notes         string  `json:"notes"`
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
	CreatedAt     string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_kpi_measurements.
type Handler struct {
	repo kpi.KPIMeasurementReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIMeasurementReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar KPIMeasurement dengan pagination dan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.ListMeasurements(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list kpi measurements: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:            e.ID.String(),
			DefinitionID:  e.DefinitionID.String(),
			MeasuredValue: e.MeasuredValue,
			Status:        string(e.Status),
			MeasuredAt:    e.MeasuredAt.Format("2006-01-02T15:04:05Z"),
			MeasuredBy:    e.MeasuredBy.String(),
			Notes:         e.Notes,
			PeriodStart:   e.PeriodStart.Format("2006-01-02T15:04:05Z"),
			PeriodEnd:     e.PeriodEnd.Format("2006-01-02T15:04:05Z"),
			CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
