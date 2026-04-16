-- tenant_id dan company_id selalu menjadi filter pertama di semua query
-- untuk memastikan isolasi data antar tenant dan company.

-- name: GetKPIDefinitionByID :one
SELECT id, kpi_code, kpi_name, category, description, formula, unit, direction,
       healthy_min, healthy_max, warning_min, warning_max, critical_min, critical_max,
       calculation_frequency, data_sources, is_active, created_at, updated_at, deleted_at
FROM kpi_definitions
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: ListKPIDefinitions :many
SELECT id, kpi_code, kpi_name, category, description, formula, unit, direction,
       healthy_min, healthy_max, warning_min, warning_max, critical_min, critical_max,
       calculation_frequency, data_sources, is_active, created_at, updated_at
FROM kpi_definitions
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountKPIDefinitions :one
SELECT COUNT(*)
FROM kpi_definitions
WHERE tenant_id = $1 AND company_id = $2 AND deleted_at IS NULL;

-- name: InsertKPIDefinition :one
INSERT INTO kpi_definitions (
    id, tenant_id, company_id, kpi_code, kpi_name, category, description,
    formula, unit, direction, healthy_min, healthy_max, warning_min, warning_max,
    critical_min, critical_max, calculation_frequency, data_sources, is_active, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
RETURNING id;

-- name: UpdateKPIDefinition :exec
UPDATE kpi_definitions
SET kpi_code = $4, kpi_name = $5, category = $6, description = $7,
    formula = $8, unit = $9, direction = $10,
    healthy_min = $11, healthy_max = $12, warning_min = $13, warning_max = $14,
    critical_min = $15, critical_max = $16,
    calculation_frequency = $17, data_sources = $18, is_active = $19, updated_at = $20
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;

-- name: DeleteKPIDefinition :exec
UPDATE kpi_definitions
SET deleted_at = NOW(), updated_at = NOW()
WHERE tenant_id = $1 AND company_id = $2 AND id = $3 AND deleted_at IS NULL;
