// Package update_dsr_status menangani command untuk mengubah status DataSubjectRequest.
package update_dsr_status

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dsr"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_dsr_status"

// Command berisi data untuk mengubah status DataSubjectRequest.
type Command struct {
	ID             uuid.UUID
	Status         dsr.RequestStatus
	VerifiedBy     *uuid.UUID
	RejectionReason string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_dsr_status.
type Handler struct {
	repo dsr.WriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo dsr.WriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status, dan menyimpan kembali.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.Status == "" {
		return dsr.ErrInvalidStatus
	}

	existing, err := h.getExisting(ctx, s, cmd.ID)
	if err != nil {
		return err
	}

	if existing.Status == dsr.RequestStatusCompleted || existing.Status == dsr.RequestStatusCancelled {
		return dsr.ErrAlreadyCompleted
	}

	existing.Status = cmd.Status
	existing.UpdatedAt = time.Now().UTC()

	if cmd.Status == dsr.RequestStatusVerified && cmd.VerifiedBy != nil {
		now := time.Now().UTC()
		existing.VerifiedBy = cmd.VerifiedBy
		existing.VerifiedAt = &now
	}

	if cmd.Status == dsr.RequestStatusRejected {
		existing.RejectionReason = cmd.RejectionReason
	}

	if err := h.repo.Update(ctx, s, existing); err != nil {
		return fmt.Errorf("update dsr status: %w", err)
	}

	return nil
}

func (h *Handler) getExisting(ctx context.Context, s scope.Scope, id uuid.UUID) (*dsr.DataSubjectRequest, error) {
	readRepo, ok := h.repo.(dsr.ReadRepository)
	if !ok {
		return nil, fmt.Errorf("repository does not implement ReadRepository")
	}
	entity, err := readRepo.GetByID(ctx, s, id)
	if err != nil {
		return nil, fmt.Errorf("get dsr: %w", err)
	}
	return entity, nil
}
