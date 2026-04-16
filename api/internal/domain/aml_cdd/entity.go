// Package aml_cdd adalah domain CQRS untuk AML Customer Due Diligence (ADR-K027).
//
// Mengelola CDD records, risk assessment, PEP screening, dan beneficial ownership
// untuk compliance Anti-Money Laundering.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package aml_cdd

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// RiskCategory menyimpan detail risk per kategori.
type RiskCategory struct {
	CustomerRisk   string `json:"customer_risk"`
	ProductRisk    string `json:"product_risk"`
	GeographicRisk string `json:"geographic_risk"`
	ChannelRisk    string `json:"channel_risk"`
}

// CustomerDueDiligence adalah entity utama untuk CDD/EDD records.
type CustomerDueDiligence struct {
	ID                      uuid.UUID      `db:"id"                        json:"id"`
	NasabahID               uuid.UUID      `db:"nasabah_id"                json:"nasabah_id"`
	RiskLevel               string         `db:"risk_level"                json:"risk_level"`
	RiskScore               int            `db:"risk_score"                json:"risk_score"`
	RiskCategory            json.RawMessage `db:"risk_category"            json:"risk_category"`
	CDDLevel                string         `db:"cdd_level"                 json:"cdd_level"`
	CDDPurpose              string         `db:"cdd_purpose"               json:"cdd_purpose,omitempty"`
	SourceOfFunds           *string        `db:"source_of_funds"           json:"source_of_funds,omitempty"`
	SourceOfWealth          *string        `db:"source_of_wealth"          json:"source_of_wealth,omitempty"`
	IsPEP                   bool           `db:"is_pep"                    json:"is_pep"`
	PEPType                 string         `db:"pep_type"                  json:"pep_type"`
	PEPPosition             *string        `db:"pep_position"              json:"pep_position,omitempty"`
	PEPCountry              *string        `db:"pep_country"               json:"pep_country,omitempty"`
	PEPRelationship          string        `db:"pep_relationship"          json:"pep_relationship"`
	PEPScreeningDate        *time.Time     `db:"pep_screening_date"        json:"pep_screening_date,omitempty"`
	PEPScreeningSource      *string        `db:"pep_screening_source"      json:"pep_screening_source,omitempty"`
	BeneficialOwnerName     *string        `db:"beneficial_owner_name"     json:"beneficial_owner_name,omitempty"`
	BeneficialOwnerIDNo     *string        `db:"beneficial_owner_id_no"    json:"beneficial_owner_id_no,omitempty"`
	BeneficialOwnershipPct  *float64       `db:"beneficial_ownership_pct"  json:"beneficial_ownership_pct,omitempty"`
	BOVerified              bool           `db:"bo_verified"               json:"bo_verified"`
	NextReviewDate          *time.Time     `db:"next_review_date"          json:"next_review_date,omitempty"`
	LastReviewedAt          *time.Time     `db:"last_reviewed_at"          json:"last_reviewed_at,omitempty"`
	LastReviewedBy          *uuid.UUID     `db:"last_reviewed_by"          json:"last_reviewed_by,omitempty"`
	ReviewCount             int            `db:"review_count"              json:"review_count"`
	Status                  string         `db:"status"                    json:"status"`
	RestrictionReason       *string        `db:"restriction_reason"        json:"restriction_reason,omitempty"`
	ExitReason              *string        `db:"exit_reason"               json:"exit_reason,omitempty"`
	CreatedAt               time.Time      `db:"created_at"                json:"created_at"`
	UpdatedAt               time.Time      `db:"updated_at"                json:"updated_at"`
	CreatedBy               uuid.UUID      `db:"created_by"                json:"created_by"`
	UpdatedBy               uuid.UUID      `db:"updated_by"                json:"updated_by"`
	DeletedAt               *time.Time     `db:"deleted_at"                json:"-"`
}

// Risk level constants.
const (
	RiskLevelLow        = "low"
	RiskLevelMedium     = "medium"
	RiskLevelHigh       = "high"
	RiskLevelProhibited = "prohibited"
)

// CDD level constants.
const (
	CDDLevelSDD = "sdd"
	CDDLevelCDD = "cdd"
	CDDLevelEDD = "edd"
)

// CDD status constants.
const (
	CDDStatusPending    = "pending"
	CDDStatusActive     = "active"
	CDDStatusEscalated  = "escalated"
	CDDStatusRestricted = "restricted"
	CDDStatusExited     = "exited"
)

// ListFilter menyimpan filter opsional untuk List query.
type ListFilter struct {
	RiskLevel *string
	CDDLevel  *string
	Status    *string
	IsPEP     *bool
}

// WriteRepository mendefinisikan operasi write untuk CDD records.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *CustomerDueDiligence) error
	Update(ctx context.Context, s scope.Scope, e *CustomerDueDiligence) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk CDD records.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*CustomerDueDiligence, error)
	GetByNasabahID(ctx context.Context, s scope.Scope, nasabahID uuid.UUID) (*CustomerDueDiligence, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*CustomerDueDiligence, int, error)
}
