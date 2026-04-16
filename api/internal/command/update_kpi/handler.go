// Package update_kpi menangani command untuk mengubah KPIDefinition.
package update_kpi

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_kpi"

// Command berisi data untuk mengubah KPIDefinition.
type Command struct {
	ID                uuid.UUID
	Name              string
	Code              string
	Category          kpi.KPICategory
	Frequency         kpi.KPIFrequency
	Unit              string
	TargetValue       float64
	WarningThreshold  float64
	CriticalThreshold float64
	IsActive          bool
	Description       string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_kpi.
type Handler struct {
	repo kpi.KPIDefinitionWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.KPIDefinitionWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah field, dan menyimpan kembali.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(kpi.KPIDefinitionReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement KPIDefinitionReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get kpi: %w", err)
	}

	entity.Name = cmd.Name
	entity.Code = cmd.Code
	entity.Category = cmd.Category
	entity.Frequency = cmd.Frequency
	entity.Unit = cmd.Unit
	entity.TargetValue = cmd.TargetValue
	entity.WarningThreshold = cmd.WarningThreshold
	entity.CriticalThreshold = cmd.CriticalThreshold
	entity.IsActive = cmd.IsActive
	entity.Description = cmd.Description
	entity.UpdatedAt = time.Now().UTC()

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update kpi: %w", err)
	}
	return nil
}
