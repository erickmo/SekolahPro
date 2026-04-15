// Package update_example menangani command untuk mengupdate Example yang sudah ada.
package update_example

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_example"

// Command berisi data yang dibutuhkan untuk mengupdate Example.
type Command struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsActive    bool
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_example.
type Handler struct {
	readRepo  example.ReadRepository
	writeRepo example.WriteRepository
	eventBus  eventbus.EventBus
}

// NewHandler membuat instance Handler baru.
func NewHandler(rr example.ReadRepository, wr example.WriteRepository, eb eventbus.EventBus) *Handler {
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
	if cmd.Name == "" {
		return example.ErrNameEmpty
	}
	if len(cmd.Name) > 255 {
		return example.ErrNameTooLong
	}

	entity, err := h.readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get example for update: %w", err)
	}

	entity.Name = cmd.Name
	entity.Description = cmd.Description
	entity.IsActive = cmd.IsActive
	entity.UpdatedAt = time.Now().UTC()

	if err := h.writeRepo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update example: %w", err)
	}

	if err := h.eventBus.Publish(ctx, example.ExampleUpdatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        entity.ID,
		Name:      entity.Name,
	}); err != nil {
		return fmt.Errorf("publish example.updated: %w", err)
	}

	return nil
}
