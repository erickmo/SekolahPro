// Package create_example menangani command untuk membuat Example baru.
package create_example

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/example"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/eventbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_example"

// Command berisi data yang dibutuhkan untuk membuat Example baru.
type Command struct {
	Name        string
	Description string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_example.
type Handler struct {
	repo     example.WriteRepository
	eventBus eventbus.EventBus
}

// NewHandler membuat instance Handler baru dengan dependensi yang diinjeksikan.
func NewHandler(repo example.WriteRepository, eb eventbus.EventBus) *Handler {
	return &Handler{repo: repo, eventBus: eb}
}

// Handle memvalidasi command, membuat entity, menyimpan ke repository,
// dan mempublikasikan domain event.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	// Application layer membaca Scope dari context dan meneruskan ke repo sebagai parameter eksplisit.
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

	id := uuid.New()
	now := time.Now().UTC()

	entity := &example.Example{
		ID:          id,
		Name:        cmd.Name,
		Description: cmd.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Scope diteruskan sebagai parameter eksplisit — bukan dari context.
	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save example: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)

	if err := h.eventBus.Publish(ctx, example.ExampleCreatedEvent{
		TenantID:  s.TenantID,
		CompanyID: s.CompanyID,
		ID:        id,
		Name:      cmd.Name,
	}); err != nil {
		return fmt.Errorf("publish example.created: %w", err)
	}

	return nil
}
