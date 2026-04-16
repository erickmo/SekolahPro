// Package list_aml_cdds menangani query untuk mengambil daftar CDD record.
package list_aml_cdds

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_aml_cdds"

// Query berisi parameter untuk mengambil daftar CDD record.
type Query struct {
	Params pagination.ListParams
	Filter aml_cdd.ListFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu CDD record dalam hasil list.
type Item struct {
	ID         string `json:"id"`
	NasabahID  string `json:"nasabah_id"`
	RiskLevel  string `json:"risk_level"`
	RiskScore  int    `json:"risk_score"`
	CDDLevel   string `json:"cdd_level"`
	IsPEP      bool   `json:"is_pep"`
	Status     string `json:"status"`
	ReviewCount int   `json:"review_count"`
	CreatedAt  string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_aml_cdds.
type Handler struct {
	repo aml_cdd.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo aml_cdd.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar CDD record dengan pagination, sorting, dan filter scope organisasi.
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
		return nil, fmt.Errorf("list aml cdds: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:          e.ID.String(),
			NasabahID:   e.NasabahID.String(),
			RiskLevel:   e.RiskLevel,
			RiskScore:   e.RiskScore,
			CDDLevel:    e.CDDLevel,
			IsPEP:       e.IsPEP,
			Status:      e.Status,
			ReviewCount: e.ReviewCount,
			CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
