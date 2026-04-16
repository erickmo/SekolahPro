// Package freeze_reserve_fund menangani command untuk membekukan ReserveFund.
package freeze_reserve_fund

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "freeze_reserve_fund"

// Command berisi data untuk membekukan ReserveFund.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command freeze_reserve_fund.
type Handler struct {
	repo reserve.ReserveFundWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah status ke frozen, dan menyimpan.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(reserve.ReserveFundReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement ReserveFundReadRepository")
	}

	entity, err := readRepo.GetByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get reserve fund: %w", err)
	}

	if entity.Status == reserve.FundStatusFrozen {
		return reserve.ErrFundFrozen
	}

	entity.Status = reserve.FundStatusFrozen
	entity.UpdatedAt = time.Now().UTC()

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("freeze reserve fund: %w", err)
	}
	return nil
}
