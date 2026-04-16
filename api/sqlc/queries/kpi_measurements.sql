-- KPI Measurement queries.

-- name: InsertKPIMeasurement :one
INSERT INTO kpi_measurements (
    id, tenant_id, company_id, kpi_definition_id, measurement_date, measurement_period,
    value, previous_value, change_pct, health_status, trend, alert_triggered, calculated_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id;

-- name: GetLatestMeasurementByKPI :one
SELECT id, kpi_definition_id, measurement_date, measurement_period,
       value, previous_value, change_pct, health_status, trend,
       alert_triggered, calculated_at, created_at
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2 AND kpi_definition_id = $3
ORDER BY measurement_date DESC
LIMIT 1;

-- name: ListKPIMeasurementsByKPI :many
SELECT id, kpi_definition_id, measurement_date, measurement_period,
       value, previous_value, change_pct, health_status, trend,
       alert_triggered, calculated_at, created_at
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2 AND kpi_definition_id = $3
ORDER BY measurement_date DESC
LIMIT $4 OFFSET $5;

-- name: CountKPIMeasurementsByKPI :one
SELECT COUNT(*)
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2 AND kpi_definition_id = $3;

-- name: ListLatestMeasurementsAll :many
SELECT DISTINCT ON (kpi_definition_id)
    id, kpi_definition_id, measurement_date, measurement_period,
    value, previous_value, change_pct, health_status, trend,
    alert_triggered, calculated_at, created_at
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2
ORDER BY kpi_definition_id, measurement_date DESC;

-- name: ListMeasurementsByStatus :many
SELECT id, kpi_definition_id, measurement_date, measurement_period,
       value, previous_value, change_pct, health_status, trend,
       alert_triggered, calculated_at, created_at
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2 AND health_status = $3
ORDER BY measurement_date DESC
LIMIT $4 OFFSET $5;

-- name: CountMeasurementsByStatus :one
SELECT COUNT(*)
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2 AND health_status = $3;

-- name: GetPreviousMeasurement :one
SELECT value
FROM kpi_measurements
WHERE tenant_id = $1 AND company_id = $2 AND kpi_definition_id = $3
ORDER BY measurement_date DESC
LIMIT 1;
