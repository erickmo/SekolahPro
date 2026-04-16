// Package create_backup_strategy menangani command untuk membuat BackupStrategy baru.
package create_backup_strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/bcp"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_backup_strategy"

// Command berisi data yang dibutuhkan untuk membuat BackupStrategy baru.
type Command struct {
	StrategyType      string
	Frequency         string
	RetentionDays     int
	StorageLocation   string
	EncryptionEnabled bool
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_backup_strategy.
type Handler struct {
	repo     bcp.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo bcp.WriteRepository, eb eventbus.EventBus) *Handler {
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

	if err := validateStrategyType(cmd.StrategyType); err != nil {
		return err
	}
	if cmd.Frequency == "" {
		return bcp.ErrFrequencyEmpty
	}
	if len(cmd.Frequency) > 100 {
		return bcp.ErrFrequencyTooLong
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &bcp.BackupStrategy{
		ID:                id,
		StrategyType:      bcp.StrategyType(cmd.StrategyType),
		Frequency:         cmd.Frequency,
		RetentionDays:     cmd.RetentionDays,
		StorageLocation:   cmd.StorageLocation,
		EncryptionEnabled: cmd.EncryptionEnabled,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save backup strategy: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, bcp.BackupStrategyCreatedEvent{
		TenantID:     s.TenantID,
		CompanyID:    s.CompanyID,
		ID:           id,
		StrategyType: cmd.StrategyType,
		Frequency:    cmd.Frequency,
	}); err != nil {
		return fmt.Errorf("publish backup_strategy.created: %w", err)
	}

	return nil
}

// validStrategyTypes berisi daftar strategy type yang diterima.
var validStrategyTypes = map[string]bool{
	"full": true, "incremental": true, "differential": true, "snapshot": true,
}

func validateStrategyType(t string) error {
	if t == "" {
		return bcp.ErrStrategyTypeEmpty
	}
	if !validStrategyTypes[t] {
		return bcp.ErrInvalidStrategy
	}
	return nil
}
