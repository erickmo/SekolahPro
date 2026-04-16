// Package update_aml_rule menangani command untuk mengupdate monitoring rule.
package update_aml_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_aml_rule"

// Command berisi data yang dibutuhkan untuk mengupdate monitoring rule.
type Command struct {
	ID          uuid.UUID
	RuleName    string
	Description string
	Parameters  json.RawMessage
	AppliesTo   string
	AutoAlert   bool
	AutoBlock   bool
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_aml_rule.
type Handler struct {
	readRepo  aml_rule.ReadRepository
	writeRepo aml_rule.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr aml_rule.ReadRepository, wr aml_rule.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{readRepo: rr, writeRepo: wr, eventBus: eb}
}

// Handle memvalidasi command, mengambil entity yang ada, mengupdate, dan mempublish event.
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
		return fmt.Errorf("get aml rule for update: %w", err)
	}

	if cmd.RuleName != "" {
		entity.RuleName = cmd.RuleName
	}
	if cmd.Description != "" {
		entity.Description = cmd.Description
	}
	if cmd.Parameters != nil {
		entity.Parameters = cmd.Parameters
	}
	if cmd.AppliesTo != "" {
		entity.AppliesTo = cmd.AppliesTo
	}
	entity.AutoAlert = cmd.AutoAlert
	entity.AutoBlock = cmd.AutoBlock
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update aml rule: %w", err)
	}

	if err := h.eventBus.Publish(ctx, aml_rule.RuleUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
	}); err != nil {
		return fmt.Errorf("publish aml_rule.updated: %w", err)
	}

	return nil
}
