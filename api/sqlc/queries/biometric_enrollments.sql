-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetBiometricEnrollmentByID :one
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListBiometricEnrollments :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListBiometricEnrollmentsByNasabah :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListBiometricEnrollmentsByType :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND biometric_type = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListBiometricEnrollmentsByStatus :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListBiometricEnrollmentsByNasabahAndType :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND biometric_type = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListBiometricEnrollmentsByNasabahAndStatus :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND status = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListBiometricEnrollmentsByTypeAndStatus :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND biometric_type = $3 AND status = $4 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListBiometricEnrollmentsByNasabahAndTypeAndStatus :many
SELECT id, nasabah_id, biometric_type, device_type, device_id, storage_type,
       template_hash, template_encrypted, encryption_key_ref,
       quality_score, enrollment_attempts, liveness_verified,
       status, disabled_reason, expires_at,
       consent_id, consent_given_at,
       enrolled_at, enrolled_by, created_at
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND biometric_type = $4 AND status = $5 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountBiometricEnrollments :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByNasabah :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByType :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND biometric_type = $3 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByStatus :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND status = $3 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByNasabahAndType :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND biometric_type = $4 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByNasabahAndStatus :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND status = $4 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByTypeAndStatus :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND biometric_type = $3 AND status = $4 AND deleted_at IS NULL;

-- name: CountBiometricEnrollmentsByNasabahAndTypeAndStatus :one
SELECT COUNT(*)
FROM biometric_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND biometric_type = $4 AND status = $5 AND deleted_at IS NULL;

-- name: InsertBiometricEnrollment :exec
INSERT INTO biometric_enrollments (
    id, tenant_id, company_id, nasabah_id,
    biometric_type, device_type, device_id,
    storage_type, template_hash, template_encrypted, encryption_key_ref,
    quality_score, enrollment_attempts, liveness_verified,
    status, disabled_reason, expires_at,
    consent_id, consent_given_at,
    enrolled_at, enrolled_by, created_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14,
    $15, $16, $17,
    $18, $19,
    $20, $21, $22
);

-- name: UpdateBiometricEnrollment :exec
UPDATE biometric_enrollments
SET status = $4, disabled_reason = $5, enrollment_attempts = $6
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: SoftDeleteBiometricEnrollment :exec
UPDATE biometric_enrollments
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
