// Package list_reserve_transactions menangani query untuk mengambil daftar ReserveFundTransaction.
package list_reserve_transactions

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_reserve_transactions"

// Query berisi parameter untuk mengambil daftar ReserveFundTransaction.
type Query struct {
	Params pagination.ListParams
	Filter reserve.ReserveFundTransactionFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu ReserveFundTransaction dalam hasil list.
type Item struct {
	ID              string `json:"id"`
	ReserveFundID   string `json:"reserve_fund_id"`
	TransactionType string `json:"transaction_type"`
	Amount          int64  `json:"amount"`
	BalanceBefore   int64  `json:"balance_before"`
	BalanceAfter    int64  `json:"balance_after"`
	ReferenceNo     string `json:"reference_no"`
	Description     string `json:"description"`
	ProcessedBy     string `json:"processed_by"`
	ProcessedAt     string `json:"processed_at"`
	CreatedAt       string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_reserve_transactions.
type Handler struct {
	repo reserve.ReserveFundTransactionReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundTransactionReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar ReserveFundTransaction dengan pagination dan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.ListTransactions(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list reserve transactions: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:              e.ID.String(),
			ReserveFundID:   e.ReserveFundID.String(),
			TransactionType: string(e.TransactionType),
			Amount:          e.Amount,
			BalanceBefore:   e.BalanceBefore,
			BalanceAfter:    e.BalanceAfter,
			ReferenceNo:     e.ReferenceNo,
			Description:     e.Description,
			ProcessedBy:     e.ProcessedBy.String(),
			ProcessedAt:     e.ProcessedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
