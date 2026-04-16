// Package list_reserve_funds menangani query untuk mengambil daftar ReserveFund.
package list_reserve_funds

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_reserve_funds"

// Query berisi parameter untuk mengambil daftar ReserveFund.
type Query struct {
	Params pagination.ListParams
	Filter reserve.ReserveFundFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu ReserveFund dalam hasil list.
type Item struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	FundType        string  `json:"fund_type"`
	Status          string  `json:"status"`
	TargetAmount    int64   `json:"target_amount"`
	CurrentBalance  int64   `json:"current_balance"`
	MinimumBalance  int64   `json:"minimum_balance"`
	ContributionPct float64 `json:"contribution_pct"`
	CreatedAt       string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_reserve_funds.
type Handler struct {
	repo reserve.ReserveFundReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar ReserveFund dengan pagination dan filter.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, qry.Filter, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list reserve funds: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:              e.ID.String(),
			Name:            e.Name,
			FundType:        e.FundType,
			Status:          string(e.Status),
			TargetAmount:    e.TargetAmount,
			CurrentBalance:  e.CurrentBalance,
			MinimumBalance:  e.MinimumBalance,
			ContributionPct: e.ContributionPct,
			CreatedAt:       e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
