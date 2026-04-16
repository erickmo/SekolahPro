// Package get_aml_rule_by_id menangani query untuk mengambil MonitoringRule berdasarkan ID.
package get_aml_rule_by_id

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_rule"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_aml_rule_by_id"

// Query berisi parameter untuk mengambil MonitoringRule berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID          string           `json:"id"`
	RuleCode    string           `json:"rule_code"`
	RuleName    string           `json:"rule_name"`
	RuleType    string           `json:"rule_type"`
	Description string           `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	IsActive    bool             `json:"is_active"`
	AppliesTo   string           `json:"applies_to"`
	AutoAlert   bool             `json:"auto_alert"`
	AutoBlock   bool             `json:"auto_block"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// Handler menangani Query get_aml_rule_by_id.
type Handler struct {
	repo aml_rule.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo aml_rule.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil MonitoringRule berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get aml rule by id: %w", err)
	}

	return &Result{
		ID:          entity.ID.String(),
		RuleCode:    entity.RuleCode,
		RuleName:    entity.RuleName,
		RuleType:    entity.RuleType,
		Description: entity.Description,
		Parameters:  entity.Parameters,
		IsActive:    entity.IsActive,
		AppliesTo:   entity.AppliesTo,
		AutoAlert:   entity.AutoAlert,
		AutoBlock:   entity.AutoBlock,
		CreatedAt:   entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
