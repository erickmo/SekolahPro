package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourorg/boilerplate/internal/domain/aml_cdd"
	"github.com/yourorg/boilerplate/pkg/scope"
)

// AmlCddRepository adalah concrete implementation dari aml_cdd.WriteRepository + ReadRepository.
// Semua query menyertakan tenant_id + company_id agar isolasi data antar org terjamin.
type AmlCddRepository struct {
	db *sqlx.DB
}

// NewAmlCddRepository membuat instance baru AmlCddRepository.
func NewAmlCddRepository(db *sqlx.DB) *AmlCddRepository {
	return &AmlCddRepository{db: db}
}

// ── WriteRepository ───────────────────────────────────────────────────────────

// Save menyimpan entity CustomerDueDiligence baru ke database.
func (r *AmlCddRepository) Save(ctx context.Context, s scope.Scope, e *aml_cdd.CustomerDueDiligence) error {
	const q = `
		INSERT INTO aml_customer_due_diligence (
			id, tenant_id, company_id, nasabah_id,
			risk_level, risk_score, risk_category,
			cdd_level, cdd_purpose, source_of_funds, source_of_wealth,
			is_pep, pep_type, pep_position, pep_country, pep_relationship,
			pep_screening_date, pep_screening_source,
			beneficial_owner_name, beneficial_owner_id_no, beneficial_ownership_pct, bo_verified,
			next_review_date, last_reviewed_at, last_reviewed_by, review_count,
			status, restriction_reason, exit_reason,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33)`

	_, err := r.db.ExecContext(ctx, q,
		e.ID, s.TenantID, s.CompanyID, e.NasabahID,
		e.RiskLevel, e.RiskScore, e.RiskCategory,
		e.CDDLevel, e.CDDPurpose, e.SourceOfFunds, e.SourceOfWealth,
		e.IsPEP, e.PEPType, e.PEPPosition, e.PEPCountry, e.PEPRelationship,
		e.PEPScreeningDate, e.PEPScreeningSource,
		e.BeneficialOwnerName, e.BeneficialOwnerIDNo, e.BeneficialOwnershipPct, e.BOVerified,
		e.NextReviewDate, e.LastReviewedAt, e.LastReviewedBy, e.ReviewCount,
		e.Status, e.RestrictionReason, e.ExitReason,
		e.CreatedBy, e.UpdatedBy, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save cdd record: %w", err)
	}
	return nil
}

// Update mengupdate CustomerDueDiligence yang sudah ada berdasarkan ID + scope.
func (r *AmlCddRepository) Update(ctx context.Context, s scope.Scope, e *aml_cdd.CustomerDueDiligence) error {
	const q = `
		UPDATE aml_customer_due_diligence
		SET nasabah_id = $4, risk_level = $5, risk_score = $6, risk_category = $7,
		    cdd_level = $8, cdd_purpose = $9, source_of_funds = $10, source_of_wealth = $11,
		    is_pep = $12, pep_type = $13, pep_position = $14, pep_country = $15, pep_relationship = $16,
		    pep_screening_date = $17, pep_screening_source = $18,
		    beneficial_owner_name = $19, beneficial_owner_id_no = $20,
		    beneficial_ownership_pct = $21, bo_verified = $22,
		    next_review_date = $23, last_reviewed_at = $24, last_reviewed_by = $25, review_count = $26,
		    status = $27, restriction_reason = $28, exit_reason = $29,
		    updated_by = $30, updated_at = $31
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q,
		s.TenantID, s.CompanyID, e.ID,
		e.NasabahID, e.RiskLevel, e.RiskScore, e.RiskCategory,
		e.CDDLevel, e.CDDPurpose, e.SourceOfFunds, e.SourceOfWealth,
		e.IsPEP, e.PEPType, e.PEPPosition, e.PEPCountry, e.PEPRelationship,
		e.PEPScreeningDate, e.PEPScreeningSource,
		e.BeneficialOwnerName, e.BeneficialOwnerIDNo,
		e.BeneficialOwnershipPct, e.BOVerified,
		e.NextReviewDate, e.LastReviewedAt, e.LastReviewedBy, e.ReviewCount,
		e.Status, e.RestrictionReason, e.ExitReason,
		e.UpdatedBy, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update cdd record: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return aml_cdd.ErrNotFound
	}
	return nil
}

// Delete melakukan soft delete (mengisi deleted_at = now()).
func (r *AmlCddRepository) Delete(ctx context.Context, s scope.Scope, id uuid.UUID) error {
	const q = `
		UPDATE aml_customer_due_diligence SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, q, s.TenantID, s.CompanyID, id)
	if err != nil {
		return fmt.Errorf("delete cdd record: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return aml_cdd.ErrNotFound
	}
	return nil
}

