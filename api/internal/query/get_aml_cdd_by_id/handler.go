// Package get_aml_cdd_by_id menangani query untuk mengambil CDD record berdasarkan ID.
package get_aml_cdd_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_aml_cdd_by_id"

// Query berisi parameter untuk mengambil CDD record berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                     string           `json:"id"`
	NasabahID              string           `json:"nasabah_id"`
	RiskLevel              string           `json:"risk_level"`
	RiskScore              int              `json:"risk_score"`
	RiskCategory           aml_cdd.RiskCategory `json:"risk_category"`
	CDDLevel               string           `json:"cdd_level"`
	CDDPurpose             string           `json:"cdd_purpose,omitempty"`
	SourceOfFunds          *string          `json:"source_of_funds,omitempty"`
	SourceOfWealth         *string          `json:"source_of_wealth,omitempty"`
	IsPEP                  bool             `json:"is_pep"`
	PEPType                string           `json:"pep_type"`
	PEPPosition            *string          `json:"pep_position,omitempty"`
	PEPCountry             *string          `json:"pep_country,omitempty"`
	PEPRelationship        string           `json:"pep_relationship"`
	BeneficialOwnerName    *string          `json:"beneficial_owner_name,omitempty"`
	BeneficialOwnerIDNo    *string          `json:"beneficial_owner_id_no,omitempty"`
	BeneficialOwnershipPct *float64         `json:"beneficial_ownership_pct,omitempty"`
	BOVerified             bool             `json:"bo_verified"`
	NextReviewDate         *string          `json:"next_review_date,omitempty"`
	ReviewCount            int              `json:"review_count"`
	Status                 string           `json:"status"`
	CreatedAt              string           `json:"created_at"`
	UpdatedAt              string           `json:"updated_at"`
	CreatedBy              string           `json:"created_by"`
	UpdatedBy              string           `json:"updated_by"`
}

// Handler menangani Query get_aml_cdd_by_id.
type Handler struct {
	repo aml_cdd.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo aml_cdd.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil CDD record berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get aml cdd by id: %w", err)
	}

	result := &Result{
		ID:                     entity.ID.String(),
		NasabahID:              entity.NasabahID.String(),
		RiskLevel:              entity.RiskLevel,
		RiskScore:              entity.RiskScore,
		CDDLevel:               entity.CDDLevel,
		CDDPurpose:             entity.CDDPurpose,
		SourceOfFunds:          entity.SourceOfFunds,
		SourceOfWealth:         entity.SourceOfWealth,
		IsPEP:                  entity.IsPEP,
		PEPType:                entity.PEPType,
		PEPPosition:            entity.PEPPosition,
		PEPCountry:             entity.PEPCountry,
		PEPRelationship:        entity.PEPRelationship,
		BeneficialOwnerName:    entity.BeneficialOwnerName,
		BeneficialOwnerIDNo:    entity.BeneficialOwnerIDNo,
		BeneficialOwnershipPct: entity.BeneficialOwnershipPct,
		BOVerified:             entity.BOVerified,
		ReviewCount:            entity.ReviewCount,
		Status:                 entity.Status,
		CreatedAt:              entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:              entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		CreatedBy:              entity.CreatedBy.String(),
		UpdatedBy:              entity.UpdatedBy.String(),
	}
	if entity.NextReviewDate != nil {
		nrd := entity.NextReviewDate.Format("2006-01-02T15:04:05Z")
		result.NextReviewDate = &nrd
	}

	return result, nil
}
