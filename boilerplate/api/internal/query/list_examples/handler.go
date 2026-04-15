// Package list_examples menangani query untuk mengambil daftar Example.
package list_examples

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_examples"

// Query berisi parameter untuk mengambil daftar Example.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu Example dalam hasil list.
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_examples.
type Handler struct {
	repo example.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo example.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar Example dengan pagination, sorting, dan filter scope organisasi.
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
		return nil, fmt.Errorf("list examples: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:          e.ID.String(),
			Name:        e.Name,
			Description: e.Description,
			IsActive:    e.IsActive,
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
