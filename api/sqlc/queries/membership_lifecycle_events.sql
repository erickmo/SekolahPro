-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetLifecycleEventByID :one
SELECT id, tenant_id, company_id, nasabah_id,
       event_type, from_status, to_status,
       reason, initiated_by, approved_by, supporting_doc_ids,
       simpanan_pokok_refund, simpanan_wajib_refund, refund_amount, refund_status,
       outstanding_loans, loan_settlement_plan,
       event_date, effective_date, created_at, created_by, deleted_at
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListLifecycleEvents :many
SELECT id, tenant_id, company_id, nasabah_id,
       event_type, from_status, to_status,
       reason, initiated_by, approved_by, supporting_doc_ids,
       simpanan_pokok_refund, simpanan_wajib_refund, refund_amount, refund_status,
       outstanding_loans, loan_settlement_plan,
       event_date, effective_date, created_at, created_by
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY event_date DESC
LIMIT $3 OFFSET $4;

-- name: ListLifecycleEventsByNasabah :many
SELECT id, tenant_id, company_id, nasabah_id,
       event_type, from_status, to_status,
       reason, initiated_by, approved_by, supporting_doc_ids,
       simpanan_pokok_refund, simpanan_wajib_refund, refund_amount, refund_status,
       outstanding_loans, loan_settlement_plan,
       event_date, effective_date, created_at, created_by
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL
ORDER BY event_date DESC;

-- name: ListLifecycleEventsByType :many
SELECT id, tenant_id, company_id, nasabah_id,
       event_type, from_status, to_status,
       reason, initiated_by, approved_by, supporting_doc_ids,
       simpanan_pokok_refund, simpanan_wajib_refund, refund_amount, refund_status,
       outstanding_loans, loan_settlement_plan,
       event_date, effective_date, created_at, created_by
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND event_type = $3 AND deleted_at IS NULL
ORDER BY event_date DESC
LIMIT $4 OFFSET $5;

-- name: ListLifecycleEventsByNasabahAndType :many
SELECT id, tenant_id, company_id, nasabah_id,
       event_type, from_status, to_status,
       reason, initiated_by, approved_by, supporting_doc_ids,
       simpanan_pokok_refund, simpanan_wajib_refund, refund_amount, refund_status,
       outstanding_loans, loan_settlement_plan,
       event_date, effective_date, created_at, created_by
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND event_type = $4 AND deleted_at IS NULL
ORDER BY event_date DESC
LIMIT $5 OFFSET $6;

-- name: CountLifecycleEvents :one
SELECT COUNT(*)
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: CountLifecycleEventsByNasabah :one
SELECT COUNT(*)
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND deleted_at IS NULL;

-- name: CountLifecycleEventsByType :one
SELECT COUNT(*)
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND event_type = $3 AND deleted_at IS NULL;

-- name: CountLifecycleEventsByNasabahAndType :one
SELECT COUNT(*)
FROM membership_lifecycle_events
WHERE tenant_id = $1 AND company_id = $2 AND nasabah_id = $3 AND event_type = $4 AND deleted_at IS NULL;

-- name: InsertLifecycleEvent :exec
INSERT INTO membership_lifecycle_events (
    id, tenant_id, company_id, nasabah_id,
    event_type, from_status, to_status,
    reason, initiated_by, approved_by, supporting_doc_ids,
    simpanan_pokok_refund, simpanan_wajib_refund, refund_amount, refund_status,
    outstanding_loans, loan_settlement_plan,
    event_date, effective_date, created_at, created_by
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14, $15,
    $16, $17,
    $18, $19, $20, $21
);

-- name: SoftDeleteLifecycleEvent :exec
UPDATE membership_lifecycle_events
SET deleted_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
