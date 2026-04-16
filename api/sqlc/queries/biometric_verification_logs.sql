-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: InsertBiometricVerificationLog :exec
INSERT INTO biometric_verification_logs (
    id, tenant_id, company_id, nasabah_id, enrollment_id,
    verification_type, purpose, related_entity_type, related_entity_id,
    result, confidence_score, match_threshold, failure_reason,
    verified_at, ip_address, device_info, created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13,
    $14, $15, $16, $17
);

-- name: ListBiometricVerificationLogs :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListBiometricVerificationLogsByNasabah :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListBiometricVerificationLogsByResult :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND result = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListBiometricVerificationLogsByPurpose :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND purpose = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListBiometricVerificationLogsByNasabahAndResult :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND result = $4
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListBiometricVerificationLogsByNasabahAndPurpose :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND purpose = $4
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListBiometricVerificationLogsByResultAndPurpose :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND result = $3 AND purpose = $4
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListBiometricVerificationLogsByNasabahAndResultAndPurpose :many
SELECT id, nasabah_id, enrollment_id, verification_type, purpose,
       related_entity_type, related_entity_id,
       result, confidence_score, match_threshold, failure_reason,
       verified_at, ip_address, device_info, created_at
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND result = $4 AND purpose = $5
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountBiometricVerificationLogs :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2;

-- name: CountBiometricVerificationLogsByNasabah :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3;

-- name: CountBiometricVerificationLogsByResult :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND result = $3;

-- name: CountBiometricVerificationLogsByPurpose :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND purpose = $3;

-- name: CountBiometricVerificationLogsByNasabahAndResult :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND result = $4;

-- name: CountBiometricVerificationLogsByNasabahAndPurpose :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND purpose = $4;

-- name: CountBiometricVerificationLogsByResultAndPurpose :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND result = $3 AND purpose = $4;

-- name: CountBiometricVerificationLogsByNasabahAndResultAndPurpose :one
SELECT COUNT(*)
FROM biometric_verification_logs
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND result = $4 AND purpose = $5;
