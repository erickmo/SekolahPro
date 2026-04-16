package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/biometric"
	"github.com/yourorg/boilerplate/internal/domain/biometric_log"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// BiometricRepository adalah concrete implementation dari biometric.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type BiometricRepository struct {
	db *sqlx.DB
}

// NewBiometricRepository membuat instance baru BiometricRepository.
func NewBiometricRepository(db *sqlx.DB) *BiometricRepository {
	return &BiometricRepository{db: db}
}

// ── Biometric WriteRepository ─────────────────────────────────────────────────

// Save menyimpan entity BiometricEnrollment baru ke database.
func (r *BiometricRepository) Save(ctx context.Context, s scope.Scope, e *biometric.BiometricEnrollment) error {
	const q = `
		INSERT INTO biometric_enrollments (id, tenant_id, company_id, nasabah_id, biometric_type, device_info,
			template_hash, is_active, verified_at, verified_by, failed_attempts, last_attempt_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.NasabahID, e.BiometricType, e.DeviceInfo,
		e.TemplateHash, e.IsActive, e.VerifiedAt, e.VerifiedBy,
		e.FailedAttempts, e.LastAttemptAt,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save biometric enrollment: %w", err)
	}
	return nil
}

// Update mengupdate BiometricEnrollment yang sudah ada berdasarkan ID + scope.
func (r *BiometricRepository) Update(ctx context.Context, s scope.Scope, e *biometric.BiometricEnrollment) error {
	const q = `
		UPDATE biometric_enrollments
		SET nasabah_id = $4, biometric_type = $5, device_info = $6,
		    template_hash = $7, is_active = $8, verified_at = $9, verified_by = $10,
		    failed_attempts = $11, last_attempt_at = $12, updated_at = $13
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.NasabahID, e.BiometricType, e.DeviceInfo,
		e.TemplateHash, e.IsActive, e.VerifiedAt, e.VerifiedBy,
		e.FailedAttempts, e.LastAttemptAt,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update biometric enrollment: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return biometric.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *BiometricRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE biometric_enrollments SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete biometric enrollment: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return biometric.ErrNotFound
	}
	return nil
}

// ── Biometric ReadRepository ──────────────────────────────────────────────────

// GetByID mengambil satu BiometricEnrollment berdasarkan ID + scope.
func (r *BiometricRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*biometric.BiometricEnrollment, error) {
	const q = `
		SELECT id, nasabah_id, biometric_type, device_info, template_hash,
		       is_active, verified_at, verified_by, failed_attempts, last_attempt_at,
		       created_at, updated_at, deleted_at
		FROM biometric_enrollments
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e biometric.BiometricEnrollment
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.NasabahID, &e.BiometricType, &e.DeviceInfo, &e.TemplateHash,
		&e.IsActive, &e.VerifiedAt, &e.VerifiedBy, &e.FailedAttempts, &e.LastAttemptAt,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, biometric.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get biometric enrollment by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar BiometricEnrollment dengan dynamic filter, pagination, dan sorting.
func (r *BiometricRepository) List(ctx context.Context, s scope.Scope, filter biometric.BiometricFilter, limit, offset int, sortBy, order string) ([]*biometric.BiometricEnrollment, int, error) {
	total, err := r.countBiometrics(ctx, s, filter)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectBiometrics(ctx, s, filter, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *BiometricRepository) countBiometrics(ctx context.Context, s scope.Scope, filter biometric.BiometricFilter) (int, error) {
	where, args := r.buildWhereClause(s, filter)
	q := fmt.Sprintf(`SELECT COUNT(*) FROM biometric_enrollments WHERE %s`, where)

	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count biometric enrollments: %w", err)
	}
	return total, nil
}

func (r *BiometricRepository) selectBiometrics(ctx context.Context, s scope.Scope, filter biometric.BiometricFilter, limit, offset int, sortBy, order string) ([]*biometric.BiometricEnrollment, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":     "created_at",
		"nasabah_id":     "nasabah_id",
		"biometric_type": "biometric_type",
	}, "created_at")
	dir := safeOrder(order)

	where, args := r.buildWhereClause(s, filter)
	q := fmt.Sprintf(
		`SELECT id, nasabah_id, biometric_type, device_info, template_hash,
		        is_active, verified_at, verified_by, failed_attempts, last_attempt_at,
		        created_at, updated_at, deleted_at
		 FROM biometric_enrollments
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list biometric enrollments: %w", err)
	}
	defer rows.Close()
	return scanBiometricEnrollments(rows)
}

