// Package create_kpi menangani command untuk membuat KPIDefinition baru.
package create_kpi

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_kpi"

// Command berisi data yang dibutuhkan untuk membuat KPIDefinition baru.
type Command struct {
	Name              string
	Code              string
	Category          kpi.KPICategory
	Frequency         kpi.KPIFrequency
	Unit              string
	TargetValue       float64
	WarningThreshold  float64
	CriticalThreshold float64
	Description       string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_kpi.
type Handler struct {
	repo kpi.KPIDefinitionWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIDefinitionWriteRepository) *Handler {
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

	if cmd.Name == "" {
		return kpi.ErrNameEmpty
	}
	if cmd.Code == "" {
		return kpi.ErrCodeEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &kpi.KPIDefinition{
		ID:                id,
		Name:              cmd.Name,
		Code:              cmd.Code,
		Category:          cmd.Category,
		Frequency:         cmd.Frequency,
		Unit:              cmd.Unit,
		TargetValue:       cmd.TargetValue,
		WarningThreshold:  cmd.WarningThreshold,
		CriticalThreshold: cmd.CriticalThreshold,
		IsActive:          true,
		Description:       cmd.Description,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save kpi: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
