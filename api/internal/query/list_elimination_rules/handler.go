// Package list_elimination_rules menangani query untuk mengambil daftar EliminationRule.
package list_elimination_rules

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/branch_report"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_elimination_rules"

// Query berisi parameter untuk mengambil daftar EliminationRule.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu EliminationRule dalam hasil list.
type Item struct {
	ID           string `json:"id"`
	RuleName     string `json:"rule_name"`
	RuleType     string `json:"rule_type"`
	FromBranchID string `json:"from_branch_id"`
	ToBranchID   string `json:"to_branch_id"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_elimination_rules.
type Handler struct {
	repo branch_report.EliminationRuleReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo branch_report.EliminationRuleReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar EliminationRule dengan pagination dan sorting.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	rules, total, err := h.repo.ListRules(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list elimination rules: %w", err)
	}

	items := make([]*Item, 0, len(rules))
	for _, r := range rules {
		items = append(items, &Item{
			ID:           r.ID.String(),
			RuleName:     r.RuleName,
			RuleType:     string(r.RuleType),
			FromBranchID: r.FromBranchID.String(),
			ToBranchID:   r.ToBranchID.String(),
			IsActive:     r.IsActive,
			Description:  r.Description,
			CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
