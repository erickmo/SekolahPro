-- Queries untuk sqlc code generation: Customer Due Diligence (ADR-K027).
-- tenant_id dan company_id selalu menjadi filter pertama di semua query.

-- name: GetAmlCddByID :one
SELECT id, nasabah_id, risk_level::text, risk_score, cdd_level::text,
       is_pep, pep_type::text, pep_position, purpose, source_of_funds,
       source_of_wealth, next_review_date, status::text, created_at, updated_at, deleted_at
FROM aml_customer_due_diligences
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListAmlCdds :many
SELECT id, nasabah_id, risk_level::text, risk_score, cdd_level::text,
       is_pep, pep_type::text, pep_position, purpose, source_of_funds,
       source_of_wealth, next_review_date, status::text, created_at, updated_at
FROM aml_customer_due_diligences
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAmlCdds :one
SELECT COUNT(*)
FROM aml_customer_due_diligences
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertAmlCdd :exec
INSERT INTO aml_customer_due_diligences (
    id, tenant_id, company_id, nasabah_id, risk_level, risk_score, cdd_level,
    is_pep, pep_type, pep_position, purpose, source_of_funds, source_of_wealth,
    next_review_date, status, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5::aml_risk_level, $6, $7::aml_cdd_level,
          $8, $9::aml_pep_type, $10, $11, $12, $13, $14, $15::aml_cdd_status, $16, $17);

-- name: UpdateAmlCdd :exec
UPDATE aml_customer_due_diligences
SET risk_level = $4::aml_risk_level, risk_score = $5, cdd_level = $6::aml_cdd_level,
    is_pep = $7, pep_type = $8::aml_pep_type, pep_position = $9,
    purpose = $10, source_of_funds = $11, source_of_wealth = $12,
    next_review_date = $13, status = $14::aml_cdd_status, updated_at = $15
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: UpdateAmlCddRisk :exec
UPDATE aml_customer_due_diligences
SET risk_level = $4::aml_risk_level, risk_score = $5, cdd_level = $6::aml_cdd_level,
    next_review_date = $7, status = $8::aml_cdd_status, updated_at = $9
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteAmlCdd :exec
UPDATE aml_customer_due_diligences
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
