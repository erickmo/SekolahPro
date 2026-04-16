// Package get_insurance_product_by_id menangani query untuk mengambil Product berdasarkan ID.
package get_insurance_product_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_insurance_product_by_id"

// Query berisi parameter untuk mengambil Product berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Code             string `json:"code"`
	ProductType      string `json:"product_type"`
	Status           string `json:"status"`
	ProviderName     string `json:"provider_name"`
	PremiumAmount    int64  `json:"premium_amount"`
	CoverageAmount   int64  `json:"coverage_amount"`
	PremiumFrequency string `json:"premium_frequency"`
	TermMonths       int    `json:"term_months"`
	Description      string `json:"description"`
	TermsConditions  string `json:"terms_conditions"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// Handler menangani Query get_insurance_product_by_id.
type Handler struct {
	repo insurance.ProductReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ProductReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil Product berdasarkan ID.
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
		return nil, fmt.Errorf("get insurance product by id: %w", err)
	}

	return &Result{
		ID:               entity.ID.String(),
		Name:             entity.Name,
		Code:             entity.Code,
		ProductType:      string(entity.ProductType),
		Status:           string(entity.Status),
		ProviderName:     entity.ProviderName,
		PremiumAmount:    entity.PremiumAmount,
		CoverageAmount:   entity.CoverageAmount,
		PremiumFrequency: entity.PremiumFrequency,
		TermMonths:       entity.TermMonths,
		Description:      entity.Description,
		TermsConditions:  entity.TermsConditions,
		CreatedAt:        entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
