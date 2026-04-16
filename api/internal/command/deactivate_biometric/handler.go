// Package deactivate_biometric menangani command untuk menonaktifkan biometric enrollment.
package deactivate_biometric

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "deactivate_biometric"

// Command berisi data untuk menonaktifkan biometric enrollment.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command deactivate_biometric.
type Handler struct {
	readRepo  biometric.ReadRepository
	writeRepo biometric.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr biometric.ReadRepository, wr biometric.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle menonaktifkan biometric enrollment (soft-deactivate).
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
		return fmt.Errorf("get biometric enrollment for deactivate: %w", err)
	}

	entity.IsActive = false

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("deactivate biometric enrollment: %w", err)
	}

	if err := h.eventBus.Publish(ctx, biometric.BiometricDeactivatedEvent{
		TenantID:      s.TenantID,
		CompanyID:     s.CompanyID,
		ID:            entity.ID,
		NasabahID:     entity.NasabahID,
		BiometricType: entity.BiometricType,
	}); err != nil {
		return fmt.Errorf("publish biometric.deactivated: %w", err)
	}

	return nil
}
