// Package get_kpi_by_id menangani query untuk mengambil KPIDefinition berdasarkan ID.
package get_kpi_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_kpi_by_id"

// Query berisi parameter untuk mengambil KPIDefinition berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
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
	Description       string  `json:"description"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// Handler menangani Query get_kpi_by_id.
type Handler struct {
	repo kpi.KPIDefinitionReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIDefinitionReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil KPIDefinition berdasarkan ID.
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
		return nil, fmt.Errorf("get kpi by id: %w", err)
	}

	return &Result{
		ID:                entity.ID.String(),
		Name:              entity.Name,
		Code:              entity.Code,
		Category:          string(entity.Category),
		Frequency:         string(entity.Frequency),
		Unit:              entity.Unit,
		TargetValue:       entity.TargetValue,
		WarningThreshold:  entity.WarningThreshold,
		CriticalThreshold: entity.CriticalThreshold,
		IsActive:          entity.IsActive,
		Description:       entity.Description,
		CreatedAt:         entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
