// Package list_aml_rules menangani query untuk mengambil daftar MonitoringRule.
package list_aml_rules

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_aml_rules"

// Query berisi parameter untuk mengambil daftar MonitoringRule.
type Query struct {
	Params pagination.ListParams
	Filter aml_rule.ListFilter
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu MonitoringRule dalam hasil list.
type Item struct {
	ID        string `json:"id"`
	RuleCode  string `json:"rule_code"`
	RuleName  string `json:"rule_name"`
	RuleType  string `json:"rule_type"`
	IsActive  bool   `json:"is_active"`
	AppliesTo string `json:"applies_to"`
	AutoAlert bool   `json:"auto_alert"`
	AutoBlock bool   `json:"auto_block"`
	CreatedAt string `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_aml_rules.
type Handler struct {
	repo aml_rule.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo aml_rule.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar MonitoringRule dengan pagination, sorting, dan filter scope organisasi.
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
		return nil, fmt.Errorf("list aml rules: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:        e.ID.String(),
			RuleCode:  e.RuleCode,
			RuleName:  e.RuleName,
			RuleType:  e.RuleType,
			IsActive:  e.IsActive,
			AppliesTo: e.AppliesTo,
			AutoAlert: e.AutoAlert,
			AutoBlock: e.AutoBlock,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
