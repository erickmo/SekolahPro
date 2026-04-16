// Package get_insurance_policy_by_id menangani query untuk mengambil Policy berdasarkan ID.
package get_insurance_policy_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_insurance_policy_by_id"

// Query berisi parameter untuk mengambil Policy berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                  string `json:"id"`
	ProductID           string `json:"product_id"`
	NasabahID           string `json:"nasabah_id"`
	PolicyNo            string `json:"policy_no"`
	Status              string `json:"status"`
	StartDate           string `json:"start_date"`
	EndDate             string `json:"end_date"`
	PremiumAmount       int64  `json:"premium_amount"`
	CoverageAmount      int64  `json:"coverage_amount"`
	LastPremiumPaid     string `json:"last_premium_paid,omitempty"`
	NextPremiumDue      string `json:"next_premium_due,omitempty"`
	BeneficiaryName     string `json:"beneficiary_name"`
	BeneficiaryRelation string `json:"beneficiary_relation"`
	IssuedAt            string `json:"issued_at,omitempty"`
	IssuedBy            string `json:"issued_by,omitempty"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

// Handler menangani Query get_insurance_policy_by_id.
type Handler struct {
	repo insurance.PolicyReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.PolicyReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil Policy berdasarkan ID.
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
		return nil, fmt.Errorf("get insurance policy by id: %w", err)
	}

	result := &Result{
		ID:                  entity.ID.String(),
		ProductID:           entity.ProductID.String(),
		NasabahID:           entity.NasabahID.String(),
		PolicyNo:            entity.PolicyNo,
		Status:              string(entity.Status),
		StartDate:           entity.StartDate.Format("2006-01-02T15:04:05Z"),
		EndDate:             entity.EndDate.Format("2006-01-02T15:04:05Z"),
		PremiumAmount:       entity.PremiumAmount,
		CoverageAmount:      entity.CoverageAmount,
		BeneficiaryName:     entity.BeneficiaryName,
		BeneficiaryRelation: entity.BeneficiaryRelation,
		CreatedAt:           entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if entity.LastPremiumPaid != nil {
		result.LastPremiumPaid = entity.LastPremiumPaid.Format("2006-01-02T15:04:05Z")
	}
	if entity.NextPremiumDue != nil {
		result.NextPremiumDue = entity.NextPremiumDue.Format("2006-01-02T15:04:05Z")
	}
	if entity.IssuedAt != nil {
		result.IssuedAt = entity.IssuedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.IssuedBy != nil {
		result.IssuedBy = entity.IssuedBy.String()
	}

	return result, nil
}
