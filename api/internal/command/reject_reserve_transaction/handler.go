// Package reject_reserve_transaction menangani command untuk menolak ReserveFundTransaction.
package reject_reserve_transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "reject_reserve_transaction"

// Command berisi data untuk menolak ReserveFundTransaction.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command reject_reserve_transaction.
type Handler struct {
	repo reserve.ReserveFundTransactionWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundTransactionWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle menghapus ReserveFundTransaction yang ditolak.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if err := h.repo.DeleteTransaction(ctx, s, cmd.ID); err != nil {
		return fmt.Errorf("reject reserve transaction: %w", err)
	}
	return nil
}
