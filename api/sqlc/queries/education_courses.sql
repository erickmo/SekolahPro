-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetEducationCourseByID :one
SELECT id, tenant_id, company_id, course_code, title, description, program_type,
       content_type, content_url, content_doc_id, duration_minutes,
       prerequisite_ids, mandatory, has_quiz, passing_score, certificate_template,
       target_audience, applicable_mode, is_active, created_at, updated_at, deleted_at
FROM education_courses
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListEducationCourses :many
SELECT id, tenant_id, company_id, course_code, title, description, program_type,
       content_type, content_url, content_doc_id, duration_minutes,
       prerequisite_ids, mandatory, has_quiz, passing_score, certificate_template,
       target_audience, applicable_mode, is_active, created_at, updated_at
FROM education_courses
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountEducationCourses :one
SELECT COUNT(*)
FROM education_courses
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertEducationCourse :one
INSERT INTO education_courses (
    id, tenant_id, company_id, course_code, title, description, program_type,
    content_type, content_url, content_doc_id, duration_minutes,
    prerequisite_ids, mandatory, has_quiz, passing_score, certificate_template,
    target_audience, applicable_mode, is_active, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING id;

-- name: UpdateEducationCourse :exec
UPDATE education_courses
SET course_code = $4, title = $5, description = $6, program_type = $7,
    content_type = $8, content_url = $9, content_doc_id = $10, duration_minutes = $11,
    prerequisite_ids = $12, mandatory = $13, has_quiz = $14, passing_score = $15,
    certificate_template = $16, target_audience = $17, applicable_mode = $18,
    is_active = $19, updated_at = $20
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteEducationCourse :exec
UPDATE education_courses
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
