// Package create_elimination_rule menangani command untuk membuat EliminationRule baru.
package create_elimination_rule

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_elimination_rule"

// Command berisi data yang dibutuhkan untuk membuat EliminationRule baru.
type Command struct {
	RuleName     string
	RuleType     string
	FromBranchID uuid.UUID
	ToBranchID   uuid.UUID
	Description  string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_elimination_rule.
type Handler struct {
	repo branch_report.EliminationRuleWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo branch_report.EliminationRuleWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle memvalidasi command, membuat entity, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.RuleName == "" {
		return branch_report.ErrRuleNameEmpty
	}

	rt, ok := branch_report.ValidRuleTypes[cmd.RuleType]
	if !ok {
		return branch_report.ErrInvalidRuleType
	}

	if cmd.FromBranchID == cmd.ToBranchID {
		return branch_report.ErrSameBranchTransfer
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &branch_report.EliminationRule{
		ID:           id,
		RuleName:     cmd.RuleName,
		RuleType:     rt,
		FromBranchID: cmd.FromBranchID,
		ToBranchID:   cmd.ToBranchID,
		IsActive:     true,
		Description:  cmd.Description,
		CreatedAt:    now,
	}

	if err := h.repo.SaveRule(ctx, s, entity); err != nil {
		return fmt.Errorf("save elimination rule: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
