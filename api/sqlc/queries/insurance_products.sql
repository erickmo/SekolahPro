-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetInsuranceProductByID :one
SELECT id, product_code, product_name, product_type, provider_id,
       coverage_type, coverage_percentage, max_coverage_amount,
       premium_type, premium_rate, premium_paid_by,
       min_age, max_age, health_check_required, max_plafon,
       applicable_mode, is_active, created_at, updated_at, deleted_at
FROM insurance_products
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListInsuranceProducts :many
SELECT id, product_code, product_name, product_type, provider_id,
       coverage_type, coverage_percentage, max_coverage_amount,
       premium_type, premium_rate, premium_paid_by,
       min_age, max_age, health_check_required, max_plafon,
       applicable_mode, is_active, created_at, updated_at
FROM insurance_products
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountInsuranceProducts :one
SELECT COUNT(*)
FROM insurance_products
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertInsuranceProduct :exec
INSERT INTO insurance_products (
    id, tenant_id, company_id, product_code, product_name, product_type, provider_id,
    coverage_type, coverage_percentage, max_coverage_amount,
    premium_type, premium_rate, premium_paid_by,
    min_age, max_age, health_check_required, max_plafon,
    applicable_mode, is_active, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);

-- name: UpdateInsuranceProduct :exec
UPDATE insurance_products
SET product_code = $4, product_name = $5, product_type = $6, provider_id = $7,
    coverage_type = $8, coverage_percentage = $9, max_coverage_amount = $10,
    premium_type = $11, premium_rate = $12, premium_paid_by = $13,
    min_age = $14, max_age = $15, health_check_required = $16, max_plafon = $17,
    applicable_mode = $18, is_active = $19, updated_at = $20
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteInsuranceProduct :exec
UPDATE insurance_products
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
