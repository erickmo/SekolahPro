// Package review_insurance_claim menangani command untuk meninjau Claim asuransi.
package review_insurance_claim

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "review_insurance_claim"

// Command berisi data untuk meninjau Claim.
type Command struct {
	ID         uuid.UUID
	ReviewedBy uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command review_insurance_claim.
type Handler struct {
	repo insurance.ClaimWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ClaimWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke under_review, dan menyimpan.
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

	if entity.Status != insurance.ClaimStatusSubmitted {
		return fmt.Errorf("claim tidak bisa ditinjau: status %s", entity.Status)
	}

	now := time.Now().UTC()
	entity.Status = insurance.ClaimStatusUnderReview
	entity.ReviewedBy = &cmd.ReviewedBy
	entity.ReviewedAt = &now
	entity.UpdatedAt = now

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("review insurance claim: %w", err)
	}
	return nil
}
