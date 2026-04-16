// Package create_stress_test menangani command untuk membuat StressTestScenario baru.
package create_stress_test

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/kpi"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_stress_test"

// Command berisi data untuk membuat StressTestScenario baru.
type Command struct {
	Name        string
	Description string
	Parameters  string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_stress_test.
type Handler struct {
	repo kpi.StressTestWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo kpi.StressTestWriteRepository) *Handler {
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

	id := uuid.New()
	now := time.Now().UTC()

	entity := &kpi.StressTestScenario{
		ID:             id,
		Name:           cmd.Name,
		Description:    cmd.Description,
		ScenarioStatus: kpi.ScenarioStatusDraft,
		Parameters:     cmd.Parameters,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.repo.SaveScenario(ctx, s, entity); err != nil {
		return fmt.Errorf("save stress test: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
