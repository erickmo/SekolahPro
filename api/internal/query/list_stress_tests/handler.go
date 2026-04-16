// Package list_stress_tests menangani query untuk mengambil daftar StressTestScenario.
package list_stress_tests

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_stress_tests"

// Query berisi parameter untuk mengambil daftar StressTestScenario.
type Query struct {
	Params pagination.ListParams
	Filter kpi.StressTestFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu StressTestScenario dalam hasil list.
type Item struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ScenarioStatus string `json:"scenario_status"`
	Parameters     string `json:"parameters"`
	SimulatedAt    string `json:"simulated_at,omitempty"`
	SimulatedBy    string `json:"simulated_by,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_stress_tests.
type Handler struct {
	repo kpi.StressTestReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.StressTestReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar StressTestScenario dengan pagination dan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.ListScenarios(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list stress tests: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		item := &Item{
			ID:             e.ID.String(),
			Name:           e.Name,
			Description:    e.Description,
			ScenarioStatus: string(e.ScenarioStatus),
			Parameters:     e.Parameters,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.SimulatedAt != nil {
			item.SimulatedAt = e.SimulatedAt.Format("2006-01-02T15:04:05Z")
		}
		if e.SimulatedBy != nil {
			item.SimulatedBy = e.SimulatedBy.String()
		}
		if e.CompletedAt != nil {
			item.CompletedAt = e.CompletedAt.Format("2006-01-02T15:04:05Z")
		}
		items = append(items, item)
	}

	return &Result{Data: items, Total: total}, nil
}
