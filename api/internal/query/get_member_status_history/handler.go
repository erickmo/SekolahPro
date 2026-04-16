// Package get_member_status_history menangani query untuk mengambil riwayat status anggota.
package get_member_status_history

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_member_status_history"

// Query berisi parameter untuk mengambil riwayat status anggota.
type Query struct {
	NasabahID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// StatusEntry adalah satu entri riwayat status.
type StatusEntry struct {
	ID               string `json:"id"`
	CurrentStatus    string `json:"current_status"`
	PreviousStatus   string `json:"previous_status,omitempty"`
	LastTransition   string `json:"last_transition"`
	TransitionReason string `json:"transition_reason"`
	AppliedAt        string `json:"applied_at"`
	UpdatedAt        string `json:"updated_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	NasabahID  string         `json:"nasabah_id"`
	History    []*StatusEntry `json:"history"`
	TotalCount int            `json:"total_count"`
}

// Handler menangani Query get_member_status_history.
type Handler struct {
	repo lifecycle.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil riwayat status anggota berdasarkan NasabahID.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	filter := lifecycle.LifecycleFilter{
		NasabahID: &qry.NasabahID,
	}

	entities, total, err := h.repo.List(ctx, s, filter, 100, 0, "created_at", "asc")
	if err != nil {
		return nil, fmt.Errorf("get member status history: %w", err)
	}

	history := make([]*StatusEntry, 0, len(entities))
	for _, e := range entities {
		entry := &StatusEntry{
			ID:               e.ID.String(),
			CurrentStatus:    string(e.CurrentStatus),
			LastTransition:   string(e.LastTransition),
			TransitionReason: e.TransitionReason,
			AppliedAt:        e.AppliedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        e.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.PreviousStatus != nil {
			entry.PreviousStatus = string(*e.PreviousStatus)
		}
		history = append(history, entry)
	}

	return &Result{
		NasabahID:  qry.NasabahID.String(),
		History:    history,
		TotalCount: total,
	}, nil
}
