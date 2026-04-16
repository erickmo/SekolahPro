-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetReserveFundByID :one
SELECT id, fund_type, fund_name, description, current_balance, target_balance, min_balance,
       shu_allocation_pct, max_balance_pct, auto_allocate, investment_instrument,
       investment_maturity, investment_rate, status, frozen_reason,
       created_at, updated_at, created_by, updated_by, deleted_at
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListReserveFunds :many
SELECT id, fund_type, fund_name, description, current_balance, target_balance, min_balance,
       shu_allocation_pct, max_balance_pct, auto_allocate, investment_instrument,
       investment_maturity, investment_rate, status, frozen_reason,
       created_at, updated_at, created_by, updated_by
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListReserveFundsByType :many
SELECT id, fund_type, fund_name, description, current_balance, target_balance, min_balance,
       shu_allocation_pct, max_balance_pct, auto_allocate, investment_instrument,
       investment_maturity, investment_rate, status, frozen_reason,
       created_at, updated_at, created_by, updated_by
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND fund_type = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListReserveFundsByStatus :many
SELECT id, fund_type, fund_name, description, current_balance, target_balance, min_balance,
       shu_allocation_pct, max_balance_pct, auto_allocate, investment_instrument,
       investment_maturity, investment_rate, status, frozen_reason,
       created_at, updated_at, created_by, updated_by
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListReserveFundsByTypeAndStatus :many
SELECT id, fund_type, fund_name, description, current_balance, target_balance, min_balance,
       shu_allocation_pct, max_balance_pct, auto_allocate, investment_instrument,
       investment_maturity, investment_rate, status, frozen_reason,
       created_at, updated_at, created_by, updated_by
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND fund_type = $3 AND status = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountReserveFunds :one
SELECT COUNT(*)
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CountReserveFundsByType :one
SELECT COUNT(*)
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND fund_type = $3 AND deleted_at IS NULL;

-- name: CountReserveFundsByStatus :one
SELECT COUNT(*)
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL;

-- name: CountReserveFundsByTypeAndStatus :one
SELECT COUNT(*)
FROM reserve_funds
WHERE tenant_id = $1 AND company_id = $2 AND fund_type = $3 AND status = $4 AND deleted_at IS NULL;

-- name: InsertReserveFund :exec
INSERT INTO reserve_funds (id, tenant_id, company_id, fund_type, fund_name, description,
    current_balance, target_balance, min_balance, shu_allocation_pct, max_balance_pct,
    auto_allocate, investment_instrument, investment_maturity, investment_rate,
    status, frozen_reason, created_at, updated_at, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);

-- name: UpdateReserveFund :exec
UPDATE reserve_funds
SET fund_type = $4, fund_name = $5, description = $6, current_balance = $7,
    target_balance = $8, min_balance = $9, shu_allocation_pct = $10, max_balance_pct = $11,
    auto_allocate = $12, investment_instrument = $13, investment_maturity = $14,
    investment_rate = $15, status = $16, frozen_reason = $17, updated_at = $18, updated_by = $19
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteReserveFund :exec
UPDATE reserve_funds
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
