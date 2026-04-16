// Package list_enrollments menangani query untuk mengambil daftar EducationEnrollment.
package list_enrollments

import (
	"context"
	"fmt"

	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/pagination"
	"github.com/yourorg/boilerplate/pkg/scope"
)

const queryName = "list_enrollments"

// Query berisi parameter untuk mengambil daftar EducationEnrollment.
type Query struct {
	Params pagination.ListParams
}

func (q Query) QueryName() string { return queryName }

// Item adalah representasi satu EducationEnrollment dalam hasil list.
type Item struct {
	ID             string  `json:"id"`
	CourseID       string  `json:"course_id"`
	NasabahID      string  `json:"nasabah_id"`
	Status         string  `json:"status"`
	EnrolledAt     string  `json:"enrolled_at"`
	CompletedAt    string  `json:"completed_at"`
	Score          float64 `json:"score"`
	Attempts       int     `json:"attempts"`
	CertificateURL string  `json:"certificate_url"`
	CreatedAt      string  `json:"created_at"`
}

// Result adalah data yang dikembalikan oleh query ini.
type Result struct {
	Data  []*Item `json:"data"`
	Total int     `json:"total"`
}

// Handler menangani Query list_enrollments.
type Handler struct {
	repo enrollment.ReadRepository
}

// NewHandler membuat instance Handler baru.
func NewHandler(repo enrollment.ReadRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle mengambil daftar EducationEnrollment dengan pagination, sorting, dan filter scope organisasi.
func (h *Handler) Handle(ctx context.Context, qry Query) (*Result, error) {
	s, ok := scope.ScopeFromContext(ctx)
	if !ok {
		return nil, scope.ErrMissingTenant
	}
	if err := s.IsValid(); err != nil {
		return nil, err
	}

	p := qry.Params
	entities, total, err := h.repo.ListEnrollments(ctx, s, p.Limit, p.Offset, p.SortBy, p.Order)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}

	items := make([]*Item, 0, len(entities))
	for _, e := range entities {
		completedAt := ""
		if e.CompletedAt != nil {
			completedAt = e.CompletedAt.Format("2006-01-02T15:04:05Z")
		}
		items = append(items, &Item{
			ID:             e.ID.String(),
			CourseID:       e.CourseID.String(),
			NasabahID:      e.NasabahID.String(),
			Status:         string(e.Status),
			EnrolledAt:     e.EnrolledAt.Format("2006-01-02T15:04:05Z"),
			CompletedAt:    completedAt,
			Score:          e.Score,
			Attempts:       e.Attempts,
			CertificateURL: e.CertificateURL,
			CreatedAt:      e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &Result{Data: items, Total: total}, nil
}
