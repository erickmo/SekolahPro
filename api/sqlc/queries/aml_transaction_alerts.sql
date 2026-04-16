-- Queries untuk sqlc code generation: AML Transaction Alerts (ADR-K027).
-- tenant_id dan company_id selalu menjadi filter pertama di semua query.

-- name: GetAmlAlertByID :one
SELECT id, nasabah_id, transaksi_id, rule_id, alert_type::text, severity::text,
       description, transaction_details, status::text, investigator_id,
       investigation_notes, outcome::text, ltkm_filed, ltkm_reference,
       created_at, updated_at, deleted_at
FROM aml_transaction_alerts
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListAmlAlerts :many
SELECT id, nasabah_id, transaksi_id, rule_id, alert_type::text, severity::text,
       description, transaction_details, status::text, investigator_id,
       investigation_notes, outcome::text, ltkm_filed, ltkm_reference,
       created_at, updated_at
FROM aml_transaction_alerts
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAmlAlerts :one
SELECT COUNT(*)
FROM aml_transaction_alerts
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertAmlAlert :exec
INSERT INTO aml_transaction_alerts (
    id, tenant_id, company_id, nasabah_id, transaksi_id, rule_id,
    alert_type, severity, description, transaction_details, status,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6,
          $7::aml_alert_type, $8::aml_severity, $9, $10,
          $11::aml_alert_status, $12, $13);

-- name: UpdateAmlAlert :exec
UPDATE aml_transaction_alerts
SET status = $4::aml_alert_status, investigator_id = $5,
    investigation_notes = $6, outcome = $7::aml_outcome,
    ltkm_filed = $8, ltkm_reference = $9, updated_at = $10
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteAmlAlert :exec
UPDATE aml_transaction_alerts
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
