// Package delete_stress_test menangani command untuk menghapus StressTestScenario.
package delete_stress_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_stress_test"

// Command berisi data untuk menghapus StressTestScenario.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_stress_test.
type Handler struct {
	repo kpi.StressTestWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.StressTestWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle menghapus StressTestScenario berdasarkan ID.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if err := h.repo.DeleteScenario(ctx, s, cmd.ID); err != nil {
		return fmt.Errorf("delete stress test: %w", err)
	}
	return nil
}
