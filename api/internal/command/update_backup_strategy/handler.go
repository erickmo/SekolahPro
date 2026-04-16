// Package update_backup_strategy menangani command untuk mengupdate BackupStrategy yang sudah ada.
package update_backup_strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/bcp"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_backup_strategy"

// Command berisi data yang dibutuhkan untuk mengupdate BackupStrategy.
type Command struct {
	ID                uuid.UUID
	StrategyType      string
	Frequency         string
	RetentionDays     int
	StorageLocation   string
	EncryptionEnabled bool
	IsActive          bool
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_backup_strategy.
type Handler struct {
	readRepo  bcp.ReadRepository
	writeRepo bcp.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr bcp.ReadRepository, wr bcp.WriteRepository, eb eventbus.EventBus) *Handler {
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

	if err := validateStrategyType(cmd.StrategyType); err != nil {
		return err
	}
	if cmd.Frequency == "" {
		return bcp.ErrFrequencyEmpty
	}
	if len(cmd.Frequency) > 100 {
		return bcp.ErrFrequencyTooLong
	}
	if len(cmd.StorageLocation) > 500 {
		return bcp.ErrStorageTooLong
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get backup strategy for update: %w", err)
	}

	entity.StrategyType = bcp.StrategyType(cmd.StrategyType)
	entity.Frequency = cmd.Frequency
	entity.RetentionDays = cmd.RetentionDays
	entity.StorageLocation = cmd.StorageLocation
	entity.EncryptionEnabled = cmd.EncryptionEnabled
	entity.IsActive = cmd.IsActive
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update backup strategy: %w", err)
	}

	if err := h.eventBus.Publish(ctx, bcp.BackupStrategyUpdatedEvent{
		TenantID:     s.TenantID,
		CompanyID:    s.CompanyID,
		ID:           entity.ID,
		StrategyType: string(entity.StrategyType),
		Frequency:    entity.Frequency,
	}); err != nil {
		return fmt.Errorf("publish backup_strategy.updated: %w", err)
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
