-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: ListBackupConfigs :many
SELECT id, config_type, schedule, retention_days, is_enabled,
       last_run_at, next_run_at, status, created_at, updated_at
FROM backup_configs
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountBackupConfigs :one
SELECT COUNT(*)
FROM backup_configs
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertBackupConfig :one
INSERT INTO backup_configs (
    id, tenant_id, company_id, config_type, schedule,
    retention_days, is_enabled, status, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id;

-- name: UpdateBackupConfig :exec
UPDATE backup_configs
SET config_type = $4, schedule = $5, retention_days = $6,
    is_enabled = $7, status = $8, updated_at = $9
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteBackupConfig :exec
UPDATE backup_configs
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
