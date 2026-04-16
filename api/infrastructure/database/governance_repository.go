package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/governance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// GovernanceRepository adalah concrete implementation dari governance.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type GovernanceRepository struct {
	db *sqlx.DB
}

// NewGovernanceRepository membuat instance baru GovernanceRepository.
func NewGovernanceRepository(db *sqlx.DB) *GovernanceRepository {
	return &GovernanceRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity GovernancePosition baru ke database.
func (r *GovernanceRepository) Save(ctx context.Context, s scope.Scope, e *governance.GovernancePosition) error {
	const q = `
		INSERT INTO governance_positions (
			id, tenant_id, company_id, nasabah_id,
			position_type, position_level, term_start, term_end, term_number,
			status, appointed_by, appointment_doc_id,
			max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID, e.NasabahID,
		e.PositionType, e.PositionLevel, e.TermStart, e.TermEnd, e.TermNumber,
		e.Status, e.AppointedBy, e.AppointmentDocID,
		e.MaxApprovalAmount, e.CanDisburse, e.CanReverse, e.CanWaivePenalty, e.CanWriteOff,
		e.CreatedBy, e.UpdatedBy, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save governance position: %w", err)
	}
	return nil
}

// Update mengupdate GovernancePosition yang sudah ada berdasarkan ID + scope.
func (r *GovernanceRepository) Update(ctx context.Context, s scope.Scope, e *governance.GovernancePosition) error {
	const q = `
		UPDATE governance_positions
		SET nasabah_id = $4, position_type = $5, position_level = $6,
		    term_start = $7, term_end = $8, term_number = $9,
		    status = $10, appointed_by = $11, appointment_doc_id = $12,
		    max_approval_amount = $13, can_disburse = $14, can_reverse = $15,
		    can_waive_penalty = $16, can_write_off = $17, updated_by = $18, updated_at = $19
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.NasabahID, e.PositionType, e.PositionLevel,
		e.TermStart, e.TermEnd, e.TermNumber,
		e.Status, e.AppointedBy, e.AppointmentDocID,
		e.MaxApprovalAmount, e.CanDisburse, e.CanReverse,
		e.CanWaivePenalty, e.CanWriteOff, e.UpdatedBy, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update governance position: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return governance.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *GovernanceRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE governance_positions SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete governance position: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return governance.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu GovernancePosition berdasarkan ID + scope.
func (r *GovernanceRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*governance.GovernancePosition, error) {
	const q = `
		SELECT id, nasabah_id, position_type, position_level,
		       term_start, term_end, term_number,
		       status, appointed_by, appointment_doc_id,
		       max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
		       created_at, updated_at, created_by, updated_by, deleted_at
		FROM governance_positions
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e governance.GovernancePosition
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.NasabahID, &e.PositionType, &e.PositionLevel,
		&e.TermStart, &e.TermEnd, &e.TermNumber,
		&e.Status, &e.AppointedBy, &e.AppointmentDocID,
		&e.MaxApprovalAmount, &e.CanDisburse, &e.CanReverse, &e.CanWaivePenalty, &e.CanWriteOff,
		&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, governance.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get governance position by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar GovernancePosition dengan pagination, sorting, dan filter.
func (r *GovernanceRepository) List(ctx context.Context, s scope.Scope, filter governance.ListFilter, limit, offset int, sortBy, order string) ([]*governance.GovernancePosition, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countPositions(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectPositions(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *GovernanceRepository) buildWhereClause(s scope.Scope, f governance.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.PositionType != nil {
		conditions = append(conditions, fmt.Sprintf("position_type = $%d", argIdx))
		args = append(args, *f.PositionType)
		argIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.PositionLevel != nil {
		conditions = append(conditions, fmt.Sprintf("position_level = $%d", argIdx))
		args = append(args, *f.PositionLevel)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countPositions menghitung total GovernancePosition aktif yang cocok dengan filter.
func (r *GovernanceRepository) countPositions(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM governance_positions WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count governance positions: %w", err)
	}
	return total, nil
}

// selectPositions mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *GovernanceRepository) selectPositions(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*governance.GovernancePosition, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":     "created_at",
		"term_start":     "term_start",
		"term_end":       "term_end",
		"position_type":  "position_type",
		"position_level": "position_level",
		"status":         "status",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, nasabah_id, position_type, position_level,
		        term_start, term_end, term_number,
		        status, appointed_by, appointment_doc_id,
		        max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
		        created_at, updated_at, created_by, updated_by, deleted_at
		 FROM governance_positions
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list governance positions: %w", err)
	}
	defer rows.Close()
	return scanGovernancePositions(rows)
}

func scanGovernancePositions(rows *sql.Rows) ([]*governance.GovernancePosition, error) {
	var results []*governance.GovernancePosition
	for rows.Next() {
		var e governance.GovernancePosition
		if err := rows.Scan(
			&e.ID, &e.NasabahID, &e.PositionType, &e.PositionLevel,
			&e.TermStart, &e.TermEnd, &e.TermNumber,
			&e.Status, &e.AppointedBy, &e.AppointmentDocID,
			&e.MaxApprovalAmount, &e.CanDisburse, &e.CanReverse, &e.CanWaivePenalty, &e.CanWriteOff,
			&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ governance.WriteRepository = (*GovernanceRepository)(nil)
	_ governance.ReadRepository  = (*GovernanceRepository)(nil)
)

// suppressUnusedImport mencegah error unused import untuk time package.
var _ = time.Time{}
