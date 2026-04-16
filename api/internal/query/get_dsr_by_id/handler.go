// Package get_dsr_by_id menangani query untuk mengambil DataSubjectRequest berdasarkan ID.
package get_dsr_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/dsr"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_dsr_by_id"

// Query berisi parameter untuk mengambil DataSubjectRequest berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID              string `json:"id"`
	RequestType     string `json:"request_type"`
	Status          string `json:"status"`
	RequestorName   string `json:"requestor_name"`
	RequestorEmail  string `json:"requestor_email"`
	SubjectID       string `json:"subject_id"`
	SubjectType     string `json:"subject_type"`
	Description     string `json:"description"`
	VerifiedAt      string `json:"verified_at,omitempty"`
	VerifiedBy      string `json:"verified_by,omitempty"`
	CompletedAt     string `json:"completed_at,omitempty"`
	CompletedBy     string `json:"completed_by,omitempty"`
	RejectionReason string `json:"rejection_reason,omitempty"`
	ResponseData    string `json:"response_data,omitempty"`
	DueDate         string `json:"due_date"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// Handler menangani Query get_dsr_by_id.
type Handler struct {
	repo dsr.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo dsr.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil DataSubjectRequest berdasarkan ID.
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
		return nil, fmt.Errorf("get dsr by id: %w", err)
	}

	result := &Result{
		ID:              entity.ID.String(),
		RequestType:     string(entity.RequestType),
		Status:          string(entity.Status),
		RequestorName:   entity.RequestorName,
		RequestorEmail:  entity.RequestorEmail,
		SubjectID:       entity.SubjectID.String(),
		SubjectType:     entity.SubjectType,
		Description:     entity.Description,
		RejectionReason: entity.RejectionReason,
		ResponseData:    entity.ResponseData,
		DueDate:         entity.DueDate.Format("2006-01-02T15:04:05Z"),
		CreatedAt:       entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if entity.VerifiedAt != nil {
		result.VerifiedAt = entity.VerifiedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.VerifiedBy != nil {
		result.VerifiedBy = entity.VerifiedBy.String()
	}
	if entity.CompletedAt != nil {
		result.CompletedAt = entity.CompletedAt.Format("2006-01-02T15:04:05Z")
	}
	if entity.CompletedBy != nil {
		result.CompletedBy = entity.CompletedBy.String()
	}

	return result, nil
}
