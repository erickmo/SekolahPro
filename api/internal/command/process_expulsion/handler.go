// Package process_expulsion menangani command untuk memproses pencabutan keanggotaan.
package process_expulsion

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "process_expulsion"

// Command berisi data untuk memproses pencabutan keanggotaan.
type Command struct {
	ID        uuid.UUID
	RevokedBy uuid.UUID
	Reason    string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command process_expulsion.
type Handler struct {
	repo lifecycle.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo lifecycle.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke revoked, dan menyimpan.
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

	if entity.CurrentStatus == lifecycle.MembershipStatusRevoked {
		return lifecycle.ErrInvalidTransition
	}

	now := time.Now().UTC()
	previous := entity.CurrentStatus
	entity.PreviousStatus = &previous
	entity.CurrentStatus = lifecycle.MembershipStatusRevoked
	entity.LastTransition = lifecycle.TransitionTypeRevoke
	entity.TransitionReason = cmd.Reason
	entity.RevokedAt = &now
	entity.RevokedBy = &cmd.RevokedBy
	entity.UpdatedAt = now

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("process expulsion: %w", err)
	}
	return nil
}
