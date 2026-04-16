// Package enrollment mendefinisikan domain untuk pendaftaran kursus dan KPI pendidikan anggota.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package enrollment

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// EnrollmentStatus mendefinisikan status pendaftaran kursus.
type EnrollmentStatus string

const (
	StatusEnrolled    EnrollmentStatus = "enrolled"
	StatusInProgress  EnrollmentStatus = "in_progress"
	StatusCompleted   EnrollmentStatus = "completed"
	StatusFailed      EnrollmentStatus = "failed"
	StatusExpired     EnrollmentStatus = "expired"
)

// EducationEnrollment adalah entity untuk pendaftaran kursus anggota.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type EducationEnrollment struct {
	ID             uuid.UUID        `db:"id"`
	CourseID       uuid.UUID        `db:"course_id"`
	NasabahID      uuid.UUID        `db:"nasabah_id"`
	Status         EnrollmentStatus `db:"status"`
	EnrolledAt     time.Time        `db:"enrolled_at"`
	CompletedAt    *time.Time       `db:"completed_at"`
	Score          float64          `db:"score"`
	Attempts       int              `db:"attempts"`
	CertificateURL string           `db:"certificate_url"`
	CreatedAt      time.Time        `db:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at"`
	DeletedAt      *time.Time       `db:"deleted_at"`
}

// MetricType mendefinisikan jenis metrik KPI pendidikan.
type MetricType string

const (
	MetricCompletionRate  MetricType = "completion_rate"
	MetricAvgScore        MetricType = "avg_score"
	MetricEnrollmentCount MetricType = "enrollment_count"
	MetricPassRate        MetricType = "pass_rate"
)

// EducationKPI adalah entity untuk menyimpan metrik KPI pendidikan.
type EducationKPI struct {
	ID          uuid.UUID `db:"id"`
	MetricType  MetricType `db:"metric_type"`
	PeriodMonth string    `db:"period_month"`
	Value       float64   `db:"value"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	SaveEnrollment(ctx context.Context, s scope.Scope, e *EducationEnrollment) error
	UpdateEnrollment(ctx context.Context, s scope.Scope, e *EducationEnrollment) error
	SaveKPI(ctx context.Context, s scope.Scope, k *EducationKPI) error
}

// ReadRepository mendefinisikan operasi read untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetEnrollmentByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*EducationEnrollment, error)
	ListEnrollments(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*EducationEnrollment, int, error)
	ListKPIs(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*EducationKPI, int, error)
}
