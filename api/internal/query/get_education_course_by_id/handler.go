// Package get_education_course_by_id menangani query untuk mengambil EducationCourse berdasarkan ID.
package get_education_course_by_id

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "get_education_course_by_id"

// Query berisi parameter untuk mengambil EducationCourse berdasarkan ID.
type Query struct {
	ID uuid.UUID
}

func (q Query) QueryName() string { return queryName }

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	CourseType    string   `json:"course_type"`
	Category      string   `json:"category"`
	DurationHours int      `json:"duration_hours"`
	ContentURL    string   `json:"content_url"`
	IsActive      bool     `json:"is_active"`
	PassingScore  float64  `json:"passing_score"`
	MaxAttempts   int      `json:"max_attempts"`
	MandatoryFor  []string `json:"mandatory_for"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// Handler menangani Query get_education_course_by_id.
type Handler struct {
	repo education.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo education.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil EducationCourse berdasarkan ID dengan filter scope organisasi.
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
		return nil, fmt.Errorf("get education course by id: %w", err)
	}

	return &Result{
		ID:            entity.ID.String(),
		Title:         entity.Title,
		Description:   entity.Description,
		CourseType:    string(entity.CourseType),
		Category:      entity.Category,
		DurationHours: entity.DurationHours,
		ContentURL:    entity.ContentURL,
		IsActive:      entity.IsActive,
		PassingScore:  entity.PassingScore,
		MaxAttempts:   entity.MaxAttempts,
		MandatoryFor:  entity.MandatoryFor,
		CreatedAt:     entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     entity.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
