// Package list_incidents menangani query untuk mengambil daftar IncidentRecord.
package list_incidents

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/incident"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_incidents"

// Query berisi parameter untuk mengambil daftar IncidentRecord.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu IncidentRecord dalam hasil list.
type Item struct {
	ID              string   `json:"id"`
	Severity        string   `json:"severity"`
	Status          string   `json:"status"`
	IncidentType    string   `json:"incident_type"`
	Title           string   `json:"title"`
	AffectedSystems []string `json:"affected_systems"`
	CreatedAt       string   `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_incidents.
type Handler struct {
	repo incident.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo incident.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar IncidentRecord dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list incidents: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:              e.ID.String(),
			Severity:        string(e.Severity),
			Status:          string(e.Status),
			IncidentType:    e.IncidentType,
			Title:           e.Title,
			AffectedSystems: e.AffectedSystems,
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
