-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetReserveFundTxnByID :one
SELECT id, fund_id, transaction_type, amount, balance_after,
       shu_id, jurnal_id, reference_type, reference_id,
       requires_approval, approved_by, approved_at, rejection_reason,
       transaction_date, created_at, created_by, deleted_at
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListReserveFundTxns :many
SELECT id, fund_id, transaction_type, amount, balance_after,
       shu_id, jurnal_id, reference_type, reference_id,
       requires_approval, approved_by, approved_at, rejection_reason,
       transaction_date, created_at, created_by
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListReserveFundTxnsByFund :many
SELECT id, fund_id, transaction_type, amount, balance_after,
       shu_id, jurnal_id, reference_type, reference_id,
       requires_approval, approved_by, approved_at, rejection_reason,
       transaction_date, created_at, created_by
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND fund_id = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListReserveFundTxnsByType :many
SELECT id, fund_id, transaction_type, amount, balance_after,
       shu_id, jurnal_id, reference_type, reference_id,
       requires_approval, approved_by, approved_at, rejection_reason,
       transaction_date, created_at, created_by
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND transaction_type = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListReserveFundTxnsByFundAndType :many
SELECT id, fund_id, transaction_type, amount, balance_after,
       shu_id, jurnal_id, reference_type, reference_id,
       requires_approval, approved_by, approved_at, rejection_reason,
       transaction_date, created_at, created_by
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND fund_id = $3 AND transaction_type = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountReserveFundTxns :one
SELECT COUNT(*)
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CountReserveFundTxnsByFund :one
SELECT COUNT(*)
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND fund_id = $3 AND deleted_at IS NULL;

-- name: CountReserveFundTxnsByType :one
SELECT COUNT(*)
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND transaction_type = $3 AND deleted_at IS NULL;

-- name: CountReserveFundTxnsByFundAndType :one
SELECT COUNT(*)
FROM reserve_fund_transactions
WHERE tenant_id = $1 AND company_id = $2 AND fund_id = $3 AND transaction_type = $4 AND deleted_at IS NULL;

-- name: InsertReserveFundTxn :exec
INSERT INTO reserve_fund_transactions (id, tenant_id, company_id, fund_id,
    transaction_type, amount, balance_after,
    shu_id, jurnal_id, reference_type, reference_id,
    requires_approval, approved_by, approved_at, rejection_reason,
    transaction_date, created_at, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);

-- name: UpdateReserveFundTxnApproval :exec
UPDATE reserve_fund_transactions
SET approved_by = $4, approved_at = $5, rejection_reason = $6
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
