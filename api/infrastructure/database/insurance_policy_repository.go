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

// InsurancePolicyRepository adalah concrete implementation dari insurance.PolicyWriteRepository + PolicyReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type InsurancePolicyRepository struct {
	db *sqlx.DB
}

// NewInsurancePolicyRepository membuat instance baru InsurancePolicyRepository.
func NewInsurancePolicyRepository(db *sqlx.DB) *InsurancePolicyRepository {
	return &InsurancePolicyRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity Policy baru ke database.
func (r *InsurancePolicyRepository) Save(ctx context.Context, s scope.Scope, e *insurance.Policy) error {
	const q = `
		INSERT INTO insurance_policies (
			id, tenant_id, company_id, product_id, nasabah_id, policy_no, status,
			start_date, end_date, premium_amount, coverage_amount,
			last_premium_paid, next_premium_due,
			beneficiary_name, beneficiary_relation,
			issued_at, issued_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID,
		e.ProductID, e.NasabahID, e.PolicyNo, e.Status,
		e.StartDate, e.EndDate, e.PremiumAmount, e.CoverageAmount,
		e.LastPremiumPaid, e.NextPremiumDue,
		e.BeneficiaryName, e.BeneficiaryRelation,
		e.IssuedAt, e.IssuedBy,
		e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save insurance policy: %w", err)
	}
	return nil
}

// Update mengupdate Policy yang sudah ada berdasarkan ID + scope.
func (r *InsurancePolicyRepository) Update(ctx context.Context, s scope.Scope, e *insurance.Policy) error {
	const q = `
		UPDATE insurance_policies
		SET product_id = $4, nasabah_id = $5, policy_no = $6, status = $7,
		    start_date = $8, end_date = $9, premium_amount = $10, coverage_amount = $11,
		    last_premium_paid = $12, next_premium_due = $13,
		    beneficiary_name = $14, beneficiary_relation = $15,
		    issued_at = $16, issued_by = $17, updated_at = $18
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.ProductID, e.NasabahID, e.PolicyNo, e.Status,
		e.StartDate, e.EndDate, e.PremiumAmount, e.CoverageAmount,
		e.LastPremiumPaid, e.NextPremiumDue,
		e.BeneficiaryName, e.BeneficiaryRelation,
		e.IssuedAt, e.IssuedBy,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update insurance policy: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return insurance.ErrPolicyNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *InsurancePolicyRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE insurance_policies SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete insurance policy: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return insurance.ErrPolicyNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu Policy berdasarkan ID + scope.
func (r *InsurancePolicyRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*insurance.Policy, error) {
	const q = `
		SELECT id, product_id, nasabah_id, policy_no, status,
		       start_date, end_date, premium_amount, coverage_amount,
		       last_premium_paid, next_premium_due,
		       beneficiary_name, beneficiary_relation,
		       issued_at, issued_by,
		       created_at, updated_at, deleted_at
		FROM insurance_policies
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e insurance.Policy
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.ProductID, &e.NasabahID, &e.PolicyNo, &e.Status,
		&e.StartDate, &e.EndDate, &e.PremiumAmount, &e.CoverageAmount,
		&e.LastPremiumPaid, &e.NextPremiumDue,
		&e.BeneficiaryName, &e.BeneficiaryRelation,
		&e.IssuedAt, &e.IssuedBy,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, insurance.ErrPolicyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get insurance policy by id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar Policy dengan filter, pagination, dan sorting.
func (r *InsurancePolicyRepository) List(ctx context.Context, s scope.Scope, filter insurance.PolicyFilter, limit, offset int, sortBy, order string) ([]*insurance.Policy, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countPolicies(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectPolicies(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *InsurancePolicyRepository) buildWhereClause(s scope.Scope, f insurance.PolicyFilter) (string, []any) {
	conditions := []string{"tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	paramIdx := 3

	if f.ProductID != nil {
		conditions = append(conditions, fmt.Sprintf("product_id = $%d", paramIdx))
		args = append(args, *f.ProductID)
		paramIdx++
	}
	if f.NasabahID != nil {
		conditions = append(conditions, fmt.Sprintf("nasabah_id = $%d", paramIdx))
		args = append(args, *f.NasabahID)
		paramIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramIdx))
		args = append(args, string(*f.Status))
		paramIdx++
	}

	return strings.Join(conditions, " AND "), args
}

// countPolicies menghitung total Policy yang cocok dengan filter.
func (r *InsurancePolicyRepository) countPolicies(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM insurance_policies WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count insurance policies: %w", err)
	}
	return total, nil
}

// selectPolicies mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *InsurancePolicyRepository) selectPolicies(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*insurance.Policy, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":  "created_at",
		"policy_no":   "policy_no",
		"status":      "status",
		"start_date":  "start_date",
		"end_date":    "end_date",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, product_id, nasabah_id, policy_no, status,
		        start_date, end_date, premium_amount, coverage_amount,
		        last_premium_paid, next_premium_due,
		        beneficiary_name, beneficiary_relation,
		        issued_at, issued_by,
		        created_at, updated_at, deleted_at
		 FROM insurance_policies
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list insurance policies: %w", err)
	}
	defer rows.Close()
	return scanInsurancePolicies(rows)
}

func scanInsurancePolicies(rows *sql.Rows) ([]*insurance.Policy, error) {
	var results []*insurance.Policy
	for rows.Next() {
		var e insurance.Policy
		if err := rows.Scan(
			&e.ID, &e.ProductID, &e.NasabahID, &e.PolicyNo, &e.Status,
			&e.StartDate, &e.EndDate, &e.PremiumAmount, &e.CoverageAmount,
			&e.LastPremiumPaid, &e.NextPremiumDue,
			&e.BeneficiaryName, &e.BeneficiaryRelation,
			&e.IssuedAt, &e.IssuedBy,
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
	_ insurance.PolicyWriteRepository = (*InsurancePolicyRepository)(nil)
	_ insurance.PolicyReadRepository  = (*InsurancePolicyRepository)(nil)
)
