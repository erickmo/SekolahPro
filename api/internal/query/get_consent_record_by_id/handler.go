// Package get_consent_record_by_id menangani query untuk mengambil ConsentRecord berdasarkan ID.
package get_consent_record_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/consent"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_consent_record_by_id"

// Query berisi parameter untuk mengambil ConsentRecord berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                 string     `json:"id"`
	NasabahID          string     `json:"nasabah_id"`
	ConsentType        string     `json:"consent_type"`
	ConsentPurpose     string     `json:"consent_purpose"`
	LegalBasis         string     `json:"legal_basis"`
	ConsentText        string     `json:"consent_text"`
	ConsentVersion     string     `json:"consent_version"`
	ConsentGiven       bool       `json:"consent_given"`
	ConsentMethod      string     `json:"consent_method"`
	Withdrawn          bool       `json:"withdrawn"`
	WithdrawnAt        *string    `json:"withdrawn_at,omitempty"`
	WithdrawalReason   *string    `json:"withdrawal_reason,omitempty"`
	ParentID           *string    `json:"parent_id,omitempty"`
	ParentRelationship *string    `json:"parent_relationship,omitempty"`
	ParentConsentGiven *bool      `json:"parent_consent_given,omitempty"`
	GivenAt            string     `json:"given_at"`
	IPAddress          *string    `json:"ip_address,omitempty"`
	UserAgent          *string    `json:"user_agent,omitempty"`
	WitnessID          *string    `json:"witness_id,omitempty"`
	CreatedAt          string     `json:"created_at"`
}

// Handler menangani Query get_consent_record_by_id.
type Handler struct {
	repo consent.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo consent.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil ConsentRecord berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get consent record by id: %w", err)
	}

	result := &Result{
		ID:                 entity.ID.String(),
		NasabahID:          entity.NasabahID.String(),
		ConsentType:        entity.ConsentType,
		ConsentPurpose:     entity.ConsentPurpose,
		LegalBasis:         entity.LegalBasis,
		ConsentText:        entity.ConsentText,
		ConsentVersion:     entity.ConsentVersion,
		ConsentGiven:       entity.ConsentGiven,
		ConsentMethod:      entity.ConsentMethod,
		Withdrawn:          entity.Withdrawn,
		WithdrawalReason:   entity.WithdrawalReason,
		ParentRelationship: entity.ParentRelationship,
		ParentConsentGiven: entity.ParentConsentGiven,
		GivenAt:            entity.GivenAt.Format("2006-01-02T15:04:05Z"),
		IPAddress:          entity.IPAddress,
		UserAgent:          entity.UserAgent,
		CreatedAt:          entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if entity.WithdrawnAt != nil {
		wa := entity.WithdrawnAt.Format("2006-01-02T15:04:05Z")
		result.WithdrawnAt = &wa
	}
	if entity.ParentID != nil {
		pid := entity.ParentID.String()
		result.ParentID = &pid
	}
	if entity.WitnessID != nil {
		wid := entity.WitnessID.String()
		result.WitnessID = &wid
	}

	return result, nil
}
