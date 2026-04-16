package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/dissolution"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// DissolutionRepository adalah concrete implementation dari dissolution.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type DissolutionRepository struct {
	db *sqlx.DB
}

// NewDissolutionRepository membuat instance baru DissolutionRepository.
func NewDissolutionRepository(db *sqlx.DB) *DissolutionRepository {
	return &DissolutionRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity DissolutionProcess baru ke database.
func (r *DissolutionRepository) Save(ctx context.Context, s scope.Scope, e *dissolution.DissolutionProcess) error {
	const q = `
		INSERT INTO dissolution_processes (
			id, tenant_id, company_id, dissolution_type, rat_meeting_id, reason,
			effective_date, liquidator_ids, supervisor_id, stage, claim_deadline,
			total_assets, total_liabilities, net_equity, distribution_per_member,
			status, final_report_doc_id, closed_at, initiated_at, initiated_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.DissolutionType, e.RatMeetingID, e.Reason,
		e.EffectiveDate, uuidArrayToInterface(e.LiquidatorIDs), e.SupervisorID,
		e.Stage, e.ClaimDeadline,
		e.TotalAssets, e.TotalLiabilities, e.NetEquity, e.DistributionPerMember,
		e.Status, e.FinalReportDocID, e.ClosedAt,
		e.InitiatedAt, e.InitiatedBy, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save dissolution: %w", err)
	}
	return nil
}

// Update mengupdate DissolutionProcess yang sudah ada berdasarkan ID + scope.
func (r *DissolutionRepository) Update(ctx context.Context, s scope.Scope, e *dissolution.DissolutionProcess) error {
	const q = `
		UPDATE dissolution_processes
		SET dissolution_type = $4, rat_meeting_id = $5, reason = $6,
		    effective_date = $7, liquidator_ids = $8, supervisor_id = $9,
		    stage = $10, claim_deadline = $11,
		    total_assets = $12, total_liabilities = $13, net_equity = $14,
		    distribution_per_member = $15, status = $16,
		    final_report_doc_id = $17, closed_at = $18
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.DissolutionType, e.RatMeetingID, e.Reason,
		e.EffectiveDate, uuidArrayToInterface(e.LiquidatorIDs), e.SupervisorID,
		e.Stage, e.ClaimDeadline,
		e.TotalAssets, e.TotalLiabilities, e.NetEquity, e.DistributionPerMember,
		e.Status, e.FinalReportDocID, e.ClosedAt,
	)
	if err != nil {
		return fmt.Errorf("update dissolution: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return dissolution.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *DissolutionRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE dissolution_processes SET deleted_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete dissolution: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return dissolution.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu DissolutionProcess berdasarkan ID + scope.
func (r *DissolutionRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*dissolution.DissolutionProcess, error) {
	const q = `
		SELECT id, dissolution_type, rat_meeting_id, reason, effective_date,
		       liquidator_ids, supervisor_id, stage, claim_deadline,
		       total_assets, total_liabilities, net_equity, distribution_per_member,
		       status, final_report_doc_id, closed_at,
		       initiated_at, initiated_by, deleted_at, created_at
		FROM dissolution_processes
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e dissolution.DissolutionProcess
	var liquidatorIDs []uuid.UUID

	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.DissolutionType, &e.RatMeetingID, &e.Reason, &e.EffectiveDate,
		&liquidatorIDs, &e.SupervisorID, &e.Stage, &e.ClaimDeadline,
		&e.TotalAssets, &e.TotalLiabilities, &e.NetEquity, &e.DistributionPerMember,
		&e.Status, &e.FinalReportDocID, &e.ClosedAt,
		&e.InitiatedAt, &e.InitiatedBy, &e.DeletedAt, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, dissolution.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get dissolution by id: %w", err)
	}

	e.LiquidatorIDs = liquidatorIDs
	return &e, nil
}

// List mengambil daftar DissolutionProcess dengan pagination, sorting, dan filter.
func (r *DissolutionRepository) List(ctx context.Context, s scope.Scope, filter dissolution.ListFilter, limit, offset int, sortBy, order string) ([]*dissolution.DissolutionProcess, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countDissolutions(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectDissolutions(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *DissolutionRepository) buildWhereClause(s scope.Scope, f dissolution.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.DissolutionType != nil {
		conditions = append(conditions, fmt.Sprintf("dissolution_type = $%d", argIdx))
		args = append(args, string(*f.DissolutionType))
		argIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*f.Status))
		argIdx++
	}
	if f.Stage != nil {
		conditions = append(conditions, fmt.Sprintf("stage = $%d", argIdx))
		args = append(args, string(*f.Stage))
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countDissolutions menghitung total DissolutionProcess aktif yang cocok dengan filter.
func (r *DissolutionRepository) countDissolutions(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM dissolution_processes WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count dissolutions: %w", err)
	}
	return total, nil
}

// selectDissolutions mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *DissolutionRepository) selectDissolutions(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*dissolution.DissolutionProcess, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":     "created_at",
		"effective_date": "effective_date",
		"stage":          "stage",
		"status":         "status",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, dissolution_type, rat_meeting_id, reason, effective_date,
		        liquidator_ids, supervisor_id, stage, claim_deadline,
		        total_assets, total_liabilities, net_equity, distribution_per_member,
		        status, final_report_doc_id, closed_at,
		        initiated_at, initiated_by, deleted_at, created_at
		 FROM dissolution_processes
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list dissolutions: %w", err)
	}
	defer rows.Close()
	return scanDissolutions(rows)
}

func scanDissolutions(rows *sql.Rows) ([]*dissolution.DissolutionProcess, error) {
	var results []*dissolution.DissolutionProcess
	for rows.Next() {
		var e dissolution.DissolutionProcess
		var liquidatorIDs []uuid.UUID
		if err := rows.Scan(
			&e.ID, &e.DissolutionType, &e.RatMeetingID, &e.Reason, &e.EffectiveDate,
			&liquidatorIDs, &e.SupervisorID, &e.Stage, &e.ClaimDeadline,
			&e.TotalAssets, &e.TotalLiabilities, &e.NetEquity, &e.DistributionPerMember,
			&e.Status, &e.FinalReportDocID, &e.ClosedAt,
			&e.InitiatedAt, &e.InitiatedBy, &e.DeletedAt, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		e.LiquidatorIDs = liquidatorIDs
		results = append(results, &e)
	}
	return results, rows.Err()
}

// uuidArrayToInterface mengkonversi []uuid.UUID ke []any untuk parameterized query.
func uuidArrayToInterface(ids []uuid.UUID) any {
	if ids == nil {
		return []uuid.UUID{}
	}
	return ids
}

// Compile-time interface check.
var (
	_ dissolution.WriteRepository = (*DissolutionRepository)(nil)
	_ dissolution.ReadRepository  = (*DissolutionRepository)(nil)
)

// suppressUnusedImport mencegah error unused import untuk time package.
var _ = time.Time{}
