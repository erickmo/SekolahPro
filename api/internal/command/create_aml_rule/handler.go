// Package create_aml_rule menangani command untuk membuat monitoring rule baru.
package create_aml_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_aml_rule"

// Command berisi data yang dibutuhkan untuk membuat monitoring rule baru.
type Command struct {
	RuleCode    string
	RuleName    string
	RuleType    string
	Description string
	Parameters  json.RawMessage
	IsActive    bool
	AppliesTo   string
	AutoAlert   bool
	AutoBlock   bool
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_aml_rule.
type Handler struct {
	repo     aml_rule.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo aml_rule.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{repo: repo, eventBus: eb}
}

// Handle memvalidasi command, membuat entity, menyimpan ke repository,
// dan mempublikasikan domain event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.RuleCode == "" {
		return fmt.Errorf("rule_code tidak boleh kosong")
	}
	if cmd.RuleName == "" {
		return fmt.Errorf("rule_name tidak boleh kosong")
	}
	if cmd.RuleType == "" {
		return fmt.Errorf("rule_type tidak boleh kosong")
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &aml_rule.MonitoringRule{
		ID:          id,
		RuleCode:    cmd.RuleCode,
		RuleName:    cmd.RuleName,
		RuleType:    cmd.RuleType,
		Description: cmd.Description,
		Parameters:  cmd.Parameters,
		IsActive:    cmd.IsActive,
		AppliesTo:   cmd.AppliesTo,
		AutoAlert:   cmd.AutoAlert,
		AutoBlock:   cmd.AutoBlock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if !entity.IsActive {
		entity.IsActive = true // default aktif
	}
	if entity.AppliesTo == "" {
		entity.AppliesTo = aml_rule.AppliesToAll
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save aml rule: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, aml_rule.RuleCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		RuleCode:  cmd.RuleCode,
		RuleType:  cmd.RuleType,
	}); err != nil {
		return fmt.Errorf("publish aml_rule.created: %w", err)
	}

	return nil
}
