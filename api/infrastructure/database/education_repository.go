package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yourorg/boilerplate/internal/domain/education"
	"github.com/yourorg/boilerplate/internal/domain/enrollment"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// EducationRepository adalah concrete implementation dari education.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type EducationRepository struct {
	db *sqlx.DB
}

// NewEducationRepository membuat instance baru EducationRepository.
func NewEducationRepository(db *sqlx.DB) *EducationRepository {
	return &EducationRepository{db: db}
}

// ── EducationCourse WriteRepository ────────────────────────────────────────────

// Save menyimpan entity EducationCourse baru ke database.
func (r *EducationRepository) Save(ctx context.Context, s scope.Scope, e *education.EducationCourse) error {
	const q = `
		INSERT INTO education_courses (id, tenant_id, company_id, title, description, course_type, category,
			duration_hours, content_url, is_active, passing_score, max_attempts, mandatory_for, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.Title, e.Description, e.CourseType, e.Category,
		e.DurationHours, e.ContentURL, e.IsActive,
		e.PassingScore, e.MaxAttempts,
		pq.Array(effectiveMandatoryFor(e.MandatoryFor)),
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save education course: %w", err)
	}
	return nil
}

// Update mengupdate EducationCourse yang sudah ada berdasarkan ID + scope.
func (r *EducationRepository) Update(ctx context.Context, s scope.Scope, e *education.EducationCourse) error {
	const q = `
		UPDATE education_courses
		SET title = $4, description = $5, course_type = $6, category = $7,
		    duration_hours = $8, content_url = $9, is_active = $10,
		    passing_score = $11, max_attempts = $12, mandatory_for = $13, updated_at = $14
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Title, e.Description, e.CourseType, e.Category,
		e.DurationHours, e.ContentURL, e.IsActive,
		e.PassingScore, e.MaxAttempts,
		pq.Array(effectiveMandatoryFor(e.MandatoryFor)),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update education course: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return education.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *EducationRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE education_courses SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete education course: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return education.ErrNotFound
	}
	return nil
}

// ── EducationCourse ReadRepository ─────────────────────────────────────────────

// GetByID mengambil satu EducationCourse berdasarkan ID + scope.
func (r *EducationRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*education.EducationCourse, error) {
	const q = `
		SELECT id, title, description, course_type, category, duration_hours, content_url,
		       is_active, passing_score, max_attempts, mandatory_for, created_at, updated_at, deleted_at
		FROM education_courses
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var (
		e          education.EducationCourse
		mandatoryFor pq.StringArray
	)
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.Title, &e.Description, &e.CourseType, &e.Category,
		&e.DurationHours, &e.ContentURL, &e.IsActive,
		&e.PassingScore, &e.MaxAttempts,
		&mandatoryFor,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	e.MandatoryFor = []string(mandatoryFor)
	if err == sql.ErrNoRows {
		return nil, education.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get education course by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar EducationCourse dengan pagination + sorting.
func (r *EducationRepository) List(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*education.EducationCourse, int, error) {
	total, err := r.countCourses(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectCourses(ctx, s, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *EducationRepository) countCourses(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM education_courses WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count education courses: %w", err)
	}
	return total, nil
}

func (r *EducationRepository) selectCourses(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*education.EducationCourse, error) {
	col := safeColumn(sortBy, map[string]string{
		"title": "title", "created_at": "created_at", "category": "category",
	}, "created_at")
	dir := safeOrder(order)
	q := fmt.Sprintf(
		`SELECT id, title, description, course_type, category, duration_hours, content_url,
		        is_active, passing_score, max_attempts, mandatory_for, created_at, updated_at, deleted_at
		 FROM education_courses
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)
	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list education courses: %w", err)
	}
	defer rows.Close()
	return scanCourses(rows)
}

