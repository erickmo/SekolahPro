// Package create_lifecycle_event menangani command untuk membuat MembershipLifecycle baru.
package create_lifecycle_event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_lifecycle_event"

// Command berisi data yang dibutuhkan untuk membuat MembershipLifecycle baru.
type Command struct {
	NasabahID         uuid.UUID
	MembershipNo      string
	TransitionReason  string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_lifecycle_event.
type Handler struct {
	repo lifecycle.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.WriteRepository) *Handler {
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

	if cmd.NasabahID == uuid.Nil {
		return lifecycle.ErrNasabahIDEmpty
	}
	if cmd.MembershipNo == "" {
		return lifecycle.ErrMembershipNoEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &lifecycle.MembershipLifecycle{
		ID:               id,
		NasabahID:        cmd.NasabahID,
		CurrentStatus:    lifecycle.MembershipStatusPending,
		LastTransition:   lifecycle.TransitionTypeApply,
		TransitionReason: cmd.TransitionReason,
		MembershipNo:     cmd.MembershipNo,
		AppliedAt:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save lifecycle: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
