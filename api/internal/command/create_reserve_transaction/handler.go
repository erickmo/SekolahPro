// Package create_reserve_transaction menangani command untuk membuat ReserveFundTransaction baru.
package create_reserve_transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/commandbus"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const commandName = "create_reserve_transaction"

// Command berisi data untuk membuat ReserveFundTransaction baru.
type Command struct {
	ReserveFundID   uuid.UUID
	TransactionType reserve.TransactionType
	Amount          int64
	ReferenceNo     string
	Description     string
	ProcessedBy     uuid.UUID
}

func (c Command) CommandName() string { return commandName }

// Handler menangani Command create_reserve_transaction.
type Handler struct {
	fundRepo        reserve.ReserveFundReadRepository
	transactionRepo reserve.ReserveFundTransactionWriteRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(fundRepo reserve.ReserveFundReadRepository, transactionRepo reserve.ReserveFundTransactionWriteRepository) *Handler {
	return &Handler{fundRepo: fundRepo, transactionRepo: transactionRepo}
}

// Handle memvalidasi, membuat entity, dan menyimpan ke repository.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return err
	}

	if cmd.Amount <= 0 {
		return reserve.ErrInvalidAmount
	}

	fund, err := h.fundRepo.GetByID(ctx, s, cmd.ReserveFundID)
	if err != nil {
		return fmt.Errorf("get reserve fund: %w", err)
	}

	if fund.Status == reserve.FundStatusFrozen {
		return reserve.ErrFundFrozen
	}

	if cmd.TransactionType == reserve.TransactionTypeWithdrawal && fund.CurrentBalance < cmd.Amount {
		return reserve.ErrInsufficientBalance
	}

	id := uuid.New()
	now := time.Now().UTC()

	balanceBefore := fund.CurrentBalance
	var balanceAfter int64
	switch cmd.TransactionType {
	case reserve.TransactionTypeContribution, reserve.TransactionTypeInterest:
		balanceAfter = balanceBefore + cmd.Amount
	case reserve.TransactionTypeWithdrawal:
		balanceAfter = balanceBefore - cmd.Amount
	default:
		balanceAfter = balanceBefore + cmd.Amount
	}

	entity := &reserve.ReserveFundTransaction{
		ID:              id,
		ReserveFundID:   cmd.ReserveFundID,
		TransactionType: cmd.TransactionType,
		Amount:          cmd.Amount,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		ReferenceNo:     cmd.ReferenceNo,
		Description:     cmd.Description,
		ProcessedBy:     cmd.ProcessedBy,
		ProcessedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := h.transactionRepo.SaveTransaction(ctx, s, entity); err != nil {
		return fmt.Errorf("save reserve transaction: %w", err)
	}

	commandbus.SetCreatedID(ctx, id)
	return nil
}
