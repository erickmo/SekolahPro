-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetInsuranceClaimByID :one
SELECT id, policy_id, claim_number, claim_type, claim_date, incident_date,
       document_ids, description,
       assessed_amount, approved_amount, assessor_id, assessment_notes,
       status, settlement_date, settlement_type,
       submitted_at, resolved_at, created_at, updated_at, deleted_at
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListInsuranceClaims :many
SELECT id, policy_id, claim_number, claim_type, claim_date, incident_date,
       document_ids, description,
       assessed_amount, approved_amount, assessor_id, assessment_notes,
       status, settlement_date, settlement_type,
       submitted_at, resolved_at, created_at, updated_at
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountInsuranceClaims :one
SELECT COUNT(*)
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertInsuranceClaim :exec
INSERT INTO insurance_claims (
    id, tenant_id, company_id, policy_id,
    claim_number, claim_type, claim_date, incident_date,
    document_ids, description,
    assessed_amount, approved_amount, assessor_id, assessment_notes,
    status, settlement_date, settlement_type,
    submitted_at, resolved_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);

-- name: UpdateInsuranceClaim :exec
UPDATE insurance_claims
SET claim_type = $4, description = $5,
    assessed_amount = $6, approved_amount = $7, assessor_id = $8, assessment_notes = $9,
    status = $10, settlement_date = $11, settlement_type = $12,
    resolved_at = $13, updated_at = $14
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListInsuranceClaimsByPolicy :many
SELECT id, policy_id, claim_number, claim_type, claim_date, incident_date,
       document_ids, description,
       assessed_amount, approved_amount, assessor_id, assessment_notes,
       status, settlement_date, settlement_type,
       submitted_at, resolved_at, created_at, updated_at
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND policy_id = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListInsuranceClaimsByStatus :many
SELECT id, policy_id, claim_number, claim_type, claim_date, incident_date,
       document_ids, description,
       assessed_amount, approved_amount, assessor_id, assessment_notes,
       status, settlement_date, settlement_type,
       submitted_at, resolved_at, created_at, updated_at
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountInsuranceClaimsByPolicy :one
SELECT COUNT(*)
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND policy_id = $3 AND deleted_at IS NULL;

-- name: CountInsuranceClaimsByStatus :one
SELECT COUNT(*)
FROM insurance_claims
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL;