func scanCourses(rows *sql.Rows) ([]*education.EducationCourse, error) {
	var results []*education.EducationCourse
	for rows.Next() {
		var (
			e            education.EducationCourse
			mandatoryFor pq.StringArray
		)
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.CourseType, &e.Category,
			&e.DurationHours, &e.ContentURL, &e.IsActive,
			&e.PassingScore, &e.MaxAttempts,
			&mandatoryFor,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		e.MandatoryFor = []string(mandatoryFor)
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ── Enrollment WriteRepository ─────────────────────────────────────────────────

// SaveEnrollment menyimpan entity EducationEnrollment baru ke database.
func (r *EducationRepository) SaveEnrollment(ctx context.Context, s scope.Scope, e *enrollment.EducationEnrollment) error {
	const q = `
		INSERT INTO education_enrollments (id, tenant_id, company_id, course_id, nasabah_id, status,
			enrolled_at, completed_at, score, attempts, certificate_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.CourseID, e.NasabahID, e.Status,
		e.EnrolledAt, e.CompletedAt, e.Score, e.Attempts,
		e.CertificateURL, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save enrollment: %w", err)
	}
	return nil
}

// UpdateEnrollment mengupdate EducationEnrollment yang sudah ada berdasarkan ID + scope.
func (r *EducationRepository) UpdateEnrollment(ctx context.Context, s scope.Scope, e *enrollment.EducationEnrollment) error {
	const q = `
		UPDATE education_enrollments
		SET status = $4, completed_at = $5, score = $6, attempts = $7,
		    certificate_url = $8, updated_at = $9
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.Status, e.CompletedAt, e.Score, e.Attempts,
		e.CertificateURL, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update enrollment: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return enrollment.ErrEnrollmentNotFound
	}
	return nil
}

// SaveKPI menyimpan entity EducationKPI baru ke database.
func (r *EducationRepository) SaveKPI(ctx context.Context, s scope.Scope, k *enrollment.EducationKPI) error {
	const q = `
		INSERT INTO education_kpis (id, tenant_id, company_id, metric_type, period_month, value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, q,
		k.ID, s.TenantID, s.CompanyID,
		k.MetricType, k.PeriodMonth, k.Value,
		k.CreatedAt, k.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save education kpi: %w", err)
	}
	return nil
}

// ── Enrollment ReadRepository ──────────────────────────────────────────────────

// GetEnrollmentByID mengambil satu EducationEnrollment berdasarkan ID + scope.
func (r *EducationRepository) GetEnrollmentByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*enrollment.EducationEnrollment, error) {
	const q = `
		SELECT id, course_id, nasabah_id, status, enrolled_at, completed_at,
		       score, attempts, certificate_url, created_at, updated_at, deleted_at
		FROM education_enrollments
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e enrollment.EducationEnrollment
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.CourseID, &e.NasabahID, &e.Status,
		&e.EnrolledAt, &e.CompletedAt,
		&e.Score, &e.Attempts, &e.CertificateURL,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, enrollment.ErrEnrollmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get enrollment by id: %w", err)
	}
	return &e, nil
}

// ListEnrollments mengambil daftar EducationEnrollment dengan pagination + sorting.
func (r *EducationRepository) ListEnrollments(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*enrollment.EducationEnrollment, int, error) {
	total, err := r.countEnrollments(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectEnrollments(ctx, s, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *EducationRepository) countEnrollments(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM education_enrollments WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count enrollments: %w", err)
	}
	return total, nil
}

func (r *EducationRepository) selectEnrollments(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*enrollment.EducationEnrollment, error) {
	col := safeColumn(sortBy, map[string]string{
		"enrolled_at": "enrolled_at", "created_at": "created_at", "status": "status",
	}, "created_at")
	dir := safeOrder(order)
	q := fmt.Sprintf(
		`SELECT id, course_id, nasabah_id, status, enrolled_at, completed_at,
		        score, attempts, certificate_url, created_at, updated_at, deleted_at
		 FROM education_enrollments
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)
	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	defer rows.Close()
	return scanEnrollments(rows)
}

func scanEnrollments(rows *sql.Rows) ([]*enrollment.EducationEnrollment, error) {
	var results []*enrollment.EducationEnrollment
	for rows.Next() {
		var e enrollment.EducationEnrollment
		if err := rows.Scan(
			&e.ID, &e.CourseID, &e.NasabahID, &e.Status,
			&e.EnrolledAt, &e.CompletedAt,
			&e.Score, &e.Attempts, &e.CertificateURL,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ListKPIs mengambil daftar EducationKPI dengan pagination + sorting.
func (r *EducationRepository) ListKPIs(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*enrollment.EducationKPI, int, error) {
	total, err := r.countKPIs(ctx, s)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectKPIs(ctx, s, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *EducationRepository) countKPIs(ctx context.Context, s scope.Scope) (int, error) {
	const q = `SELECT COUNT(*) FROM education_kpis WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count education kpis: %w", err)
	}
	return total, nil
}

func (r *EducationRepository) selectKPIs(ctx context.Context, s scope.Scope, limit, offset int, sortBy, order string) ([]*enrollment.EducationKPI, error) {
	col := safeColumn(sortBy, map[string]string{
		"period_month": "period_month", "created_at": "created_at", "metric_type": "metric_type",
	}, "created_at")
	dir := safeOrder(order)
	q := fmt.Sprintf(
		`SELECT id, metric_type, period_month, value, created_at, updated_at, deleted_at
		 FROM education_kpis
		 WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
		 ORDER BY %s %s LIMIT $3 OFFSET $4`, col, dir,
	)
	rows, err := r.db.QueryContext(ctx, q, s.TenantID, s.CompanyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list education kpis: %w", err)
	}
	defer rows.Close()
	return scanKPIs(rows)
}

func scanKPIs(rows *sql.Rows) ([]*enrollment.EducationKPI, error) {
	var results []*enrollment.EducationKPI
	for rows.Next() {
		var k enrollment.EducationKPI
		if err := rows.Scan(
			&k.ID, &k.MetricType, &k.PeriodMonth, &k.Value,
			&k.CreatedAt, &k.UpdatedAt, &k.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &k)
	}
	return results, rows.Err()
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// effectiveMandatoryFor mengembalikan slice yang aman untuk disimpan ke database.
func effectiveMandatoryFor(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
