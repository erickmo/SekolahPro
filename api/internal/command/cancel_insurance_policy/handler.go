// Package cancel_insurance_policy menangani command untuk membatalkan Policy asuransi.
package cancel_insurance_policy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "cancel_insurance_policy"

// Command berisi data untuk membatalkan Policy.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command cancel_insurance_policy.
type Handler struct {
	repo insurance.PolicyWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.PolicyWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke cancelled, dan menyimpan.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(insurance.PolicyReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement PolicyReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get insurance policy: %w", err)
	}

	entity.Status = insurance.PolicyStatusCancelled
	entity.UpdatedAt = time.Now().UTC()

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("cancel insurance policy: %w", err)
	}
	return nil
}
