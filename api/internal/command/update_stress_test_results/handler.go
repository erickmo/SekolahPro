// Package update_stress_test_results menangani command untuk mengubah hasil StressTestScenario.
package update_stress_test_results

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_stress_test_results"

// Command berisi data untuk mengubah hasil StressTestScenario.
type Command struct {
	ID           uuid.UUID
	ScenarioStatus kpi.ScenarioStatus
	Results      string
	SimulatedBy  *uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_stress_test_results.
type Handler struct {
	repo kpi.StressTestWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.StressTestWriteRepository) *Handler {
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

	readRepo, ok := h.repo.(kpi.StressTestReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement StressTestReadRepository")
	}

	entity, err := readRepo.GetScenarioByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get stress test: %w", err)
	}

	if entity.ScenarioStatus == kpi.ScenarioStatusRunning {
		return kpi.ErrScenarioRunning
	}

	entity.ScenarioStatus = cmd.ScenarioStatus
	entity.Results = cmd.Results
	entity.UpdatedAt = time.Now().UTC()

	if cmd.SimulatedBy != nil {
		entity.SimulatedBy = cmd.SimulatedBy
		now := time.Now().UTC()
		entity.SimulatedAt = &now
	}

	if cmd.ScenarioStatus == kpi.ScenarioStatusCompleted {
		now := time.Now().UTC()
		entity.CompletedAt = &now
	}

	if err := h.repo.UpdateScenario(ctx, s, entity); err != nil {
		return fmt.Errorf("update stress test results: %w", err)
	}
	return nil
}
