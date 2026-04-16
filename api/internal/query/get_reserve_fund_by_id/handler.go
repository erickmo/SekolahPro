// Package get_reserve_fund_by_id menangani query untuk mengambil ReserveFund berdasarkan ID.
package get_reserve_fund_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/reserve"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_reserve_fund_by_id"

// Query berisi parameter untuk mengambil ReserveFund berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	FundType         string  `json:"fund_type"`
	Status           string  `json:"status"`
	TargetAmount     int64   `json:"target_amount"`
	CurrentBalance   int64   `json:"current_balance"`
	MinimumBalance   int64   `json:"minimum_balance"`
	ContributionPct  float64 `json:"contribution_pct"`
	Description      string  `json:"description"`
	LastCalculatedAt string  `json:"last_calculated_at,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// Handler menangani Query get_reserve_fund_by_id.
type Handler struct {
	repo reserve.ReserveFundReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo reserve.ReserveFundReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil ReserveFund berdasarkan ID.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get reserve fund by id: %w", err)
	}

	result := &Result{
		ID:              entity.ID.String(),
		Name:            entity.Name,
		FundType:        entity.FundType,
		Status:          string(entity.Status),
		TargetAmount:    entity.TargetAmount,
		CurrentBalance:  entity.CurrentBalance,
		MinimumBalance:  entity.MinimumBalance,
		ContributionPct: entity.ContributionPct,
		Description:     entity.Description,
		CreatedAt:       entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if entity.LastCalculatedAt != nil {
		result.LastCalculatedAt = entity.LastCalculatedAt.Format("2006-01-02T15:04:05Z")
	}

	return result, nil
}
