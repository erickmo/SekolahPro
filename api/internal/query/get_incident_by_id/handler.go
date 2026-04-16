// Package get_incident_by_id menangani query untuk mengambil IncidentRecord berdasarkan ID.
package get_incident_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/incident"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_incident_by_id"

// Query berisi parameter untuk mengambil IncidentRecord berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID              string   `json:"id"`
	Severity        string   `json:"severity"`
	Status          string   `json:"status"`
	IncidentType    string   `json:"incident_type"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	AffectedSystems []string `json:"affected_systems"`
	RootCause       *string  `json:"root_cause"`
	Resolution      *string  `json:"resolution"`
	ResolutionTime  *int     `json:"resolution_time"`
	ResolvedBy      *string  `json:"resolved_by"`
	ResolvedAt      *string  `json:"resolved_at"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// Handler menangani Query get_incident_by_id.
type Handler struct {
	repo incident.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo incident.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil IncidentRecord berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get incident by id: %w", err)
	}

	result := &Result{
		ID:              entity.ID.String(),
		Severity:        string(entity.Severity),
		Status:          string(entity.Status),
		IncidentType:    entity.IncidentType,
		Title:           entity.Title,
		Description:     entity.Description,
		AffectedSystems: entity.AffectedSystems,
		RootCause:       entity.RootCause,
		Resolution:      entity.Resolution,
		ResolutionTime:  entity.ResolutionTime,
		CreatedAt:       entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.ResolvedBy != nil {
		v := entity.ResolvedBy.String()
		result.ResolvedBy = &v
	}
	if entity.ResolvedAt != nil {
		v := entity.ResolvedAt.Format("2006-01-02T15:04:05Z")
		result.ResolvedAt = &v
	}

	return result, nil
}
