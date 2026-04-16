// Package delete_reserve_fund menangani command untuk menghapus ReserveFund.
package delete_reserve_fund

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "delete_reserve_fund"

// Command berisi data untuk menghapus ReserveFund.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command delete_reserve_fund.
type Handler struct {
	repo reserve.ReserveFundWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle menghapus ReserveFund berdasarkan ID.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if err := h.repo.Delete(ctx, s, cmd.ID); err != nil {
		return fmt.Errorf("delete reserve fund: %w", err)
	}
	return nil
}
