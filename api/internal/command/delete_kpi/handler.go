// Package delete_kpi menangani command untuk menghapus KPIDefinition.
package delete_kpi

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_kpi"

// Command berisi data untuk menghapus KPIDefinition.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_kpi.
type Handler struct {
	repo kpi.KPIDefinitionWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIDefinitionWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle menghapus KPIDefinition berdasarkan ID.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if err := h.repo.Delete(ctx, s, cmd.ID); err != nil {
		return fmt.Errorf("delete kpi: %w", err)
	}
	return nil
}
