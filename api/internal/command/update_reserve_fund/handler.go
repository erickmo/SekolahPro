// Package update_reserve_fund menangani command untuk mengubah ReserveFund.
package update_reserve_fund

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "update_reserve_fund"

// Command berisi data untuk mengubah ReserveFund.
type Command struct {
	ID              uuid.UUID
	Name            string
	FundType        string
	TargetAmount    int64
	MinimumBalance  int64
	ContributionPct float64
	Description     string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command update_reserve_fund.
type Handler struct {
	repo reserve.ReserveFundWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil entity, mengubah field, dan menyimpan kembali.
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

	entity.Name = cmd.Name
	entity.FundType = cmd.FundType
	entity.TargetAmount = cmd.TargetAmount
	entity.MinimumBalance = cmd.MinimumBalance
	entity.ContributionPct = cmd.ContributionPct
	entity.Description = cmd.Description
	entity.UpdatedAt = time.Now().UTC()

	if err := h.repo.Update(ctx, s, entity); err != nil {
		return fmt.Errorf("update reserve fund: %w", err)
	}
	return nil
}
