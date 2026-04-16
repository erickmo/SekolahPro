// Package create_reserve_fund menangani command untuk membuat ReserveFund baru.
package create_reserve_fund

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_reserve_fund"

// Command berisi data yang dibutuhkan untuk membuat ReserveFund baru.
type Command struct {
	Name            string
	FundType        string
	TargetAmount    int64
	MinimumBalance  int64
	ContributionPct float64
	Description     string
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_reserve_fund.
type Handler struct {
	repo reserve.ReserveFundWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundWriteRepository) *Handler {
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

	if cmd.Name == "" {
		return reserve.ErrNameEmpty
	}
	if cmd.FundType == "" {
		return reserve.ErrFundTypeEmpty
	}

	id := uuid.New()
	now := time.Now().UTC()

	entity := &reserve.ReserveFund{
		ID:              id,
		Name:            cmd.Name,
		FundType:        cmd.FundType,
		Status:          reserve.FundStatusActive,
		TargetAmount:    cmd.TargetAmount,
		CurrentBalance:  0,
		MinimumBalance:  cmd.MinimumBalance,
		ContributionPct: cmd.ContributionPct,
		Description:     cmd.Description,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.repo.Save(ctx, s, entity); err != nil {
		return fmt.Errorf("save reserve fund: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
