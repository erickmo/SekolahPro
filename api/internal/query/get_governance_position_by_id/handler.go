// Package get_governance_position_by_id menangani query untuk mengambil GovernancePosition berdasarkan ID.
package get_governance_position_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_governance_position_by_id"

// Query berisi parameter untuk mengambil GovernancePosition berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                string     `json:"id"`
	NasabahID         string     `json:"nasabah_id"`
	PositionType      string     `json:"position_type"`
	PositionLevel     string     `json:"position_level"`
	TermStart         string     `json:"term_start"`
	TermEnd           string     `json:"term_end"`
	TermNumber        int        `json:"term_number"`
	Status            string     `json:"status"`
	AppointedBy       string     `json:"appointed_by"`
	AppointmentDocID  *string    `json:"appointment_doc_id,omitempty"`
	MaxApprovalAmount int64      `json:"max_approval_amount"`
	CanDisburse       bool       `json:"can_disburse"`
	CanReverse        bool       `json:"can_reverse"`
	CanWaivePenalty   bool       `json:"can_waive_penalty"`
	CanWriteOff       bool       `json:"can_write_off"`
	CreatedAt         string     `json:"created_at"`
	UpdatedAt         string     `json:"updated_at"`
	CreatedBy         string     `json:"created_by"`
	UpdatedBy         string     `json:"updated_by"`
}

// Handler menangani Query get_governance_position_by_id.
type Handler struct {
	repo governance.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo governance.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil GovernancePosition berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get governance position by id: %w", err)
	}

	result := &Result{
		ID:                entity.ID.String(),
		NasabahID:         entity.NasabahID.String(),
		PositionType:      entity.PositionType,
		PositionLevel:     entity.PositionLevel,
		TermStart:         entity.TermStart.Format("2006-01-02T15:04:05Z"),
		TermEnd:           entity.TermEnd.Format("2006-01-02T15:04:05Z"),
		TermNumber:        entity.TermNumber,
		Status:            entity.Status,
		AppointedBy:       entity.AppointedBy,
		MaxApprovalAmount: entity.MaxApprovalAmount,
		CanDisburse:       entity.CanDisburse,
		CanReverse:        entity.CanReverse,
		CanWaivePenalty:   entity.CanWaivePenalty,
		CanWriteOff:       entity.CanWriteOff,
		CreatedAt:         entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		CreatedBy:         entity.CreatedBy.String(),
		UpdatedBy:         entity.UpdatedBy.String(),
	}
	if entity.AppointmentDocID != nil {
		aid := entity.AppointmentDocID.String()
		result.AppointmentDocID = &aid
	}

	return result, nil
}
