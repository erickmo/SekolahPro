-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetGovernancePositionByID :one
SELECT id, tenant_id, company_id, nasabah_id,
       position_type, position_level, term_start, term_end, term_number,
       status, appointed_by,
       max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
       created_at, updated_at, deleted_at
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListGovernancePositions :many
SELECT id, tenant_id, company_id, nasabah_id,
       position_type, position_level, term_start, term_end, term_number,
       status, appointed_by,
       max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
       created_at, updated_at
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListGovernancePositionsByType :many
SELECT id, tenant_id, company_id, nasabah_id,
       position_type, position_level, term_start, term_end, term_number,
       status, appointed_by,
       max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
       created_at, updated_at
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND position_type = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListGovernancePositionsByStatus :many
SELECT id, tenant_id, company_id, nasabah_id,
       position_type, position_level, term_start, term_end, term_number,
       status, appointed_by,
       max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
       created_at, updated_at
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListGovernancePositionsByTypeAndStatus :many
SELECT id, tenant_id, company_id, nasabah_id,
       position_type, position_level, term_start, term_end, term_number,
       status, appointed_by,
       max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
       created_at, updated_at
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND position_type = $3 AND status = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountGovernancePositions :one
SELECT COUNT(*)
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CountGovernancePositionsByType :one
SELECT COUNT(*)
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND position_type = $3 AND deleted_at IS NULL;

-- name: CountGovernancePositionsByStatus :one
SELECT COUNT(*)
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL;

-- name: CountGovernancePositionsByTypeAndStatus :one
SELECT COUNT(*)
FROM governance_positions
WHERE tenant_id = $1 AND company_id = $2 AND position_type = $3 AND status = $4 AND deleted_at IS NULL;

-- name: InsertGovernancePosition :exec
INSERT INTO governance_positions (
    id, tenant_id, company_id, nasabah_id,
    position_type, position_level, term_start, term_end, term_number,
    status, appointed_by,
    max_approval_amount, can_disburse, can_reverse, can_waive_penalty, can_write_off,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);

-- name: UpdateGovernancePosition :exec
UPDATE governance_positions
SET nasabah_id = $4,
    position_type = $5,
    position_level = $6,
    term_start = $7,
    term_end = $8,
    term_number = $9,
    status = $10,
    appointed_by = $11,
    max_approval_amount = $12,
    can_disburse = $13,
    can_reverse = $14,
    can_waive_penalty = $15,
    can_write_off = $16,
    updated_at = $17
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: UpdateGovernanceStatus :exec
UPDATE governance_positions
SET status = $4, updated_at = $5
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteGovernancePosition :exec
UPDATE governance_positions
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
