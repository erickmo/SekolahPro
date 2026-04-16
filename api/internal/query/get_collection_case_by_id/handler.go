// Package get_collection_case_by_id menangani query untuk mengambil CollectionCase berdasarkan ID.
package get_collection_case_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/collection"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_collection_case_by_id"

// Query berisi parameter untuk mengambil CollectionCase berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                       string  `json:"id"`
	CaseNumber               string  `json:"case_number"`
	CurrentDPD               int     `json:"current_dpd"`
	AgingBucket              string  `json:"aging_bucket"`
	TotalOverdueAmount       int64   `json:"total_overdue_amount"`
	TotalOverdueInstallments int     `json:"total_overdue_installments"`
	PinjamanID               string  `json:"pinjaman_id"`
	NasabahID                string  `json:"nasabah_id"`
	AssignedCollectorID      *string `json:"assigned_collector_id"`
	AssignedAt               *string `json:"assigned_at"`
	EscalationLevel          int     `json:"escalation_level"`
	Status                   string  `json:"status"`
	ResolutionType           *string `json:"resolution_type"`
	OpenedAt                 string  `json:"opened_at"`
	ClosedAt                 *string `json:"closed_at"`
	CreatedAt                string  `json:"created_at"`
	UpdatedAt                string  `json:"updated_at"`
}

// Handler menangani Query get_collection_case_by_id.
type Handler struct {
	repo collection.CaseReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo collection.CaseReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil CollectionCase berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetCaseByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get collection case by id: %w", err)
	}

	result := &Result{
		ID:                       entity.ID.String(),
		CaseNumber:               entity.CaseNumber,
		CurrentDPD:               entity.CurrentDPD,
		AgingBucket:              string(entity.AgingBucket),
		TotalOverdueAmount:       entity.TotalOverdueAmount,
		TotalOverdueInstallments: entity.TotalOverdueInstallments,
		PinjamanID:               entity.PinjamanID.String(),
		NasabahID:                entity.NasabahID.String(),
		EscalationLevel:          entity.EscalationLevel,
		Status:                   string(entity.Status),
		OpenedAt:                 entity.OpenedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:                entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:                entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.AssignedCollectorID != nil {
		s := entity.AssignedCollectorID.String()
		result.AssignedCollectorID = &s
	}
	if entity.AssignedAt != nil {
		s := entity.AssignedAt.Format("2006-01-02T15:04:05Z")
		result.AssignedAt = &s
	}
	if entity.ResolutionType != nil {
		s := string(*entity.ResolutionType)
		result.ResolutionType = &s
	}
	if entity.ClosedAt != nil {
		s := entity.ClosedAt.Format("2006-01-02T15:04:05Z")
		result.ClosedAt = &s
	}

	return result, nil
}
