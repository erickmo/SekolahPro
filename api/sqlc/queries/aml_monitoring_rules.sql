-- Queries untuk sqlc code generation: AML Monitoring Rules (ADR-K027).
-- tenant_id dan company_id selalu menjadi filter pertama di semua query.

-- name: GetAmlRuleByID :one
SELECT id, rule_code, rule_name, rule_type::text, description, parameters,
       is_active, applies_to::text, auto_alert, auto_block, created_at, updated_at, deleted_at
FROM aml_monitoring_rules
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListAmlRules :many
SELECT id, rule_code, rule_name, rule_type::text, description, parameters,
       is_active, applies_to::text, auto_alert, auto_block, created_at, updated_at
FROM aml_monitoring_rules
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAmlRules :one
SELECT COUNT(*)
FROM aml_monitoring_rules
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertAmlRule :exec
INSERT INTO aml_monitoring_rules (
    id, tenant_id, company_id, rule_code, rule_name, rule_type, description,
    parameters, is_active, applies_to, auto_alert, auto_block, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6::aml_rule_type, $7,
          $8, $9, $10::aml_rule_applies_to, $11, $12, $13, $14);

-- name: UpdateAmlRule :exec
UPDATE aml_monitoring_rules
SET rule_name = $4, rule_type = $5::aml_rule_type, description = $6,
    parameters = $7, applies_to = $8::aml_rule_applies_to,
    auto_alert = $9, auto_block = $10, updated_at = $11
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ToggleAmlRule :exec
UPDATE aml_monitoring_rules
SET is_active = $4, updated_at = $5
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteAmlRule :exec
UPDATE aml_monitoring_rules
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
