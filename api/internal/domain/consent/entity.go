// Package consent adalah domain CQRS untuk data privacy consent management (ADR-K028).
//
// Mengelola consent records, data subject requests, dan data breach notifications
// untuk compliance UU PDP No. 27/2022.
//
// Aturan layer:
//   - Package ini TIDAK boleh import infrastructure packages.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package consent

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// ConsentRecord adalah entity untuk persetujuan pemrosesan data pribadi.
type ConsentRecord struct {
	ID                  uuid.UUID  `db:"id"                    json:"id"`
	NasabahID           uuid.UUID  `db:"nasabah_id"            json:"nasabah_id"`
	ConsentType         string     `db:"consent_type"          json:"consent_type"`
	ConsentPurpose      string     `db:"consent_purpose"       json:"consent_purpose"`
	LegalBasis          string     `db:"legal_basis"           json:"legal_basis"`
	ConsentText         string     `db:"consent_text"          json:"consent_text"`
	ConsentVersion      string     `db:"consent_version"       json:"consent_version"`
	ConsentGiven        bool       `db:"consent_given"         json:"consent_given"`
	ConsentMethod       string     `db:"consent_method"        json:"consent_method"`
	Withdrawn           bool       `db:"withdrawn"             json:"withdrawn"`
	WithdrawnAt         *time.Time `db:"withdrawn_at"          json:"withdrawn_at,omitempty"`
	WithdrawalReason    *string    `db:"withdrawal_reason"     json:"withdrawal_reason,omitempty"`
	ParentID            *uuid.UUID `db:"parent_id"             json:"parent_id,omitempty"`
	ParentRelationship  *string    `db:"parent_relationship"   json:"parent_relationship,omitempty"`
	ParentConsentGiven  *bool      `db:"parent_consent_given"  json:"parent_consent_given,omitempty"`
	GivenAt             time.Time  `db:"given_at"              json:"given_at"`
	IPAddress           *string    `db:"ip_address"            json:"ip_address,omitempty"`
	UserAgent           *string    `db:"user_agent"            json:"user_agent,omitempty"`
	WitnessID           *uuid.UUID `db:"witness_id"            json:"witness_id,omitempty"`
	CreatedAt           time.Time  `db:"created_at"            json:"created_at"`
	DeletedAt           *time.Time `db:"deleted_at"            json:"-"`
}

// Consent type constants.
const (
	ConsentTypeDataCollection    = "data_collection"
	ConsentTypeDataProcessing    = "data_processing"
	ConsentTypeDataSharing       = "data_sharing"
	ConsentTypeMarketing         = "marketing"
	ConsentTypeBiometric         = "biometric"
	ConsentTypeMinorParental     = "minor_parental"
	ConsentTypeCrossBorderTransfer = "cross_border_transfer"
	ConsentTypeProfiling         = "profiling"
)

// Legal basis constants.
const (
	LegalBasisConsent          = "consent"
	LegalBasisContractual      = "contractual"
	LegalBasisLegalObligation  = "legal_obligation"
	LegalBasisVitalInterest    = "vital_interest"
	LegalBasisPublicInterest   = "public_interest"
	LegalBasisLegitimateInterest = "legitimate_interest"
)

// Consent method constants.
const (
	ConsentMethodDigitalSignature = "digital_signature"
	ConsentMethodCheckbox         = "checkbox"
	ConsentMethodVerbalRecorded   = "verbal_recorded"
	ConsentMethodWritten          = "written"
)

// ListFilter menyimpan filter opsional untuk List query.
type ListFilter struct {
	ConsentType  *string
	LegalBasis   *string
	ConsentGiven *bool
	Withdrawn    *bool
	NasabahID    *uuid.UUID
}

// WriteRepository mendefinisikan operasi write untuk consent records.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *ConsentRecord) error
	Update(ctx context.Context, s scope.Scope, e *ConsentRecord) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk consent records.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*ConsentRecord, error)
	List(ctx context.Context, s scope.Scope, filter ListFilter, limit, offset int, sortBy, order string) ([]*ConsentRecord, int, error)
}
