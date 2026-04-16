// Package toggle_elimination_rule menangani command untuk mengaktifkan/menonaktifkan EliminationRule.
package toggle_elimination_rule

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "toggle_elimination_rule"

// Command berisi data untuk toggle active status EliminationRule.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command toggle_elimination_rule.
type Handler struct {
	repo branch_report.EliminationRuleWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo branch_report.EliminationRuleWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle melakukan toggle active status pada EliminationRule.
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

	_, err := h.repo.ToggleActive(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("toggle elimination rule: %w", err)
	}

	return nil
}
