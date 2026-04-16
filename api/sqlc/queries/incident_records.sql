-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetIncidentByID :one
SELECT id, incident_number, severity, incident_type, title, description,
       detected_at, acknowledged_at, mitigated_at, resolved_at, post_mortem_at,
       affected_services, affected_tenants, affected_nasabah,
       data_loss, data_loss_description,
       responder_ids, actions_taken, root_cause, remediation, post_mortem_doc_id,
       status, created_at, updated_at, deleted_at
FROM incident_records
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListIncidents :many
SELECT id, incident_number, severity, incident_type, title, description,
       detected_at, acknowledged_at, mitigated_at, resolved_at, post_mortem_at,
       affected_services, affected_tenants, affected_nasabah,
       data_loss, data_loss_description,
       responder_ids, actions_taken, root_cause, remediation, post_mortem_doc_id,
       status, created_at, updated_at
FROM incident_records
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountIncidents :one
SELECT COUNT(*)
FROM incident_records
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertIncident :one
INSERT INTO incident_records (
    id, tenant_id, company_id, incident_number, severity, incident_type,
    title, description, detected_at,
    affected_services, affected_tenants, affected_nasabah,
    data_loss, data_loss_description,
    responder_ids, actions_taken,
    status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15, $16,
    $17, $18, $19
) RETURNING id;

-- name: UpdateIncident :exec
UPDATE incident_records
SET severity = $4, incident_type = $5, title = $6, description = $7,
    mitigated_at = $8, post_mortem_at = $9,
    affected_services = $10, affected_tenants = $11, affected_nasabah = $12,
    data_loss = $13, data_loss_description = $14,
    responder_ids = $15, actions_taken = $16,
    root_cause = $17, remediation = $18, post_mortem_doc_id = $19,
    status = $20, updated_at = $21
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: AcknowledgeIncident :exec
UPDATE incident_records
SET acknowledged_at = $4, status = $5, updated_at = $6
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ResolveIncident :exec
UPDATE incident_records
SET resolved_at = $4, root_cause = $5, remediation = $6,
    post_mortem_doc_id = $7, status = $8, updated_at = $9
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteIncident :exec
UPDATE incident_records
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: GetNextIncidentNumber :one
SELECT CONCAT('INC-', EXTRACT(YEAR FROM NOW()), '-', LPAD(CAST(COALESCE(MAX(CAST(SUBSTRING(incident_number FROM 'INC-\\d{4}-(\\d+)') AS INT)), 0) + 1 AS TEXT), 5, '0'))
FROM incident_records
WHERE tenant_id = $1 AND company_id = $2;
