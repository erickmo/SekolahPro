// Package list_lifecycle_events menangani query untuk mengambil daftar MembershipLifecycle.
package list_lifecycle_events

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_lifecycle_events"

// Query berisi parameter untuk mengambil daftar MembershipLifecycle.
type Query struct {
	Params pagination.ListParams
	Filter lifecycle.LifecycleFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu MembershipLifecycle dalam hasil list.
type Item struct {
	ID               string `json:"id"`
	NasabahID        string `json:"nasabah_id"`
	CurrentStatus    string `json:"current_status"`
	LastTransition   string `json:"last_transition"`
	TransitionReason string `json:"transition_reason"`
	MembershipNo     string `json:"membership_no"`
	AppliedAt        string `json:"applied_at"`
	CreatedAt        string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_lifecycle_events.
type Handler struct {
	repo lifecycle.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar MembershipLifecycle dengan pagination dan filter.
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
		return nil, fmt.Errorf("list lifecycle events: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:               e.ID.String(),
			NasabahID:        e.NasabahID.String(),
			CurrentStatus:    string(e.CurrentStatus),
			LastTransition:   string(e.LastTransition),
			TransitionReason: e.TransitionReason,
			MembershipNo:     e.MembershipNo,
			AppliedAt:        e.AppliedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:        e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
