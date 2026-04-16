// Package reactivate_member menangani command untuk mengaktifkan kembali anggota.
package reactivate_member

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "reactivate_member"

// Command berisi data untuk mengaktifkan kembali anggota.
type Command struct {
	ID     uuid.UUID
	Reason string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command reactivate_member.
type Handler struct {
	repo lifecycle.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke active, dan menyimpan.
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

	if entity.CurrentStatus == lifecycle.MembershipStatusActive {
		return lifecycle.ErrAlreadyActive
	}

	// Hanya suspended member yang bisa diaktifkan kembali.
	if entity.CurrentStatus != lifecycle.MembershipStatusSuspended {
		return lifecycle.ErrCannotReactivate
	}

	now := time.Now().UTC()
	previous := entity.CurrentStatus
	entity.PreviousStatus = &previous
	entity.CurrentStatus = lifecycle.MembershipStatusActive
	entity.LastTransition = lifecycle.TransitionTypeReactivate
	entity.TransitionReason = cmd.Reason
	entity.ActivatedAt = &now
	entity.UpdatedAt = now

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("reactivate member: %w", err)
	}
	return nil
}
