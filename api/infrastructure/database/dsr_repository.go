package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/dsr"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// DSRRepository adalah concrete implementation dari dsr.WriteRepository + dsr.ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type DSRRepository struct {
	db *sqlx.DB
}

// NewDSRRepository membuat instance baru DSRRepository.
func NewDSRRepository(db *sqlx.DB) *DSRRepository {
	return &DSRRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity DataSubjectRequest baru ke database.
func (r *DSRRepository) Save(ctx context.Context, s scope.Scope, e *dsr.DataSubjectRequest) error {
	const q = `
		INSERT INTO data_subject_requests (
			id, tenant_id, company_id, request_type, status,
			requestor_name, requestor_email, subject_id, subject_type,
			description, verified_at, verified_by, completed_at, completed_by,
			rejection_reason, response_data, due_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.RequestType, e.Status,
		e.RequestorName, e.RequestorEmail, e.SubjectID, e.SubjectType,
		e.Description, e.VerifiedAt, e.VerifiedBy, e.CompletedAt, e.CompletedBy,
		e.RejectionReason, e.ResponseData, e.DueDate,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save dsr: %w", err)
	}
	return nil
}

// Update mengupdate DataSubjectRequest yang sudah ada berdasarkan ID + scope.
func (r *DSRRepository) Update(ctx context.Context, s scope.Scope, e *dsr.DataSubjectRequest) error {
	const q = `
		UPDATE data_subject_requests
		SET request_type = $4, status = $5,
		    requestor_name = $6, requestor_email = $7, subject_id = $8, subject_type = $9,
		    description = $10, verified_at = $11, verified_by = $12, completed_at = $13, completed_by = $14,
		    rejection_reason = $15, response_data = $16, due_date = $17, updated_at = $18
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.RequestType, e.Status,
		e.RequestorName, e.RequestorEmail, e.SubjectID, e.SubjectType,
		e.Description, e.VerifiedAt, e.VerifiedBy, e.CompletedAt, e.CompletedBy,
		e.RejectionReason, e.ResponseData, e.DueDate,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update dsr: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return dsr.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *DSRRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE data_subject_requests SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete dsr: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return dsr.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu DataSubjectRequest berdasarkan ID + scope.
func (r *DSRRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*dsr.DataSubjectRequest, error) {
	const q = `
		SELECT id, request_type, status, requestor_name, requestor_email,
		       subject_id, subject_type, description,
		       verified_at, verified_by, completed_at, completed_by,
		       rejection_reason, response_data, due_date,
		       created_at, updated_at, deleted_at
		FROM data_subject_requests
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e dsr.DataSubjectRequest
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.RequestType, &e.Status, &e.RequestorName, &e.RequestorEmail,
		&e.SubjectID, &e.SubjectType, &e.Description,
		&e.VerifiedAt, &e.VerifiedBy, &e.CompletedAt, &e.CompletedBy,
		&e.RejectionReason, &e.ResponseData, &e.DueDate,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, dsr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get dsr by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar DataSubjectRequest dengan filter, pagination, dan sorting.
func (r *DSRRepository) List(ctx context.Context, s scope.Scope, filter dsr.DSRFilter, limit, offset int, sortBy, order string) ([]*dsr.DataSubjectRequest, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countDSRs(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectDSRs(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *DSRRepository) buildWhereClause(s scope.Scope, f dsr.DSRFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.RequestType != nil {
		conditions = append(conditions, fmt.Sprintf("request_type = $%d", paramIdx))
		args = append(args, string(*f.RequestType))
		paramIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, string(*f.Status))
		paramIdx++
	}
	if f.SubjectID != nil {
		conditions = append(conditions, fmt.Sprintf("subject_id = $%d", paramIdx))
		args = append(args, *f.SubjectID)
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

// countDSRs menghitung total DataSubjectRequest yang cocok dengan filter.
func (r *DSRRepository) countDSRs(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM data_subject_requests WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count dsrs: %w", err)
	}
	return total, nil
}

// selectDSRs mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *DSRRepository) selectDSRs(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*dsr.DataSubjectRequest, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":   "created_at",
		"request_type": "request_type",
		"status":       "status",
		"due_date":     "due_date",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, request_type, status, requestor_name, requestor_email,
		        subject_id, subject_type, description,
		        verified_at, verified_by, completed_at, completed_by,
		        rejection_reason, response_data, due_date,
		        created_at, updated_at, deleted_at
		 FROM data_subject_requests
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list dsrs: %w", err)
	}
	defer rows.Close()
	return scanDSRs(rows)
}

func scanDSRs(rows *sql.Rows) ([]*dsr.DataSubjectRequest, error) {
	var results []*dsr.DataSubjectRequest
	for rows.Next() {
		var e dsr.DataSubjectRequest
		if err := rows.Scan(
			&e.ID, &e.RequestType, &e.Status, &e.RequestorName, &e.RequestorEmail,
			&e.SubjectID, &e.SubjectType, &e.Description,
			&e.VerifiedAt, &e.VerifiedBy, &e.CompletedAt, &e.CompletedBy,
			&e.RejectionReason, &e.ResponseData, &e.DueDate,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ dsr.WriteRepository = (*DSRRepository)(nil)
	_ dsr.ReadRepository  = (*DSRRepository)(nil)
)
