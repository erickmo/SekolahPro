// Package aml_alert adalah domain CQRS untuk AML transaction monitoring alerts (ADR-K027).
//
// Mengelola alert dari transaction monitoring: suspicious transactions,
// threshold exceeded, unusual patterns, dan investigation outcomes.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package aml_alert

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// TransactionAlert adalah entity untuk alert transaksi AML.
type TransactionAlert struct {
	ID                    uuid.UUID       `db:"id"                      json:"id"`
	NasabahID             uuid.UUID       `db:"nasabah_id"              json:"nasabah_id"`
	TransaksiID           *uuid.UUID      `db:"transaksi_id"            json:"transaksi_id,omitempty"`
	RuleID                uuid.UUID       `db:"rule_id"                 json:"rule_id"`
	RuleType              string          `db:"rule_type"               json:"rule_type"`
	AlertType             string          `db:"alert_type"              json:"alert_type"`
	Severity              string          `db:"severity"                json:"severity"`
	Description           string          `db:"description"             json:"description"`
	TransactionDetails    json.RawMessage `db:"transaction_details"     json:"transaction_details"`
	InvestigatedBy        *uuid.UUID      `db:"investigated_by"         json:"investigated_by,omitempty"`
	InvestigationNotes    *string         `db:"investigation_notes"     json:"investigation_notes,omitempty"`
	InvestigationOutcome  *string         `db:"investigation_outcome"   json:"investigation_outcome,omitempty"`
	LTKMFiled            bool            `db:"ltkm_filed"              json:"ltkm_filed"`
	LTKMReference        *string         `db:"ltkm_reference"          json:"ltkm_reference,omitempty"`
	Status                string          `db:"status"                  json:"status"`
	ResolvedAt            *time.Time      `db:"resolved_at"             json:"resolved_at,omitempty"`
	DetectedAt            time.Time       `db:"detected_at"             json:"detected_at"`
	CreatedAt             time.Time       `db:"created_at"              json:"created_at"`
	DeletedAt             *time.Time      `db:"deleted_at"              json:"-"`
}

// Alert type constants.
const (
	AlertTypeSuspicious       = "suspicious"
	AlertTypeThresholdExceeded = "threshold_exceeded"
	AlertTypeUnusualPattern   = "unusual_pattern"
	AlertTypeStructuring      = "structuring"
	AlertTypeRapidMovement    = "rapid_movement"
	AlertTypeMismatch         = "mismatch"
)

// Severity constants.
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Alert status constants.
const (
	AlertStatusNew           = "new"
	AlertStatusInvestigating = "investigating"
	AlertStatusResolved      = "resolved"
	AlertStatusEscalated     = "escalated"
	AlertStatusReported      = "reported"
)

// Investigation outcome constants.
const (
	OutcomeFalsePositive = "false_positive"
	OutcomeSuspicious    = "suspicious"
	OutcomeConfirmed     = "confirmed"
	OutcomeEscalated     = "escalated"
)

// ListFilter menyimpan filter opsional untuk List query.
type ListFilter struct {
	Status     *string
	Severity   *string
	AlertType  *string
	RuleType   *string
	NasabahID  *uuid.UUID
}

// WriteRepository mendefinisikan operasi write untuk transaction alerts.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *TransactionAlert) error
	Update(ctx context.Context, s scope.Scope, e *TransactionAlert) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk transaction alerts.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*TransactionAlert, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*TransactionAlert, int, error)
}
