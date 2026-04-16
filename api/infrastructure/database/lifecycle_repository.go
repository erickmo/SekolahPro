package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/lifecycle"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// LifecycleRepository adalah concrete implementation dari lifecycle.WriteRepository + lifecycle.ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type LifecycleRepository struct {
	db *sqlx.DB
}

// NewLifecycleRepository membuat instance baru LifecycleRepository.
func NewLifecycleRepository(db *sqlx.DB) *LifecycleRepository {
	return &LifecycleRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity MembershipLifecycle baru ke database.
func (r *LifecycleRepository) Save(ctx context.Context, s scope.Scope, e *lifecycle.MembershipLifecycle) error {
	const q = `
		INSERT INTO membership_lifecycles (
			id, tenant_id, company_id, nasabah_id, current_status, previous_status,
			last_transition, transition_reason, applied_at, verified_at, verified_by,
			activated_at, suspended_at, suspended_by, resigned_at, revoked_at, revoked_by,
			membership_no, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.NasabahID, e.CurrentStatus, e.PreviousStatus,
		e.LastTransition, e.TransitionReason,
		e.AppliedAt, e.VerifiedAt, e.VerifiedBy,
		e.ActivatedAt, e.SuspendedAt, e.SuspendedBy,
		e.ResignedAt, e.RevokedAt, e.RevokedBy,
		e.MembershipNo,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save membership lifecycle: %w", err)
	}
	return nil
}

// Update mengupdate MembershipLifecycle yang sudah ada berdasarkan ID + scope.
func (r *LifecycleRepository) Update(ctx context.Context, s scope.Scope, e *lifecycle.MembershipLifecycle) error {
	const q = `
		UPDATE membership_lifecycles
		SET nasabah_id = $4, current_status = $5, previous_status = $6,
		    last_transition = $7, transition_reason = $8,
		    verified_at = $9, verified_by = $10,
		    activated_at = $11, suspended_at = $12, suspended_by = $13,
		    resigned_at = $14, revoked_at = $15, revoked_by = $16,
		    membership_no = $17, updated_at = $18
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.NasabahID, e.CurrentStatus, e.PreviousStatus,
		e.LastTransition, e.TransitionReason,
		e.VerifiedAt, e.VerifiedBy,
		e.ActivatedAt, e.SuspendedAt, e.SuspendedBy,
		e.ResignedAt, e.RevokedAt, e.RevokedBy,
		e.MembershipNo,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update membership lifecycle: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return lifecycle.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *LifecycleRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE membership_lifecycles SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete membership lifecycle: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return lifecycle.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu MembershipLifecycle berdasarkan ID + scope.
func (r *LifecycleRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*lifecycle.MembershipLifecycle, error) {
	const q = `
		SELECT id, nasabah_id, current_status, previous_status,
		       last_transition, transition_reason,
		       applied_at, verified_at, verified_by,
		       activated_at, suspended_at, suspended_by,
		       resigned_at, revoked_at, revoked_by,
		       membership_no, created_at, updated_at, deleted_at
		FROM membership_lifecycles
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e lifecycle.MembershipLifecycle
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.NasabahID, &e.CurrentStatus, &e.PreviousStatus,
		&e.LastTransition, &e.TransitionReason,
		&e.AppliedAt, &e.VerifiedAt, &e.VerifiedBy,
		&e.ActivatedAt, &e.SuspendedAt, &e.SuspendedBy,
		&e.ResignedAt, &e.RevokedAt, &e.RevokedBy,
		&e.MembershipNo,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, lifecycle.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get membership lifecycle by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar MembershipLifecycle dengan filter, pagination, dan sorting.
func (r *LifecycleRepository) List(ctx context.Context, s scope.Scope, filter lifecycle.LifecycleFilter, limit, offset int, sortBy, order string) ([]*lifecycle.MembershipLifecycle, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countLifecycles(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectLifecycles(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *LifecycleRepository) buildWhereClause(s scope.Scope, f lifecycle.LifecycleFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.CurrentStatus != nil {
		conditions = append(conditions, fmt.Sprintf("current_status = $%d", paramIdx))
		args = append(args, string(*f.CurrentStatus))
		paramIdx++
	}
	if f.NasabahID != nil {
		conditions = append(conditions, fmt.Sprintf("nasabah_id = $%d", paramIdx))
		args = append(args, *f.NasabahID)
		paramIdx++
	}
	if f.LastTransition != nil {
		conditions = append(conditions, fmt.Sprintf("last_transition = $%d", paramIdx))
		args = append(args, string(*f.LastTransition))
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

// countLifecycles menghitung total MembershipLifecycle yang cocok dengan filter.
func (r *LifecycleRepository) countLifecycles(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM membership_lifecycles WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count membership lifecycles: %w", err)
	}
	return total, nil
}

// selectLifecycles mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *LifecycleRepository) selectLifecycles(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*lifecycle.MembershipLifecycle, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":       "created_at",
		"current_status":   "current_status",
		"last_transition":  "last_transition",
		"nasabah_id":       "nasabah_id",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, nasabah_id, current_status, previous_status,
		        last_transition, transition_reason,
		        applied_at, verified_at, verified_by,
		        activated_at, suspended_at, suspended_by,
		        resigned_at, revoked_at, revoked_by,
		        membership_no, created_at, updated_at, deleted_at
		 FROM membership_lifecycles
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list membership lifecycles: %w", err)
	}
	defer rows.Close()
	return scanLifecycles(rows)
}

func scanLifecycles(rows *sql.Rows) ([]*lifecycle.MembershipLifecycle, error) {
	var results []*lifecycle.MembershipLifecycle
	for rows.Next() {
		var e lifecycle.MembershipLifecycle
		if err := rows.Scan(
			&e.ID, &e.NasabahID, &e.CurrentStatus, &e.PreviousStatus,
			&e.LastTransition, &e.TransitionReason,
			&e.AppliedAt, &e.VerifiedAt, &e.VerifiedBy,
			&e.ActivatedAt, &e.SuspendedAt, &e.SuspendedBy,
			&e.ResignedAt, &e.RevokedAt, &e.RevokedBy,
			&e.MembershipNo,
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
	_ lifecycle.WriteRepository = (*LifecycleRepository)(nil)
	_ lifecycle.ReadRepository  = (*LifecycleRepository)(nil)
)
