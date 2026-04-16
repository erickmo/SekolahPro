// Package get_enrollment_by_id menangani query untuk mengambil EducationEnrollment berdasarkan ID.
package get_enrollment_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_enrollment_by_id"

// Query berisi parameter untuk mengambil EducationEnrollment berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID             string   `json:"id"`
	CourseID       string   `json:"course_id"`
	NasabahID      string   `json:"nasabah_id"`
	Status         string   `json:"status"`
	EnrolledAt     string   `json:"enrolled_at"`
	CompletedAt    string   `json:"completed_at"`
	Score          float64  `json:"score"`
	Attempts       int      `json:"attempts"`
	CertificateURL string   `json:"certificate_url"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

// Handler menangani Query get_enrollment_by_id.
type Handler struct {
	repo enrollment.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo enrollment.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil EducationEnrollment berdasarkan ID dengan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	entity, err := h.repo.GetEnrollmentByID(ctx, s, qry.ID)
	if err != nil {
		return nil, fmt.Errorf("get enrollment by id: %w", err)
	}

	completedAt := ""
	if entity.CompletedAt != nil {
		completedAt = entity.CompletedAt.Format("2006-01-02T15:04:05Z")
	}

	return &Result{
		ID:             entity.ID.String(),
		CourseID:       entity.CourseID.String(),
		NasabahID:      entity.NasabahID.String(),
		Status:         string(entity.Status),
		EnrolledAt:     entity.EnrolledAt.Format("2006-01-02T15:04:05Z"),
		CompletedAt:    completedAt,
		Score:          entity.Score,
		Attempts:       entity.Attempts,
		CertificateURL: entity.CertificateURL,
		CreatedAt:      entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
