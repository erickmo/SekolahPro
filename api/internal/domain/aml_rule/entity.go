// Package aml_rule adalah domain CQRS untuk AML monitoring rules (ADR-K027).
//
// Mengelola aturan monitoring transaksi untuk deteksi aktivitas mencurigakan:
// threshold rules, velocity rules, pattern detection, dan anomaly detection.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package aml_rule

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// MonitoringRule adalah entity untuk aturan monitoring AML.
type MonitoringRule struct {
	ID          uuid.UUID       `db:"id"           json:"id"`
	RuleCode    string          `db:"rule_code"    json:"rule_code"`
	RuleName    string          `db:"rule_name"    json:"rule_name"`
	RuleType    string          `db:"rule_type"    json:"rule_type"`
	Description string          `db:"description"  json:"description,omitempty"`
	Parameters  json.RawMessage `db:"parameters"   json:"parameters"`
	IsActive    bool            `db:"is_active"    json:"is_active"`
	AppliesTo   string          `db:"applies_to"   json:"applies_to"`
	AutoAlert   bool            `db:"auto_alert"   json:"auto_alert"`
	AutoBlock   bool            `db:"auto_block"   json:"auto_block"`
	CreatedAt   time.Time       `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"   json:"updated_at"`
	DeletedAt   *time.Time      `db:"deleted_at"   json:"-"`
}

// Rule type constants.
const (
	RuleTypeThreshold = "threshold"
	RuleTypePattern   = "pattern"
	RuleTypeVelocity  = "velocity"
	RuleTypeAnomaly   = "anomaly"
)

// Applies to constants.
const (
	AppliesToAll         = "all"
	AppliesToHighRiskOnly = "high_risk_only"
	AppliesToPEPOnly     = "pep_only"
)

// ListFilter menyimpan filter opsional untuk List query.
type ListFilter struct {
	RuleType *string
	IsActive *bool
}

// WriteRepository mendefinisikan operasi write untuk monitoring rules.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *MonitoringRule) error
	Update(ctx context.Context, s scope.Scope, e *MonitoringRule) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk monitoring rules.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*MonitoringRule, error)
	GetByRuleCode(ctx context.Context, s scope.Scope, ruleCode string) (*MonitoringRule, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*MonitoringRule, int, error)
}
