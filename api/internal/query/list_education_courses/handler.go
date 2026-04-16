// Package list_education_courses menangani query untuk mengambil daftar EducationCourse.
package list_education_courses

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_education_courses"

// Query berisi parameter untuk mengambil daftar EducationCourse.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu EducationCourse dalam hasil list.
type Item struct {
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
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_education_courses.
type Handler struct {
	repo education.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo education.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar EducationCourse dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.List(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list education courses: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		items = append(items, &Item{
			ID:            e.ID.String(),
			Title:         e.Title,
			Description:   e.Description,
			CourseType:    string(e.CourseType),
			Category:      e.Category,
			DurationHours: e.DurationHours,
			ContentURL:    e.ContentURL,
			IsActive:      e.IsActive,
			PassingScore:  e.PassingScore,
			MaxAttempts:   e.MaxAttempts,
			MandatoryFor:  e.MandatoryFor,
			CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
