// Package delete_governance_position menangani command untuk soft-delete posisi governance.
package delete_governance_position

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_governance_position"

// Command berisi data untuk menghapus posisi governance.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_governance_position.
type Handler struct {
	repo governance.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo governance.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle melakukan soft delete dan memvalidasi scope.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}
	if cmd.ID == uuid.Nil {
		return fmt.Errorf("ID tidak boleh kosong")
	}
	return h.repo.Delete(ctx, s, cmd.ID)
}
