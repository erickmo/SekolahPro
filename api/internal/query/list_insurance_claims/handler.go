// Package list_insurance_claims menangani query untuk mengambil daftar Claim asuransi.
package list_insurance_claims

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_insurance_claims"

// Query berisi parameter untuk mengambil daftar Claim.
type Query struct {
	Params pagination.ListParams
	Filter insurance.ClaimFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu Claim dalam hasil list.
type Item struct {
	ID            string `json:"id"`
	PolicyID      string `json:"policy_id"`
	ClaimType     string `json:"claim_type"`
	Status        string `json:"status"`
	ClaimAmount   int64  `json:"claim_amount"`
	ApprovedAmount *int64 `json:"approved_amount,omitempty"`
	IncidentDate  string `json:"incident_date"`
	SubmittedAt   string `json:"submitted_at"`
	SubmittedBy   string `json:"submitted_by"`
	Description   string `json:"description"`
	CreatedAt     string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_insurance_claims.
type Handler struct {
	repo insurance.ClaimReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo insurance.ClaimReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar Claim dengan pagination dan filter.
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
		return nil, fmt.Errorf("list insurance claims: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:             e.ID.String(),
			PolicyID:       e.PolicyID.String(),
			ClaimType:      string(e.ClaimType),
			Status:         string(e.Status),
			ClaimAmount:    e.ClaimAmount,
			ApprovedAmount: e.ApprovedAmount,
			IncidentDate:   e.IncidentDate.Format("2006-01-02T15:04:05Z"),
			SubmittedAt:    e.SubmittedAt.Format("2006-01-02T15:04:05Z"),
			SubmittedBy:    e.SubmittedBy.String(),
			Description:    e.Description,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
