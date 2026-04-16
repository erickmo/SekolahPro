// Package list_insurance_policies menangani query untuk mengambil daftar Policy asuransi.
package list_insurance_policies

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_insurance_policies"

// Query berisi parameter untuk mengambil daftar Policy.
type Query struct {
	Params pagination.ListParams
	Filter insurance.PolicyFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu Policy dalam hasil list.
type Item struct {
	ID             string `json:"id"`
	ProductID      string `json:"product_id"`
	NasabahID      string `json:"nasabah_id"`
	PolicyNo       string `json:"policy_no"`
	Status         string `json:"status"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	PremiumAmount  int64  `json:"premium_amount"`
	CoverageAmount int64  `json:"coverage_amount"`
	CreatedAt      string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_insurance_policies.
type Handler struct {
	repo insurance.PolicyReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.PolicyReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar Policy dengan pagination dan filter.
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
		return nil, fmt.Errorf("list insurance policies: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:             e.ID.String(),
			ProductID:      e.ProductID.String(),
			NasabahID:      e.NasabahID.String(),
			PolicyNo:       e.PolicyNo,
			Status:         string(e.Status),
			StartDate:      e.StartDate.Format("2006-01-02T15:04:05Z"),
			EndDate:        e.EndDate.Format("2006-01-02T15:04:05Z"),
			PremiumAmount:  e.PremiumAmount,
			CoverageAmount: e.CoverageAmount,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
