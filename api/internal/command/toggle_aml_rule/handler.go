// Package toggle_aml_rule menangani command untuk mengaktifkan/menonaktifkan monitoring rule.
package toggle_aml_rule

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "toggle_aml_rule"

// Command berisi data yang dibutuhkan untuk toggle monitoring rule.
type Command struct {
	ID       uuid.UUID
	IsActive bool
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command toggle_aml_rule.
type Handler struct {
	readRepo  aml_rule.ReadRepository
	writeRepo aml_rule.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr aml_rule.ReadRepository, wr aml_rule.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle mengaktifkan atau menonaktifkan monitoring rule.
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
		return fmt.Errorf("get aml rule for toggle: %w", err)
	}

	if entity.IsActive == cmd.IsActive {
		return nil // idempotent: sudah di status yang diminta
	}

	entity.IsActive = cmd.IsActive
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("toggle aml rule: %w", err)
	}

	if err := h.eventBus.Publish(ctx, aml_rule.RuleToggledEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		IsActive:  cmd.IsActive,
	}); err != nil {
		return fmt.Errorf("publish aml_rule.toggled: %w", err)
	}

	return nil
}