// ── ReadRepository ────────────────────────────────────────────────────────────

// GetByID mengambil satu CustomerDueDiligence berdasarkan ID + scope.
func (r *AmlCddRepository) GetByID(ctx context.Context, s scope.Scope, id uuid.UUID) (*aml_cdd.CustomerDueDiligence, error) {
	const q = `
		SELECT id, nasabah_id, risk_level, risk_score, risk_category,
		       cdd_level, cdd_purpose, source_of_funds, source_of_wealth,
		       is_pep, pep_type, pep_position, pep_country, pep_relationship,
		       pep_screening_date, pep_screening_source,
		       beneficial_owner_name, beneficial_owner_id_no, beneficial_ownership_pct, bo_verified,
		       next_review_date, last_reviewed_at, last_reviewed_by, review_count,
		       status, restriction_reason, exit_reason,
		       created_at, updated_at, created_by, updated_by, deleted_at
		FROM aml_customer_due_diligence
		WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL`

	var e aml_cdd.CustomerDueDiligence
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, id).Scan(
		&e.ID, &e.NasabahID, &e.RiskLevel, &e.RiskScore, &e.RiskCategory,
		&e.CDDLevel, &e.CDDPurpose, &e.SourceOfFunds, &e.SourceOfWealth,
		&e.IsPEP, &e.PEPType, &e.PEPPosition, &e.PEPCountry, &e.PEPRelationship,
		&e.PEPScreeningDate, &e.PEPScreeningSource,
		&e.BeneficialOwnerName, &e.BeneficialOwnerIDNo, &e.BeneficialOwnershipPct, &e.BOVerified,
		&e.NextReviewDate, &e.LastReviewedAt, &e.LastReviewedBy, &e.ReviewCount,
		&e.Status, &e.RestrictionReason, &e.ExitReason,
		&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, aml_cdd.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get cdd by id: %w", err)
	}
	return &e, nil
}

// GetByNasabahID mengambil CDD record berdasarkan nasabah_id + scope.
func (r *AmlCddRepository) GetByNasabahID(ctx context.Context, s scope.Scope, nasabahID uuid.UUID) (*aml_cdd.CustomerDueDiligence, error) {
	const q = `
		SELECT id, nasabah_id, risk_level, risk_score, risk_category,
		       cdd_level, cdd_purpose, source_of_funds, source_of_wealth,
		       is_pep, pep_type, pep_position, pep_country, pep_relationship,
		       pep_screening_date, pep_screening_source,
		       beneficial_owner_name, beneficial_owner_id_no, beneficial_ownership_pct, bo_verified,
		       next_review_date, last_reviewed_at, last_reviewed_by, review_count,
		       status, restriction_reason, exit_reason,
		       created_at, updated_at, created_by, updated_by, deleted_at
		FROM aml_customer_due_diligence
		WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL`

	var e aml_cdd.CustomerDueDiligence
	err := r.db.QueryRowContext(ctx, q, s.TenantID, s.CompanyID, nasabahID).Scan(
		&e.ID, &e.NasabahID, &e.RiskLevel, &e.RiskScore, &e.RiskCategory,
		&e.CDDLevel, &e.CDDPurpose, &e.SourceOfFunds, &e.SourceOfWealth,
		&e.IsPEP, &e.PEPType, &e.PEPPosition, &e.PEPCountry, &e.PEPRelationship,
		&e.PEPScreeningDate, &e.PEPScreeningSource,
		&e.BeneficialOwnerName, &e.BeneficialOwnerIDNo, &e.BeneficialOwnershipPct, &e.BOVerified,
		&e.NextReviewDate, &e.LastReviewedAt, &e.LastReviewedBy, &e.ReviewCount,
		&e.Status, &e.RestrictionReason, &e.ExitReason,
		&e.CreatedAt, &e.UpdatedAt, &e.CreatedBy, &e.UpdatedBy, &e.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, aml_cdd.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get cdd by nasabah id: %w", err)
	}
	return &e, nil
}

