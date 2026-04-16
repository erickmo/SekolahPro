// Package delete_dissolution menangani command untuk soft-delete proses pembubaran.
package delete_dissolution

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_dissolution"

// Command berisi data untuk menghapus proses pembubaran.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_dissolution.
type Handler struct {
	repo dissolution.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo dissolution.WriteRepository) *Handler {
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
