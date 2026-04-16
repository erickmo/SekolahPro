// Package update_dissolution_stage menangani command untuk mengubah tahapan proses pembubaran.
package update_dissolution_stage

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_dissolution_stage"

// Command berisi data yang dibutuhkan untuk mengubah tahapan pembubaran.
type Command struct {
	ID       uuid.UUID
	NewStage dissolution.Stage
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_dissolution_stage.
type Handler struct {
	readRepo dissolution.ReadRepository
	writeRepo dissolution.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr dissolution.ReadRepository, wr dissolution.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi transisi tahapan, mengupdate entity, dan mempublish event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get dissolution for stage update: %w", err)
	}

	if entity.Status == dissolution.StatusCompleted || entity.Status == dissolution.StatusCancelled {
		return dissolution.ErrAlreadyClosed
	}

	if entity.Stage == cmd.NewStage {
		return nil // idempotent: sudah di stage yang diminta
	}

	// Validasi transisi: hanya boleh maju satu tahap kecuali ke closed.
	next, canAdvance := dissolution.NextStage(entity.Stage)
	if !canAdvance || cmd.NewStage != next {
		return dissolution.ErrInvalidStageTransition
	}

	oldStage := entity.Stage
	entity.Stage = cmd.NewStage

	// Jika stage bukan lagi announced, status otomatis in_progress.
	if entity.Status == dissolution.StatusInitiated {
		entity.Status = dissolution.StatusInProgress
	}

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update dissolution stage: %w", err)
	}

	if err := h.eventBus.Publish(ctx, dissolution.DissolutionStageChangedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		OldStage:  oldStage,
		NewStage:  cmd.NewStage,
	}); err != nil {
		return fmt.Errorf("publish dissolution.stage_changed: %w", err)
	}

	return nil
}