// List mengambil daftar CustomerDueDiligence dengan pagination, sorting, dan filter.
func (r *AmlCddRepository) List(ctx context.Context, s scope.Scope, filter aml_cdd.ListFilter, limit, offset int, sortBy, order string) ([]*aml_cdd.CustomerDueDiligence, int, error) {
	where, args := r.buildWhereClause(s, filter)

	total, err := r.countCDDs(ctx, where, args)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.selectCDDs(ctx, where, args, limit, offset, sortBy, order)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// buildWhereClause membangun dynamic WHERE clause berdasarkan scope dan filter.
func (r *AmlCddRepository) buildWhereClause(s scope.Scope, f aml_cdd.ListFilter) (string, []any) {
	conditions := []string{"tenant_id = $1", "company_id = $2", "deleted_at IS NULL"}
	args := []any{s.TenantID, s.CompanyID}
	argIdx := 3

	if f.RiskLevel != nil {
		conditions = append(conditions, fmt.Sprintf("risk_level = $%d", argIdx))
		args = append(args, *f.RiskLevel)
		argIdx++
	}
	if f.CDDLevel != nil {
		conditions = append(conditions, fmt.Sprintf("cdd_level = $%d", argIdx))
		args = append(args, *f.CDDLevel)
		argIdx++
	}
	if f.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.IsPEP != nil {
		conditions = append(conditions, fmt.Sprintf("is_pep = $%d", argIdx))
		args = append(args, *f.IsPEP)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	return where, args
}

// countCDDs menghitung total CDD record aktif yang cocok dengan filter.
func (r *AmlCddRepository) countCDDs(ctx context.Context, where string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(*) FROM aml_customer_due_diligence WHERE %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count cdd records: %w", err)
	}
	return total, nil
}

// selectCDDs mengambil baris dengan ORDER BY + LIMIT/OFFSET.
func (r *AmlCddRepository) selectCDDs(ctx context.Context, where string, args []any, limit, offset int, sortBy, order string) ([]*aml_cdd.CustomerDueDiligence, error) {
	col := safeColumn(sortBy, map[string]string{
		"created_at":       "created_at",
		"risk_level":       "risk_level",
		"risk_score":       "risk_score",
		"cdd_level":        "cdd_level",
		"status":           "status",
		"next_review_date": "next_review_date",
	}, "created_at")
	dir := safeOrder(order)

	q := fmt.Sprintf(
		`SELECT id, nasabah_id, risk_level, risk_score, risk_category,
		        cdd_level, cdd_purpose, source_of_funds, source_of_wealth,
		        is_pep, pep_type, pep_position, pep_country, pep_relationship,
		        pep_screening_date, pep_screening_source,
		        beneficial_owner_name, beneficial_owner_id_no, beneficial_ownership_pct, bo_verified,
		        next_review_date, last_reviewed_at, last_reviewed_by, review_count,
		        status, restriction_reason, exit_reason,
		        created_at, updated_at, created_by, updated_by, deleted_at
		 FROM aml_customer_due_diligence
		 WHERE %s
		 ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, col, dir, len(args)+1, len(args)+2,
	)

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list cdd records: %w", err)
	}
	defer rows.Close()
	return scanCDDs(rows)
}

func scanCDDs(rows *sql.Rows) ([]*aml_cdd.CustomerDueDiligence, error) {
	var results []*aml_cdd.CustomerDueDiligence
	for rows.Next() {
		var e aml_cdd.CustomerDueDiligence
		if err := rows.Scan(
			&e.ID, &e.NasabahID, &e.RiskLevel, &e.RiskScore, &e.RiskCategory,
			&e.CDDLevel, &e.CDDPurpose, &e.SourceOfFunds, &e.SourceOfWealth,
			&e.IsPEP, &e.PEPType, &e.PEPPosition, &e.PEPCountry, &e.PEPRelationship,
			&e.PEPScreeningDate, &e.PEPScreeningSource,
			&e.BeneficialOwnerName, &e.BeneficialOwnerIDNo, &e.BeneficialOwnershipPct, &e.BOVerified,
			&e.NextReviewDate, &e.LastReviewedAt, &e.LastReviewedBy, &e.ReviewCount,
			&e.Status, &e.RestrictionReason, &e.ExitReason,
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
	_ aml_cdd.WriteRepository = (*AmlCddRepository)(nil)
	_ aml_cdd.ReadRepository  = (*AmlCddRepository)(nil)
)

// suppressUnusedImport mencegah error unused import.
var (
	_ = time.Time{}
	_ = json.RawMessage{}
)
