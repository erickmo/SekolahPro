-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetDataSubjectRequestByID :one
SELECT id, tenant_id, company_id, nasabah_id,
       request_type, description, data_categories,
       status, assigned_to, rejection_reason,
       response_data, response_summary,
       received_at, deadline_at, completed_at,
       created_at, updated_at, deleted_at
FROM data_subject_requests
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListDataSubjectRequests :many
SELECT id, tenant_id, company_id, nasabah_id,
       request_type, description, data_categories,
       status, assigned_to, rejection_reason,
       response_data, response_summary,
       received_at, deadline_at, completed_at,
       created_at, updated_at
FROM data_subject_requests
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: InsertDataSubjectRequest :exec
INSERT INTO data_subject_requests (
    id, tenant_id, company_id, nasabah_id,
    request_type, description, data_categories,
    status, assigned_to, rejection_reason,
    response_data, response_summary,
    received_at, deadline_at, completed_at,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);

-- name: UpdateDataSubjectRequest :exec
UPDATE data_subject_requests
SET request_type = $4, description = $5, data_categories = $6,
    status = $7, assigned_to = $8, rejection_reason = $9,
    response_data = $10, response_summary = $11,
    received_at = $12, deadline_at = $13, completed_at = $14,
    updated_at = $15
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteDataSubjectRequest :exec
UPDATE data_subject_requests
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: CountDataSubjectRequests :one
SELECT COUNT(*)
FROM data_subject_requests
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;
