// Package reject_insurance_claim menangani command untuk menolak Claim asuransi.
package reject_insurance_claim

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "reject_insurance_claim"

// Command berisi data untuk menolak Claim.
type Command struct {
	ID              uuid.UUID
	RejectionReason string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command reject_insurance_claim.
type Handler struct {
	repo insurance.ClaimWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ClaimWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke rejected, dan menyimpan.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(insurance.ClaimReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement ClaimReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get insurance claim: %w", err)
	}

	if entity.Status == insurance.ClaimStatusPaid {
		return insurance.ErrClaimAlreadyPaid
	}

	entity.Status = insurance.ClaimStatusRejected
	entity.RejectionReason = cmd.RejectionReason
	entity.UpdatedAt = time.Now().UTC()

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("reject insurance claim: %w", err)
	}
	return nil
}
