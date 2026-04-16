// Package list_insurance_products menangani query untuk mengambil daftar Product asuransi.
package list_insurance_products

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_insurance_products"

// Query berisi parameter untuk mengambil daftar Product.
type Query struct {
	Params pagination.ListParams
	Filter insurance.ProductFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu Product dalam hasil list.
type Item struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	ProductType    string `json:"product_type"`
	Status         string `json:"status"`
	ProviderName   string `json:"provider_name"`
	PremiumAmount  int64  `json:"premium_amount"`
	CoverageAmount int64  `json:"coverage_amount"`
	CreatedAt      string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_insurance_products.
type Handler struct {
	repo insurance.ProductReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ProductReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar Product dengan pagination dan filter.
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
		return nil, fmt.Errorf("list insurance products: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:             e.ID.String(),
			Name:           e.Name,
			Code:           e.Code,
			ProductType:    string(e.ProductType),
			Status:         string(e.Status),
			ProviderName:   e.ProviderName,
			PremiumAmount:  e.PremiumAmount,
			CoverageAmount: e.CoverageAmount,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
