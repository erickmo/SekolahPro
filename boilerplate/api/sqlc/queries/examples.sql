-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetExampleByID :one
SELECT id, name, description, is_active, created_at, updated_at, deleted_at
FROM examples
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListExamples :many
SELECT id, name, description, is_active, created_at, updated_at
FROM examples
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountExamples :one
SELECT COUNT(*)
FROM examples
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertExample :one
INSERT INTO examples (id, tenant_id, company_id, branch_id, name, description, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: UpdateExample :exec
UPDATE examples
SET name = $4, description = $5, is_active = $6, updated_at = $7
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteExample :exec
UPDATE examples
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
