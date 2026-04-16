// Package process_resignation menangani command untuk memproses pengunduran diri anggota.
package process_resignation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "process_resignation"

// Command berisi data untuk memproses pengunduran diri.
type Command struct {
	ID       uuid.UUID
	Reason   string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command process_resignation.
type Handler struct {
	repo lifecycle.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke resigned, dan menyimpan.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(lifecycle.ReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement ReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get lifecycle: %w", err)
	}

	if entity.CurrentStatus == lifecycle.MembershipStatusResigned {
		return lifecycle.ErrInvalidTransition
	}

	now := time.Now().UTC()
	previous := entity.CurrentStatus
	entity.PreviousStatus = &previous
	entity.CurrentStatus = lifecycle.MembershipStatusResigned
	entity.LastTransition = lifecycle.TransitionTypeResign
	entity.TransitionReason = cmd.Reason
	entity.ResignedAt = &now
	entity.UpdatedAt = now

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("process resignation: %w", err)
	}
	return nil
}
