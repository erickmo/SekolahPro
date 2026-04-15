// Package get_example_by_id menangani query untuk mengambil Example berdasarkan ID.
package get_example_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_example_by_id"

// Query berisi parameter untuk mengambil Example berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Handler menangani Query get_example_by_id.
type Handler struct {
	repo example.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo example.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil Example berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get example by id: %w", err)
	}

	return &Result{
		ID:          entity.ID.String(),
		Name:        entity.Name,
		Description: entity.Description,
		IsActive:    entity.IsActive,
		CreatedAt:   entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
