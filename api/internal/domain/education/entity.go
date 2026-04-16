// Package education mendefinisikan domain untuk manajemen kursus pendidikan anggota.
//
// Aturan layer:
//   - Package ini TIDAK boleh import pkg/tenant — domain bebas dari concern tenancy.
//   - pkg/scope boleh diimport karena Scope adalah pure value object tanpa framework dependency.
//   - Repository interface menerima scope.Scope sebagai parameter eksplisit.
package education

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/boilerplate/pkg/scope"
)

// CourseType mendefinisikan jenis kursus.
type CourseType string

const (
	CourseTypeMandatory     CourseType = "mandatory"
	CourseTypeElective      CourseType = "elective"
	CourseTypeCertification CourseType = "certification"
)

// MandatoryFor mendefinisikan untuk siapa kursus ini wajib.
type MandatoryFor string

const (
	MandatoryForNewMember MandatoryFor = "new_member"
	MandatoryForAllMember MandatoryFor = "all_member"
	MandatoryForOfficer   MandatoryFor = "officer"
)

// EducationCourse adalah entity utama untuk domain education.
// Sengaja tidak memiliki field TenantID/CompanyID — tenancy adalah infrastruktur concern.
type EducationCourse struct {
	ID            uuid.UUID  `db:"id"`
	Title         string     `db:"title"`
	Description   string     `db:"description"`
	CourseType    CourseType `db:"course_type"`
	Category      string     `db:"category"`
	DurationHours int        `db:"duration_hours"`
	ContentURL    string     `db:"content_url"`
	IsActive      bool       `db:"is_active"`
	PassingScore  float64    `db:"passing_score"`
	MaxAttempts   int        `db:"max_attempts"`
	MandatoryFor  []string   `db:"mandatory_for"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

// WriteRepository mendefinisikan operasi write untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type WriteRepository interface {
	Save(ctx context.Context, s scope.Scope, e *EducationCourse) error
	Update(ctx context.Context, s scope.Scope, e *EducationCourse) error
	Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error
}

// ReadRepository mendefinisikan operasi read untuk domain ini.
// scope.Scope diteruskan sebagai parameter eksplisit oleh application layer.
type ReadRepository interface {
	GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*EducationCourse, error)
	List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*EducationCourse, int, error)
}
