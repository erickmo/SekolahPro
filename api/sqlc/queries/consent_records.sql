-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetConsentRecordByID :one
SELECT id, tenant_id, company_id, nasabah_id,
       consent_type, purpose, legal_basis,
       consent_text, consent_version, consent_given, consent_method,
       withdrawn, withdrawn_at, withdrawal_reason,
       parent_id, parent_relationship, parent_consent_given,
       given_at, ip_address, user_agent, witness_id,
       created_at, deleted_at
FROM consent_records
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListConsentRecords :many
SELECT id, tenant_id, company_id, nasabah_id,
       consent_type, purpose, legal_basis,
       consent_text, consent_version, consent_given, consent_method,
       withdrawn, withdrawn_at, withdrawal_reason,
       parent_id, parent_relationship, parent_consent_given,
       given_at, ip_address, user_agent, witness_id,
       created_at
FROM consent_records
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: InsertConsentRecord :exec
INSERT INTO consent_records (
    id, tenant_id, company_id, nasabah_id,
    consent_type, purpose, legal_basis,
    consent_text, consent_version, consent_given, consent_method,
    withdrawn, withdrawn_at, withdrawal_reason,
    parent_id, parent_relationship, parent_consent_given,
    given_at, ip_address, user_agent, witness_id,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22);

-- name: UpdateConsentRecord :exec
UPDATE consent_records
SET consent_type = $4, purpose = $5, legal_basis = $6,
    consent_text = $7, consent_version = $8, consent_given = $9, consent_method = $10,
    withdrawn = $11, withdrawn_at = $12, withdrawal_reason = $13,
    parent_id = $14, parent_relationship = $15, parent_consent_given = $16,
    given_at = $17, ip_address = $18, user_agent = $19, witness_id = $20
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteConsentRecord :exec
UPDATE consent_records
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: CountConsentRecords :one
SELECT COUNT(*)
FROM consent_records
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;