func (r *BiometricRepository) buildWhereClause(s scope.Scope, filter biometric.BiometricFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if filter.NasabahID != nil {
		conditions = append(conditions, fmt.Sprintf("nasabah_id = $%d", paramIdx))
		args = append(args, *filter.NasabahID)
		paramIdx++
	}
	if filter.BiometricType != nil {
		conditions = append(conditions, fmt.Sprintf("biometric_type = $%d", paramIdx))
		args = append(args, *filter.BiometricType)
		paramIdx++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", paramIdx))
		args = append(args, *filter.IsActive)
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func scanBiometricEnrollments(rows *sql.Rows) ([]*biometric.BiometricEnrollment, error) {
	var results []*biometric.BiometricEnrollment
	for rows.Next() {
		var e biometric.BiometricEnrollment
		if err := rows.Scan(
			&e.ID, &e.NasabahID, &e.BiometricType, &e.DeviceInfo, &e.TemplateHash,
			&e.IsActive, &e.VerifiedAt, &e.VerifiedBy, &e.FailedAttempts, &e.LastAttemptAt,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// ── BiometricLog WriteRepository ──────────────────────────────────────────────

// SaveLog menyimpan entity BiometricVerificationLog baru ke database.
func (r *BiometricRepository) SaveLog(ctx context.Context, s scope.Scope, e *biometric_log.BiometricVerificationLog) error {
	const q = `
		INSERT INTO biometric_verification_logs (id, tenant_id, company_id, enrollment_id, nasabah_id,
			verification_result, fallback_method, device_info, ip_address, attempted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.EnrollmentID, e.NasabahID,
		e.VerificationResult, e.FallbackMethod, e.DeviceInfo, e.IPAddress,
		e.AttemptedAt, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save verification log: %w", err)
	}
	return nil
}

// ── BiometricLog ReadRepository ───────────────────────────────────────────────

// GetLogByID mengambil satu BiometricVerificationLog berdasarkan ID + scope.
func (r *BiometricRepository) GetLogByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*biometric_log.BiometricVerificationLog, error) {
	const q = `
		SELECT id, enrollment_id, nasabah_id, verification_result, fallback_method,
		       device_info, ip_address, attempted_at, created_at, deleted_at
		FROM biometric_verification_logs
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e biometric_log.BiometricVerificationLog
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.EnrollmentID, &e.NasabahID, &e.VerificationResult, &e.FallbackMethod,
		&e.DeviceInfo, &e.IPAddress, &e.AttemptedAt, &e.CreatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, biometric_log.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get biometric log by id: %w", err)
	}
	return &e, nil
}

// ListLogs mengambil daftar BiometricVerificationLog dengan dynamic filter, pagination, dan sorting.
func (r *BiometricRepository) ListLogs(ctx context.Context, s scope.Scope, filter biometric_log.LogFilter, limit, offset int, sortBy, order string) ([]*biometric_log.BiometricVerificationLog, int, error) {
	total, err := r.countLogs(ctx, s, filter)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectLogs(ctx, s, filter, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *BiometricRepository) countLogs(ctx context.Context, s scope.Scope, filter biometric_log.LogFilter) (int, error) {
	where, args := r.buildLogWhereClause(s, filter)
	q := fmt.Sprintf(`SELECT COUNT(*) FROM biometric_verification_logs WHERE %s`, where)

	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count verification logs: %w", err)
	}
	return total, nil
}

func (r *BiometricRepository) selectLogs(ctx context.Context, s scope.Scope, filter biometric_log.LogFilter, limit, offset int, sortBy, order string) ([]*biometric_log.BiometricVerificationLog, error) {
	col := safeColumn(sortBy, map[string]string{
		"attempted_at":         "attempted_at",
		"created_at":           "created_at",
		"verification_result":  "verification_result",
	}, "created_at")
	dir := safeOrder(order)

	where, args := r.buildLogWhereClause(s, filter)
	q := fmt.Sprintf(
		`SELECT id, enrollment_id, nasabah_id, verification_result, fallback_method,
		        device_info, ip_address, attempted_at, created_at, deleted_at
		 FROM biometric_verification_logs
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list verification logs: %w", err)
	}
	defer rows.Close()
	return scanVerificationLogs(rows)
}

func (r *BiometricRepository) buildLogWhereClause(s scope.Scope, filter biometric_log.LogFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if filter.EnrollmentID != nil {
		conditions = append(conditions, fmt.Sprintf("enrollment_id = $%d", paramIdx))
		args = append(args, *filter.EnrollmentID)
		paramIdx++
	}
	if filter.NasabahID != nil {
		conditions = append(conditions, fmt.Sprintf("nasabah_id = $%d", paramIdx))
		args = append(args, *filter.NasabahID)
		paramIdx++
	}
	if filter.VerificationResult != nil {
		conditions = append(conditions, fmt.Sprintf("verification_result = $%d", paramIdx))
		args = append(args, *filter.VerificationResult)
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

func scanVerificationLogs(rows *sql.Rows) ([]*biometric_log.BiometricVerificationLog, error) {
	var results []*biometric_log.BiometricVerificationLog
	for rows.Next() {
		var e biometric_log.BiometricVerificationLog
		if err := rows.Scan(
			&e.ID, &e.EnrollmentID, &e.NasabahID, &e.VerificationResult, &e.FallbackMethod,
			&e.DeviceInfo, &e.IPAddress, &e.AttemptedAt, &e.CreatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}
