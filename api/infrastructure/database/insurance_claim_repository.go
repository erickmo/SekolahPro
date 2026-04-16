package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/insurance"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// InsuranceClaimRepository adalah concrete implementation dari insurance.ClaimWriteRepository + ClaimReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type InsuranceClaimRepository struct {
	db *sqlx.DB
}

// NewInsuranceClaimRepository membuat instance baru InsuranceClaimRepository.
func NewInsuranceClaimRepository(db *sqlx.DB) *InsuranceClaimRepository {
	return &InsuranceClaimRepository{db: db}
}

// uuidArrayToInterfaceClaim mengkonversi []uuid.UUID ke []any untuk parameterized query.
func uuidArrayToInterfaceClaim(ids []uuid.UUID) any {
	if ids == nil {
		return []uuid.UUID{}
	}
	return ids
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity Claim baru ke database.
func (r *InsuranceClaimRepository) Save(ctx context.Context, s scope.Scope, e *insurance.Claim) error {
	const q = `
		INSERT INTO insurance_claims (
			id, tenant_id, company_id, policy_id, claim_type, status,
			claim_amount, approved_amount, incident_date,
			submitted_at, submitted_by, description, document_ids,
			reviewed_at, reviewed_by, rejection_reason,
			paid_at, paid_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.PolicyID, e.ClaimType, e.Status,
		e.ClaimAmount, e.ApprovedAmount, e.IncidentDate,
		e.SubmittedAt, e.SubmittedBy, e.Description,
		uuidArrayToInterfaceClaim(e.DocumentIDs),
		e.ReviewedAt, e.ReviewedBy, e.RejectionReason,
		e.PaidAt, e.PaidBy,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save insurance claim: %w", err)
	}
	return nil
}

// Update mengupdate Claim yang sudah ada berdasarkan ID + scope.
func (r *InsuranceClaimRepository) Update(ctx context.Context, s scope.Scope, e *insurance.Claim) error {
	const q = `
		UPDATE insurance_claims
		SET policy_id = $4, claim_type = $5, status = $6,
		    claim_amount = $7, approved_amount = $8, incident_date = $9,
		    description = $10, document_ids = $11,
		    reviewed_at = $12, reviewed_by = $13, rejection_reason = $14,
		    paid_at = $15, paid_by = $16, updated_at = $17
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.PolicyID, e.ClaimType, e.Status,
		e.ClaimAmount, e.ApprovedAmount, e.IncidentDate,
		e.Description, uuidArrayToInterfaceClaim(e.DocumentIDs),
		e.ReviewedAt, e.ReviewedBy, e.RejectionReason,
		e.PaidAt, e.PaidBy,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update insurance claim: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return insurance.ErrClaimNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *InsuranceClaimRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE insurance_claims SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete insurance claim: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return insurance.ErrClaimNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu Claim berdasarkan ID + scope.
func (r *InsuranceClaimRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*insurance.Claim, error) {
	const q = `
		SELECT id, policy_id, claim_type, status,
		       claim_amount, approved_amount, incident_date,
		       submitted_at, submitted_by, description, document_ids,
		       reviewed_at, reviewed_by, rejection_reason,
		       paid_at, paid_by,
		       created_at, updated_at, deleted_at
		FROM insurance_claims
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e insurance.Claim
	var documentIDs []uuid.UUID

	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.PolicyID, &e.ClaimType, &e.Status,
		&e.ClaimAmount, &e.ApprovedAmount, &e.IncidentDate,
		&e.SubmittedAt, &e.SubmittedBy, &e.Description, &documentIDs,
		&e.ReviewedAt, &e.ReviewedBy, &e.RejectionReason,
		&e.PaidAt, &e.PaidBy,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, insurance.ErrClaimNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get insurance claim by id: %w", err)
	}

	e.DocumentIDs = documentIDs
	return &e, nil
}

// List mengambil daftar Claim dengan filter, pagination, dan sorting.
func (r *InsuranceClaimRepository) List(ctx context.Context, s scope.Scope, filter insurance.ClaimFilter, limit, offset int, sortBy, order string) ([]*insurance.Claim, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countClaims(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectClaims(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *InsuranceClaimRepository) buildWhereClause(s scope.Scope, f insurance.ClaimFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.PolicyID != nil {
		conditions = append(conditions, fmt.Sprintf("policy_id = $%d", paramIdx))
		args = append(args, *f.PolicyID)
		paramIdx++
	}
	if f.ClaimType != nil {
		conditions = append(conditions, fmt.Sprintf("claim_type = $%d", paramIdx))
		args = append(args, string(*f.ClaimType))
		paramIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, string(*f.Status))
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

// countClaims menghitung total Claim yang cocok dengan filter.
func (r *InsuranceClaimRepository) countClaims(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM insurance_claims WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count insurance claims: %w", err)
	}
	return total, nil
}

// selectClaims mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *InsuranceClaimRepository) selectClaims(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*insurance.Claim, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":    "created_at",
		"submitted_at":  "submitted_at",
		"claim_type":    "claim_type",
		"status":        "status",
		"incident_date": "incident_date",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, policy_id, claim_type, status,
		        claim_amount, approved_amount, incident_date,
		        submitted_at, submitted_by, description, document_ids,
		        reviewed_at, reviewed_by, rejection_reason,
		        paid_at, paid_by,
		        created_at, updated_at, deleted_at
		 FROM insurance_claims
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list insurance claims: %w", err)
	}
	defer rows.Close()
	return scanInsuranceClaims(rows)
}

func scanInsuranceClaims(rows *sql.Rows) ([]*insurance.Claim, error) {
	var results []*insurance.Claim
	for rows.Next() {
		var e insurance.Claim
		var documentIDs []uuid.UUID
		if err := rows.Scan(
			&e.ID, &e.PolicyID, &e.ClaimType, &e.Status,
			&e.ClaimAmount, &e.ApprovedAmount, &e.IncidentDate,
			&e.SubmittedAt, &e.SubmittedBy, &e.Description, &documentIDs,
			&e.ReviewedAt, &e.ReviewedBy, &e.RejectionReason,
			&e.PaidAt, &e.PaidBy,
			&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
		); err != nil {
			return nil, err
		}
		e.DocumentIDs = documentIDs
		results = append(results, &e)
	}
	return results, rows.Err()
}

// Compile-time interface check.
var (
	_ insurance.ClaimWriteRepository = (*InsuranceClaimRepository)(nil)
	_ insurance.ClaimReadRepository  = (*InsuranceClaimRepository)(nil)
)
