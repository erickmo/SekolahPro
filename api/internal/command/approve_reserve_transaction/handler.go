// Package approve_reserve_transaction menangani command untuk menyetujui ReserveFundTransaction.
package approve_reserve_transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "approve_reserve_transaction"

// Command berisi data untuk menyetujui ReserveFundTransaction.
type Command struct {
	ID uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command approve_reserve_transaction.
type Handler struct {
	repo reserve.ReserveFundTransactionWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundTransactionWriteRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle menyetujui ReserveFundTransaction berdasarkan ID.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	readRepo, ok := h.repo.(reserve.ReserveFundTransactionReadRepository)
	if !ok {
		return fmt.Errorf("repository does not implement ReserveFundTransactionReadRepository")
	}

	entity, err := readRepo.GetTransactionByID(ctx, s, cmd.ID)
	if err != nil {
		return fmt.Errorf("get reserve transaction: %w", err)
	}

	// Transaction sudah diproses saat create, approve hanya untuk audit trail.
	_ = entity

	return nil
}
