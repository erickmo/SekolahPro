// Package get_aml_alert_by_id menangani query untuk mengambil TransactionAlert berdasarkan ID.
package get_aml_alert_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/aml_alert"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_aml_alert_by_id"

// Query berisi parameter untuk mengambil TransactionAlert berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID                   string     `json:"id"`
	NasabahID            string     `json:"nasabah_id"`
	TransaksiID          *string    `json:"transaksi_id,omitempty"`
	RuleID               string     `json:"rule_id"`
	RuleType             string     `json:"rule_type"`
	AlertType            string     `json:"alert_type"`
	Severity             string     `json:"severity"`
	Description          string     `json:"description"`
	InvestigatedBy       *string    `json:"investigated_by,omitempty"`
	InvestigationNotes   *string    `json:"investigation_notes,omitempty"`
	InvestigationOutcome *string    `json:"investigation_outcome,omitempty"`
	LTKMFiled           bool       `json:"ltkm_filed"`
	LTKMReference        *string    `json:"ltkm_reference,omitempty"`
	Status               string     `json:"status"`
	ResolvedAt           *string    `json:"resolved_at,omitempty"`
	DetectedAt           string     `json:"detected_at"`
	CreatedAt            string     `json:"created_at"`
}

// Handler menangani Query get_aml_alert_by_id.
type Handler struct {
	repo aml_alert.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo aml_alert.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil TransactionAlert berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get aml alert by id: %w", err)
	}

	result := &Result{
		ID:                 entity.ID.String(),
		NasabahID:          entity.NasabahID.String(),
		RuleID:             entity.RuleID.String(),
		RuleType:           entity.RuleType,
		AlertType:          entity.AlertType,
		Severity:           entity.Severity,
		Description:        entity.Description,
		InvestigatedBy:     stringPtrFromUUID(entity.InvestigatedBy),
		InvestigationNotes: entity.InvestigationNotes,
		InvestigationOutcome: entity.InvestigationOutcome,
		LTKMFiled:         entity.LTKMFiled,
		LTKMReference:      entity.LTKMReference,
		Status:             entity.Status,
		DetectedAt:         entity.DetectedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:          entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if entity.TransaksiID != nil {
		tid := entity.TransaksiID.String()
		result.TransaksiID = &tid
	}
	if entity.ResolvedAt != nil {
		ra := entity.ResolvedAt.Format("2006-01-02T15:04:05Z")
		result.ResolvedAt = &ra
	}

	return result, nil
}

func stringPtrFromUUID(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}
