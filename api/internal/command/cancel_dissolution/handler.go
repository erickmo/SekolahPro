// Package cancel_dissolution menangani command untuk membatalkan proses pembubaran.
package cancel_dissolution

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "cancel_dissolution"

// Command berisi data untuk membatalkan proses pembubaran.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command cancel_dissolution.
type Handler struct {
	readRepo  dissolution.ReadRepository
	writeRepo dissolution.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr dissolution.ReadRepository, wr dissolution.WriteRepository) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr}
}

// Handle memvalidasi bahwa proses masih berstatus initiated, lalu membatalkan.
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

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get dissolution for cancel: %w", err)
	}

	if entity.Status == dissolution.StatusCompleted {
		return dissolution.ErrCannotCancelCompleted
	}
	if entity.Status == dissolution.StatusInProgress {
		return dissolution.ErrCannotCancelInProgress
	}
	if entity.Status == dissolution.StatusCancelled {
		return nil // idempotent
	}

	entity.Status = dissolution.StatusCancelled

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("cancel dissolution: %w", err)
	}

	return nil
}
