-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetInsurancePolicyByID :one
SELECT id, pinjaman_id, nasabah_id, product_id, policy_number,
       coverage_amount, premium_amount, premium_type,
       effective_date, expiry_date, status, cancellation_reason,
       issued_at, issued_by, created_at, updated_at, deleted_at
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListInsurancePolicies :many
SELECT id, pinjaman_id, nasabah_id, product_id, policy_number,
       coverage_amount, premium_amount, premium_type,
       effective_date, expiry_date, status, cancellation_reason,
       issued_at, issued_by, created_at, updated_at
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountInsurancePolicies :one
SELECT COUNT(*)
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertInsurancePolicy :exec
INSERT INTO insurance_policies (
    id, tenant_id, company_id, pinjaman_id, nasabah_id, product_id,
    policy_number, coverage_amount, premium_amount, premium_type,
    effective_date, expiry_date, status,
    issued_at, issued_by, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);

-- name: UpdateInsurancePolicy :exec
UPDATE insurance_policies
SET status = $4, cancellation_reason = $5, updated_at = $6
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteInsurancePolicy :exec
UPDATE insurance_policies
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListInsurancePoliciesByPinjaman :many
SELECT id, pinjaman_id, nasabah_id, product_id, policy_number,
       coverage_amount, premium_amount, premium_type,
       effective_date, expiry_date, status, cancellation_reason,
       issued_at, issued_by, created_at, updated_at
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND pinjaman_id = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListInsurancePoliciesByNasabah :many
SELECT id, pinjaman_id, nasabah_id, product_id, policy_number,
       coverage_amount, premium_amount, premium_type,
       effective_date, expiry_date, status, cancellation_reason,
       issued_at, issued_by, created_at, updated_at
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListInsurancePoliciesByStatus :many
SELECT id, pinjaman_id, nasabah_id, product_id, policy_number,
       coverage_amount, premium_amount, premium_type,
       effective_date, expiry_date, status, cancellation_reason,
       issued_at, issued_by, created_at, updated_at
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountInsurancePoliciesByPinjaman :one
SELECT COUNT(*)
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND pinjaman_id = $3 AND deleted_at IS NULL;

-- name: CountInsurancePoliciesByNasabah :one
SELECT COUNT(*)
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL;

-- name: CountInsurancePoliciesByStatus :one
SELECT COUNT(*)
FROM insurance_policies
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL;
