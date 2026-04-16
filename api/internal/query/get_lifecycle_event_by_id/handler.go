// Package get_lifecycle_event_by_id menangani query untuk mengambil MembershipLifecycle berdasarkan ID.
package get_lifecycle_event_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_lifecycle_event_by_id"

// Query berisi parameter untuk mengambil MembershipLifecycle berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                string `json:"id"`
	NasabahID         string `json:"nasabah_id"`
	CurrentStatus     string `json:"current_status"`
	PreviousStatus    string `json:"previous_status,omitempty"`
	LastTransition    string `json:"last_transition"`
	TransitionReason  string `json:"transition_reason"`
	AppliedAt         string `json:"applied_at"`
	VerifiedAt        string `json:"verified_at,omitempty"`
	VerifiedBy        string `json:"verified_by,omitempty"`
	ActivatedAt       string `json:"activated_at,omitempty"`
	SuspendedAt       string `json:"suspended_at,omitempty"`
	SuspendedBy       string `json:"suspended_by,omitempty"`
	ResignedAt        string `json:"resigned_at,omitempty"`
	RevokedAt         string `json:"revoked_at,omitempty"`
	RevokedBy         string `json:"revoked_by,omitempty"`
	MembershipNo      string `json:"membership_no"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// Handler menangani Query get_lifecycle_event_by_id.
type Handler struct {
	repo lifecycle.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil MembershipLifecycle berdasarkan ID.
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
		return nil, fmt.Errorf("get lifecycle by id: %w", err)
	}

	result := &Result{
		ID:               entity.ID.String(),
		NasabahID:        entity.NasabahID.String(),
		CurrentStatus:    string(entity.CurrentStatus),
		LastTransition:   string(entity.LastTransition),
		TransitionReason: entity.TransitionReason,
		AppliedAt:        entity.AppliedAt.Format("2006-01-02T15:04:05Z"),
		MembershipNo:     entity.MembershipNo,
		CreatedAt:        entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.PreviousStatus != nil {
		result.PreviousStatus = string(*entity.PreviousStatus)
	}
	if entity.VerifiedAt != nil {
		result.VerifiedAt = entity.VerifiedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.VerifiedBy != nil {
		result.VerifiedBy = entity.VerifiedBy.String()
	}
	if entity.ActivatedAt != nil {
		result.ActivatedAt = entity.ActivatedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.SuspendedAt != nil {
		result.SuspendedAt = entity.SuspendedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.SuspendedBy != nil {
		result.SuspendedBy = entity.SuspendedBy.String()
	}
	if entity.ResignedAt != nil {
		result.ResignedAt = entity.ResignedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.RevokedAt != nil {
		result.RevokedAt = entity.RevokedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.RevokedBy != nil {
		result.RevokedBy = entity.RevokedBy.String()
	}

	return result, nil
}
