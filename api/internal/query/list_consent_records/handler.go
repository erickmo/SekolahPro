// Package list_consent_records menangani query untuk mengambil daftar ConsentRecord.
package list_consent_records

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/consent"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_consent_records"

// Query berisi parameter untuk mengambil daftar ConsentRecord.
type Query struct {
	Params pagination.ListParams
	Filter consent.ListFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu ConsentRecord dalam hasil list.
type Item struct {
	ID             string `json:"id"`
	NasabahID      string `json:"nasabah_id"`
	ConsentType    string `json:"consent_type"`
	ConsentPurpose string `json:"consent_purpose"`
	LegalBasis     string `json:"legal_basis"`
	ConsentGiven   bool   `json:"consent_given"`
	ConsentMethod  string `json:"consent_method"`
	Withdrawn      bool   `json:"withdrawn"`
	GivenAt        string `json:"given_at"`
	CreatedAt      string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_consent_records.
type Handler struct {
	repo consent.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo consent.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar ConsentRecord dengan pagination, sorting, dan filter scope organisasi.
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
		return nil, fmt.Errorf("list consent records: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:             e.ID.String(),
			NasabahID:      e.NasabahID.String(),
			ConsentType:    e.ConsentType,
			ConsentPurpose: e.ConsentPurpose,
			LegalBasis:     e.LegalBasis,
			ConsentGiven:   e.ConsentGiven,
			ConsentMethod:  e.ConsentMethod,
			Withdrawn:      e.Withdrawn,
			GivenAt:        e.GivenAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
