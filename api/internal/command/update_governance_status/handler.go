// Package update_governance_status menangani command untuk mengubah status posisi governance.
package update_governance_status

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_governance_status"

// Command berisi data yang dibutuhkan untuk mengubah status posisi governance.
type Command struct {
	ID        uuid.UUID
	Status    string
	UpdatedBy uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// validStatusTransitions mendefinisikan transisi status yang diizinkan.
var validStatusTransitions = map[string][]string{
	governance.StatusActive:   {governance.StatusResigned, governance.StatusRemoved, governance.StatusExpired},
	governance.StatusResigned: {},
	governance.StatusRemoved:  {},
	governance.StatusExpired:  {},
}

// Handler menangani Command update_governance_status.
type Handler struct {
	readRepo  governance.ReadRepository
	writeRepo governance.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr governance.ReadRepository, wr governance.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi transisi status dan mengupdate entity.
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
		return fmt.Errorf("get governance position for status update: %w", err)
	}

	if entity.Status == cmd.Status {
		return nil // idempotent: sudah di status yang diminta
	}

	// Validasi transisi status.
	allowed, exists := validStatusTransitions[entity.Status]
	if !exists {
		return fmt.Errorf("status saat ini tidak valid: %s", entity.Status)
	}

	validTransition := false
	for _, target := range allowed {
		if target == cmd.Status {
			validTransition = true
			break
		}
	}
	if !validTransition {
		return fmt.Errorf("transisi status dari %s ke %s tidak diizinkan", entity.Status, cmd.Status)
	}

	oldStatus := entity.Status
	entity.Status = cmd.Status
	entity.UpdatedAt = time.Now().UTC()
	if cmd.UpdatedBy != uuid.Nil {
		entity.UpdatedBy = cmd.UpdatedBy
	}

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update governance status: %w", err)
	}

	if err := h.eventBus.Publish(ctx, governance.GovernanceStatusChangedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		OldStatus: oldStatus,
		NewStatus: cmd.Status,
	}); err != nil {
		return fmt.Errorf("publish governance_position.status_changed: %w", err)
	}

	return nil
}
