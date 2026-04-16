-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetEducationEnrollmentByID :one
SELECT id, tenant_id, company_id, course_id, nasabah_id, status, progress_pct,
       started_at, completed_at, quiz_score, quiz_attempts, passed,
       certificate_id, certificate_issued_at, rating, feedback,
       enrolled_at, deadline_at, created_at, deleted_at
FROM education_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListEducationEnrollments :many
SELECT id, tenant_id, company_id, course_id, nasabah_id, status, progress_pct,
       started_at, completed_at, quiz_score, quiz_attempts, passed,
       certificate_id, certificate_issued_at, rating, feedback,
       enrolled_at, deadline_at, created_at
FROM education_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY enrolled_at DESC
LIMIT $3 OFFSET $4;

-- name: CountEducationEnrollments :one
SELECT COUNT(*)
FROM education_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertEducationEnrollment :one
INSERT INTO education_enrollments (
    id, tenant_id, company_id, course_id, nasabah_id,
    status, progress_pct, started_at, completed_at,
    quiz_score, quiz_attempts, passed,
    certificate_id, certificate_issued_at,
    rating, feedback, enrolled_at, deadline_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING id;

-- name: UpdateEducationEnrollment :exec
UPDATE education_enrollments
SET status = $4, progress_pct = $5, started_at = $6, completed_at = $7,
    quiz_score = $8, quiz_attempts = $9, passed = $10,
    certificate_id = $11, certificate_issued_at = $12,
    rating = $13, feedback = $14, deadline_at = $15
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: GetEducationStats :one
SELECT
    COUNT(*) AS total_enrollments,
    COUNT(*) FILTER (WHERE status = 'completed') AS completed_count,
    CASE WHEN COUNT(*) > 0
         THEN ROUND(100.0 * COUNT(*) FILTER (WHERE status = 'completed') / COUNT(*), 2)
         ELSE 0
    END AS completion_rate,
    ROUND(COALESCE(AVG(quiz_score), 0), 2) AS avg_quiz_score,
    ROUND(COALESCE(AVG(rating), 0), 2) AS avg_rating
FROM education_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CheckExistingEnrollment :one
SELECT id FROM education_enrollments
WHERE tenant_id = $1 AND company_id = $2 AND course_id = $3 AND nasabah_id = $4 AND deleted_at IS NULL;
